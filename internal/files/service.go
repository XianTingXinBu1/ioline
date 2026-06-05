package files

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxTextFileSize int64 = 1 << 20 // 1 MiB

var (
	ErrPathRequired           = errors.New("path is required")
	ErrInvalidPath            = errors.New("invalid path")
	ErrWorkspaceNotConfigured = errors.New("workspace is not configured")
	ErrNotRegularFile         = errors.New("path is not a regular file")
	ErrBinaryFile             = errors.New("binary or executable file cannot be opened as text")
	ErrFileTooLarge           = errors.New("file is too large to open as text")
	ErrAlreadyExists          = errors.New("path already exists")
	ErrDirectoryNotEmpty      = errors.New("directory is not empty")
	ErrSourceDestinationEqual = errors.New("source and destination paths are identical")
)

// Entry describes a file system item under the workspace.
type Entry struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Type       string    `json:"type"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
	Readonly   bool      `json:"readonly"`
	Hidden     bool      `json:"hidden"`
}

// FileContent represents a text file payload for the editor.
type FileContent struct {
	Path       string    `json:"path"`
	Content    string    `json:"content"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
	Readonly   bool      `json:"readonly"`
	Binary     bool      `json:"binary"`
	LineEnding string    `json:"lineEnding"`
}

// SaveRequest defines the payload for saving a text file.
type SaveRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// SaveResult describes the result of a save operation.
type SaveResult struct {
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

// CreateFileRequest defines the payload for creating a new file.
type CreateFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
}

// CreateDirectoryRequest defines the payload for creating a directory.
type CreateDirectoryRequest struct {
	Path string `json:"path"`
}

// MoveRequest defines a rename or move operation inside the workspace.
type MoveRequest struct {
	FromPath string `json:"fromPath"`
	ToPath   string `json:"toPath"`
}

// DeleteRequest defines a delete operation.
type DeleteRequest struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

