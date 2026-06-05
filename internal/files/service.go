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
