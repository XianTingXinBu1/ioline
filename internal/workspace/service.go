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
