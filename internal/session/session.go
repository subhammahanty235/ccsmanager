/*
Package session provides types and functions for managing Claude Code
session data, including scanning, loading, filtering, and resuming sessions.
*/
package session

import (
	"os"
	"os/exec"
	"time"
)

type Session struct {
	ID             string
	ProjectPath    string
	ProjectName    string
	EncodedDir     string
	FilePath       string
	LastActivity   time.Time
	MessageCount   int
	FirstMessage   string
	Messages       []*Message
	MessagesLoaded bool
}

type SessionList struct {
	Sessions     []*Session
	TotalSize    int64
	ProjectCount int
}

func ResumeCommand(s *Session) *exec.Cmd {
	if s.ProjectPath != "" {
		if _, err := os.Stat(s.ProjectPath); err == nil {
			os.Chdir(s.ProjectPath)
		}
	}

	cmd := exec.Command("claude", "--resume", s.ID)
	cmd.Dir = s.ProjectPath
	return cmd
}

func Delete(s *Session) error {
	if s.FilePath == "" {
		return nil
	}
	return os.Remove(s.FilePath)
}
