/*
Session resume and CLI integration.
Provides functions to resume sessions via the Claude CLI and manage session files.
*/
package main

import (
	"os"
	"os/exec"
)

func ResumeCommand(session *Session) *exec.Cmd {
	if session.ProjectPath != "" {
		if _, err := os.Stat(session.ProjectPath); err == nil {
			os.Chdir(session.ProjectPath)
		}
	}

	cmd := exec.Command("claude", "--resume", session.ID)
	cmd.Dir = session.ProjectPath
	return cmd
}

func ResumeSession(session *Session) error {
	cmd := ResumeCommand(session)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func DeleteSession(session *Session) error {
	if session.FilePath == "" {
		return nil
	}
	return os.Remove(session.FilePath)
}

func CheckClaudeCLI() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}
