/*
Package claude provides utilities for interacting with Claude Code's
local file structure, including path encoding/decoding and CLI detection.
*/
package claude

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetClaudeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

func GetProjectsDir() string {
	return filepath.Join(GetClaudeDir(), "projects")
}

func CLIAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// DecodeDirectoryName converts an encoded directory name back to a path.
// Claude Code encodes paths by replacing / with - (e.g., "-Users-you-project").
// This function attempts to decode it back, checking which hyphens are
// actual path separators vs literal hyphens in directory names.
func DecodeDirectoryName(encoded string) string {
	if encoded == "" {
		return ""
	}

	decoded := strings.ReplaceAll(encoded, "-", "/")
	if !strings.HasPrefix(decoded, "/") {
		decoded = "/" + decoded
	}

	if _, err := os.Stat(decoded); err == nil {
		return decoded
	}

	parts := strings.Split(encoded, "-")
	if len(parts) < 2 {
		return decoded
	}

	startIdx := 0
	if parts[0] == "" {
		startIdx = 1
	}

	var result strings.Builder
	for i := startIdx; i < len(parts); i++ {
		if result.Len() == 0 {
			result.WriteString("/")
			result.WriteString(parts[i])
		} else {
			testPathHyphen := result.String() + "-" + parts[i]

			if _, err := os.Stat(testPathHyphen); err == nil {
				result.WriteString("-")
				result.WriteString(parts[i])
			} else if pathExists(result.String() + "/" + parts[i]) {
				result.WriteString("/")
				result.WriteString(parts[i])
			} else {
				result.WriteString("/")
				result.WriteString(parts[i])
			}
		}
	}

	return result.String()
}

func GetProjectName(encoded string) string {
	decoded := DecodeDirectoryName(encoded)
	parts := strings.Split(decoded, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return encoded
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
