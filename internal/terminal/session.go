package terminal

import "os"

// PTY returns the pseudo terminal file for the session.
func (s *Session) PTY() *os.File {
	return s.pty
}
