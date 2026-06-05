package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrNotConfigured = errors.New("workspace is not configured")

// Info represents current workspace state.
type Info struct {
	RootPath string    `json:"rootPath,omitempty"`
	Name     string    `json:"name,omitempty"`
	IsSet    bool      `json:"isSet"`
	SetAt    time.Time `json:"setAt,omitempty"`
}

// Candidate describes one workspace suggestion for the frontend.
type Candidate struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Source string `json:"source"`
}

// Service stores the active workspace state in memory.
type Service struct {
	mu   sync.RWMutex
	info Info
}

// NewService creates a workspace service with no default workspace.
func NewService() *Service {
	return &Service{}
}

// Current returns the current workspace info.
func (s *Service) Current() Info {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.info
}

// RootPath returns the configured workspace root path.
func (s *Service) RootPath() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.info.IsSet {
		return "", ErrNotConfigured
	}
	return s.info.RootPath, nil
}

// Set validates and sets the current workspace.
func (s *Service) Set(path string) (Info, error) {
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return Info{}, err
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return Info{}, err
	}
	if !stat.IsDir() {
		return Info{}, errors.New("workspace path must be a directory")
	}

	info := Info{
		RootPath: absPath,
		Name:     filepath.Base(absPath),
		IsSet:    true,
		SetAt:    time.Now(),
	}

	s.mu.Lock()
	s.info = info
	s.mu.Unlock()

	return info, nil
}

// Clear removes the current workspace configuration.
func (s *Service) Clear() Info {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.info = Info{}
	return s.info
}

// Candidates returns a stable, de-duplicated list of suggested workspace roots.
// It always tries to include the current working directory and the user home
// directory first so the frontend workspace picker has immediate usable options.
func (s *Service) Candidates() []Candidate {
	current := s.Current()
	seen := make(map[string]struct{})
	items := make([]Candidate, 0, 6)

	appendCandidate := func(path, source string) {
		if path == "" {
			return
		}
		absPath, err := filepath.Abs(filepath.Clean(path))
		if err != nil {
			return
		}
		if _, ok := seen[absPath]; ok {
			return
		}
		stat, err := os.Stat(absPath)
		if err != nil || !stat.IsDir() {
			return
		}

		name := filepath.Base(absPath)
		if name == string(filepath.Separator) || name == "." || name == "" {
			name = absPath
		}

		seen[absPath] = struct{}{}
		items = append(items, Candidate{
			Name:   name,
			Path:   absPath,
			Exists: true,
			Source: source,
		})
	}

	cwd, err := os.Getwd()
	if err == nil {
		appendCandidate(cwd, "default")
	}

	homeDir, _ := os.UserHomeDir()
	appendCandidate(homeDir, "default")

	if current.IsSet {
		appendCandidate(current.RootPath, "current")
	}

	appendCandidate(filepath.Join(homeDir, "project"), "suggested")
	appendCandidate(filepath.Join(homeDir, "projects"), "suggested")
	appendCandidate(filepath.Join(homeDir, "workspace"), "suggested")

	return items
}