// OperationResult describes the outcome of a filesystem mutation.
type OperationResult struct {
	Path       string    `json:"path"`
	Type       string    `json:"type"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

// MoveResult describes a move or rename outcome.
type MoveResult struct {
	FromPath   string    `json:"fromPath"`
	ToPath     string    `json:"toPath"`
	Type       string    `json:"type"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

// Service provides workspace-scoped file operations.
type Service struct{}

// NewService creates a file service.
func NewService() *Service {
	return &Service{}
}

// List returns direct children of a workspace-relative directory.
func (s *Service) List(rootPath, relPath string) ([]Entry, error) {
	absPath, cleanRel, err := resolvePath(rootPath, relPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	items := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		itemPath := entry.Name()
		if cleanRel != "." {
			itemPath = filepath.ToSlash(filepath.Join(cleanRel, entry.Name()))
		}

		items = append(items, buildEntry(itemPath, entry.Name(), info))
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == "directory"
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items, nil
}

// Stat returns metadata for one workspace-relative path.
func (s *Service) Stat(rootPath, relPath string) (Entry, error) {
	absPath, cleanRel, err := resolvePath(rootPath, relPath)
	if err != nil {
		return Entry{}, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return Entry{}, err
	}

	name := info.Name()
	if cleanRel == "." {
		name = filepath.Base(rootPath)
	}

	return buildEntry(cleanRel, name, info), nil
}

// ReadText returns text content for a regular non-binary file.
func (s *Service) ReadText(rootPath, relPath string) (FileContent, error) {
	absPath, cleanRel, err := resolvePath(rootPath, relPath)
	if err != nil {
		return FileContent{}, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return FileContent{}, err
	}
	if !info.Mode().IsRegular() {
		return FileContent{}, ErrNotRegularFile
	}
	if info.Size() > maxTextFileSize {
		return FileContent{}, ErrFileTooLarge
	}
	if info.Mode()&0o111 != 0 {
		return FileContent{}, ErrBinaryFile
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return FileContent{}, err
	}
	if isBinary(content) {
		return FileContent{}, ErrBinaryFile
	}

	return FileContent{
		Path:       cleanRel,
		Content:    string(content),
		Size:       info.Size(),
		ModifiedAt: info.ModTime(),
		Readonly:   info.Mode().Perm()&0o200 == 0,
		Binary:     false,
		LineEnding: detectLineEnding(content),
	}, nil
}

// SaveText writes a text file under the workspace.
func (s *Service) SaveText(rootPath string, req SaveRequest) (SaveResult, error) {
	absPath, cleanRel, err := resolvePath(rootPath, req.Path)
	if err != nil {
		return SaveResult{}, err
	}
	if cleanRel == "." {
		return SaveResult{}, ErrInvalidPath
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return SaveResult{}, err
	}
	if err := os.WriteFile(absPath, []byte(req.Content), 0o644); err != nil {
		return SaveResult{}, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return SaveResult{}, err
	}

	return SaveResult{
		Path:       cleanRel,
		Size:       info.Size(),
		ModifiedAt: info.ModTime(),
	}, nil
}

// CreateFile creates a new file and fails if the path already exists.
func (s *Service) CreateFile(rootPath string, req CreateFileRequest) (OperationResult, error) {
	absPath, cleanRel, err := resolvePath(rootPath, req.Path)
	if err != nil {
		return OperationResult{}, err
	}
	if cleanRel == "." {
		return OperationResult{}, ErrInvalidPath
	}
	if err := ensureNotExists(absPath); err != nil {
		return OperationResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return OperationResult{}, err
	}
	if err := os.WriteFile(absPath, []byte(req.Content), 0o644); err != nil {
		return OperationResult{}, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return OperationResult{}, err
	}

	return OperationResult{
		Path:       cleanRel,
		Type:       "file",
		ModifiedAt: info.ModTime(),
	}, nil
}

// CreateDirectory creates a new directory and required parents.
func (s *Service) CreateDirectory(rootPath string, req CreateDirectoryRequest) (OperationResult, error) {
	absPath, cleanRel, err := resolvePath(rootPath, req.Path)
	if err != nil {
		return OperationResult{}, err
	}
	if cleanRel == "." {
		return OperationResult{}, ErrInvalidPath
	}
	if err := ensureNotExists(absPath); err != nil {
		return OperationResult{}, err
	}
	if err := os.MkdirAll(absPath, 0o755); err != nil {
		return OperationResult{}, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return OperationResult{}, err
	}

	return OperationResult{
		Path:       cleanRel,
		Type:       "directory",
		ModifiedAt: info.ModTime(),
	}, nil
}

// Move renames or moves a file system item inside the workspace.
func (s *Service) Move(rootPath string, req MoveRequest) (MoveResult, error) {
	fromAbs, fromRel, err := resolvePath(rootPath, req.FromPath)
	if err != nil {
		return MoveResult{}, err
	}
	toAbs, toRel, err := resolvePath(rootPath, req.ToPath)
	if err != nil {
		return MoveResult{}, err
	}
	if fromRel == "." || toRel == "." {
		return MoveResult{}, ErrInvalidPath
	}
	if fromAbs == toAbs {
		return MoveResult{}, ErrSourceDestinationEqual
	}

	info, err := os.Stat(fromAbs)
	if err != nil {
		return MoveResult{}, err
	}
	if err := ensureNotExists(toAbs); err != nil {
		return MoveResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(toAbs), 0o755); err != nil {
		return MoveResult{}, err
	}
	if err := os.Rename(fromAbs, toAbs); err != nil {
		return MoveResult{}, err
	}

	movedInfo, err := os.Stat(toAbs)
	if err != nil {
		return MoveResult{}, err
	}

	itemType := "file"
	if info.IsDir() {
		itemType = "directory"
	}

	return MoveResult{
		FromPath:   fromRel,
		ToPath:     toRel,
		Type:       itemType,
		ModifiedAt: movedInfo.ModTime(),
	}, nil
}

// Delete removes a file or directory under the workspace.
func (s *Service) Delete(rootPath string, req DeleteRequest) (OperationResult, error) {
	absPath, cleanRel, err := resolvePath(rootPath, req.Path)
	if err != nil {
		return OperationResult{}, err
	}
	if cleanRel == "." {
		return OperationResult{}, ErrInvalidPath
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return OperationResult{}, err
	}

	itemType := "file"
	if info.IsDir() {
		itemType = "directory"
		if req.Recursive {
			if err := os.RemoveAll(absPath); err != nil {
				return OperationResult{}, err
			}
		} else {
			if err := os.Remove(absPath); err != nil {
				if isDirectoryNotEmpty(err) {
					return OperationResult{}, ErrDirectoryNotEmpty
				}
				return OperationResult{}, err
			}
		}
	} else {
		if err := os.Remove(absPath); err != nil {
			return OperationResult{}, err
		}
	}

	return OperationResult{
		Path:       cleanRel,
		Type:       itemType,
		ModifiedAt: time.Now(),
	}, nil
}

func resolvePath(rootPath, relPath string) (string, string, error) {
	if rootPath == "" {
		return "", "", ErrWorkspaceNotConfigured
	}

	cleanRel := strings.TrimSpace(relPath)
	if cleanRel == "" {
		cleanRel = "."
	}
	if filepath.IsAbs(cleanRel) {
		return "", "", ErrInvalidPath
	}

	joined := filepath.Join(rootPath, cleanRel)
	cleanAbs := filepath.Clean(joined)
	relToRoot, err := filepath.Rel(rootPath, cleanAbs)
	if err != nil {
		return "", "", ErrInvalidPath
	}
	if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
		return "", "", ErrInvalidPath
	}

	return cleanAbs, filepath.ToSlash(relToRoot), nil
}

func ensureNotExists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return ErrAlreadyExists
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func isDirectoryNotEmpty(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "directory not empty")
}

func buildEntry(path, name string, info fs.FileInfo) Entry {
	entryType := "file"
	if info.IsDir() {
		entryType = "directory"
	}

	return Entry{
		Name:       name,
		Path:       filepath.ToSlash(path),
		Type:       entryType,
		Size:       info.Size(),
		ModifiedAt: info.ModTime(),
		Readonly:   info.Mode().Perm()&0o200 == 0,
		Hidden:     strings.HasPrefix(name, "."),
	}
}

func isBinary(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	return bytes.IndexByte(content, 0) >= 0
}

func detectLineEnding(content []byte) string {
	if bytes.Contains(content, []byte("\r\n")) {
		return "crlf"
	}
	return "lf"
}
