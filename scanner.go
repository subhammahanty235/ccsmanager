/*
Session discovery and management for Claude Code session files.
Scans ~/.claude/projects for JSONL session files and extracts metadata.
*/
package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

func ScanSessions() (*SessionList, error) {
	projectsDir := GetProjectsDir()

	result := &SessionList{
		Sessions: make([]*Session, 0),
	}

	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return result, nil
	}

	projects := make(map[string]bool)

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return result, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		encodedDir := entry.Name()
		projectPath := DecodeDirectoryName(encodedDir)
		projectName := GetProjectName(encodedDir)
		projects[encodedDir] = true

		projectDir := filepath.Join(projectsDir, encodedDir)
		files, err := os.ReadDir(projectDir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".jsonl") {
				continue
			}

			if strings.HasPrefix(file.Name(), "agent-") {
				continue
			}

			filePath := filepath.Join(projectDir, file.Name())
			info, err := file.Info()
			if err != nil {
				continue
			}

			result.TotalSize += info.Size()

			sessionID := strings.TrimSuffix(file.Name(), ".jsonl")
			session := quickScanSession(filePath, sessionID, projectPath, projectName, encodedDir)
			if session != nil {
				result.Sessions = append(result.Sessions, session)
			}
		}
	}

	result.ProjectCount = len(projects)

	sort.Slice(result.Sessions, func(i, j int) bool {
		return result.Sessions[i].LastActivity.After(result.Sessions[j].LastActivity)
	})

	return result, nil
}

func quickScanSession(filePath, sessionID, projectPath, projectName, encodedDir string) *Session {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	session := &Session{
		ID:          sessionID,
		ProjectPath: projectPath,
		ProjectName: projectName,
		EncodedDir:  encodedDir,
		FilePath:    filePath,
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineCount := 0
	var lastTimestamp time.Time
	var firstHumanMessage string

	for scanner.Scan() {
		lineCount++
		line := scanner.Bytes()

		msgs := ParseJSONLLine(line)
		for _, msg := range msgs {
			if msg == nil {
				continue
			}

			if !msg.Timestamp.IsZero() && msg.Timestamp.After(lastTimestamp) {
				lastTimestamp = msg.Timestamp
			}

			if firstHumanMessage == "" && msg.Type == MessageTypeHuman && msg.Content != "" {
				firstHumanMessage = msg.Content
				if len(firstHumanMessage) > 100 {
					firstHumanMessage = firstHumanMessage[:97] + "..."
				}
			}
		}
	}

	if lastTimestamp.IsZero() {
		if info, err := os.Stat(filePath); err == nil {
			lastTimestamp = info.ModTime()
		}
	}

	session.LastActivity = lastTimestamp
	session.MessageCount = lineCount
	session.FirstMessage = firstHumanMessage

	return session
}

func LoadSessionMessages(session *Session) error {
	if session.MessagesLoaded {
		return nil
	}

	file, err := os.Open(session.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	session.Messages = make([]*Message, 0)

	for scanner.Scan() {
		line := scanner.Bytes()
		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)

		msgs := ParseJSONLLine(lineCopy)
		for _, msg := range msgs {
			if msg != nil {
				session.Messages = append(session.Messages, msg)
			}
		}
	}

	session.MessagesLoaded = true
	return scanner.Err()
}

func SearchSessions(sessions []*Session, query string) []*Session {
	if query == "" {
		return sessions
	}

	query = strings.ToLower(query)
	var results []*Session

	for _, s := range sessions {
		if strings.Contains(strings.ToLower(s.ProjectName), query) ||
			strings.Contains(strings.ToLower(s.ProjectPath), query) ||
			strings.Contains(strings.ToLower(s.FirstMessage), query) ||
			strings.Contains(strings.ToLower(s.ID), query) {
			results = append(results, s)
		}
	}

	return results
}

func FilterByProject(sessions []*Session, projectName string) []*Session {
	if projectName == "" {
		return sessions
	}

	var results []*Session
	for _, s := range sessions {
		if s.ProjectName == projectName || s.EncodedDir == projectName {
			results = append(results, s)
		}
	}
	return results
}

func GetUniqueProjects(sessions []*Session) []string {
	seen := make(map[string]bool)
	var projects []string

	for _, s := range sessions {
		if !seen[s.ProjectName] {
			seen[s.ProjectName] = true
			projects = append(projects, s.ProjectName)
		}
	}

	sort.Strings(projects)
	return projects
}

// --- Path Decoding ---

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
			} else if dirExists(result.String() + "/" + parts[i]) {
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

func dirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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

// --- Formatting ---

func FormatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return formatInt(mins) + "m ago"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return formatInt(hours) + "h ago"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return formatInt(days) + "d ago"
	default:
		return t.Format("Jan 2")
	}
}

func GetSessionAge(t time.Time) int {
	if t.IsZero() {
		return 2
	}

	diff := time.Since(t)
	switch {
	case diff < 24*time.Hour:
		return 0
	case diff < 7*24*time.Hour:
		return 1
	default:
		return 2
	}
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return formatInt64(bytes) + " B"
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	val := float64(bytes) / float64(div)
	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	return formatFloat(val) + " " + units[exp]
}

func formatInt(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}

func formatFloat(f float64) string {
	intPart := int(f)
	decPart := int((f - float64(intPart)) * 10)
	if decPart == 0 {
		return formatInt(intPart)
	}
	return formatInt(intPart) + "." + string(rune('0'+decPart))
}

// --- History ---

type HistoryEntry struct {
	Prompt      string `json:"prompt"`
	Timestamp   string `json:"timestamp"`
	ProjectPath string `json:"projectPath"`
	SessionID   string `json:"sessionId"`
}

func LoadHistory() ([]HistoryEntry, error) {
	historyPath := filepath.Join(GetClaudeDir(), "history.jsonl")

	file, err := os.Open(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var entry HistoryEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries, scanner.Err()
}
