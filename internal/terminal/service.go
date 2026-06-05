package terminal

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
)

const maxSessions = 4

var (
	ErrSessionLimitReached = errors.New("terminal session limit reached")
	ErrSessionNotFound     = errors.New("terminal session not found")
	ErrInvalidSize         = errors.New("invalid terminal size")
)

// SessionInfo describes a terminal session.
type SessionInfo struct {
	ID        string    `json:"id"`
	CWD       string    `json:"cwd"`
	Shell     string    `json:"shell"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateRequest defines terminal creation input.
type CreateRequest struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// ResizeRequest defines a terminal resize input.
type ResizeRequest struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// Session represents a live terminal session.
type Session struct {
	info SessionInfo
	cmd  *exec.Cmd
	pty  *os.File
}

// Service manages active terminal sessions.
type Service struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	nextID   int
}

// NewService creates a terminal service.
func NewService() *Service {
	return &Service{sessions: make(map[string]*Session)}
}

// List returns active terminal sessions.
func (s *Service) List() []SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]SessionInfo, 0, len(s.sessions))
	for _, session := range s.sessions {
		items = append(items, session.info)
	}
	return items
}

// Create starts a new PTY-backed shell session.
func (s *Service) Create(workspaceRoot string, req CreateRequest) (SessionInfo, error) {
	if req.Cols <= 0 || req.Rows <= 0 {
		return SessionInfo{}, ErrInvalidSize
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.sessions) >= maxSessions {
		return SessionInfo{}, ErrSessionLimitReached
	}

	s.nextID++
	id := time.Now().Format("20060102150405") + "-" + filepath.Base(workspaceRoot) + "-" + string(rune('a'+(s.nextID%26)))
	if _, exists := s.sessions[id]; exists {
		id = time.Now().Format("20060102150405.000000000")
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	cmd := exec.Command(shell)
	cmd.Dir = workspaceRoot
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}

	ptyFile, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(req.Cols), Rows: uint16(req.Rows)})
	if err != nil {
		return SessionInfo{}, err
	}

	info := SessionInfo{
		ID:        id,
		CWD:       workspaceRoot,
		Shell:     shell,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	s.sessions[id] = &Session{info: info, cmd: cmd, pty: ptyFile}
	go s.watch(id, cmd)

	return info, nil
}

// Resize updates the PTY size for one session.
func (s *Service) Resize(id string, req ResizeRequest) error {
	if req.Cols <= 0 || req.Rows <= 0 {
		return ErrInvalidSize
	}

	s.mu.RLock()
	session, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return ErrSessionNotFound
	}

	return pty.Setsize(session.pty, &pty.Winsize{Cols: uint16(req.Cols), Rows: uint16(req.Rows)})
}

// Close terminates one terminal session.
func (s *Service) Close(id string) error {
	s.mu.Lock()
	session, ok := s.sessions[id]
	if !ok {
		s.mu.Unlock()
		return ErrSessionNotFound
	}
	delete(s.sessions, id)
	session.info.Status = "closed"
	s.mu.Unlock()

	if session.cmd.Process != nil {
		_ = session.cmd.Process.Kill()
	}
	_ = session.pty.Close()
	_, _ = session.cmd.Process.Wait()
	return nil
}

// Get returns one active session.
func (s *Service) Get(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *Service) watch(id string, cmd *exec.Cmd) {
	_ = cmd.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return
	}
	session.info.Status = "exited"
}
