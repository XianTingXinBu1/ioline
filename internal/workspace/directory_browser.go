package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DirectoryItem represents one child directory for workspace selection.
type DirectoryItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// DirectoryBrowseResult describes a browsable absolute directory and its child directories.
type DirectoryBrowseResult struct {
	CurrentPath string          `json:"currentPath"`
	ParentPath  string          `json:"parentPath"`
	Items       []DirectoryItem `json:"items"`
}

// BrowseDirectories returns the current absolute directory, its parent and child directories only.
func (s *Service) BrowseDirectories(path string) (DirectoryBrowseResult, error) {
	currentPath, err := s.resolveBrowseStartPath(path)
	if err != nil {
		return DirectoryBrowseResult{}, err
	}

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return DirectoryBrowseResult{}, err
	}

	items := make([]DirectoryItem, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		childPath := filepath.Join(currentPath, entry.Name())
		items = append(items, DirectoryItem{
			Name: entry.Name(),
			Path: filepath.Clean(childPath),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	parentPath := filepath.Dir(currentPath)
	if parentPath == currentPath {
		parentPath = ""
	}

	return DirectoryBrowseResult{
		CurrentPath: currentPath,
		ParentPath:  parentPath,
		Items:       items,
	}, nil
}

func (s *Service) resolveBrowseStartPath(path string) (string, error) {
	if strings.TrimSpace(path) != "" {
		return validateAbsoluteDirectory(path)
	}

	current := s.Current()
	if current.IsSet {
		return validateAbsoluteDirectory(current.RootPath)
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		if dir, dirErr := validateAbsoluteDirectory(filepath.Join(homeDir, "project")); dirErr == nil {
			return dir, nil
		}
		if dir, dirErr := validateAbsoluteDirectory(homeDir); dirErr == nil {
			return dir, nil
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return validateAbsoluteDirectory(cwd)
}

func validateAbsoluteDirectory(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", err
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", os.ErrInvalid
	}

	return absPath, nil
}
