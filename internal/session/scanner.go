package session

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ccmanager/internal/claude"
)

func Scan() (*SessionList, error) {
	projectsDir := claude.GetProjectsDir()

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
		projectPath := claude.DecodeDirectoryName(encodedDir)
		projectName := claude.GetProjectName(encodedDir)
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
			session := quickScan(filePath, sessionID, projectPath, projectName, encodedDir)
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

func quickScan(filePath, sessionID, projectPath, projectName, encodedDir string) *Session {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	s := &Session{
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

	s.LastActivity = lastTimestamp
	s.MessageCount = lineCount
	s.FirstMessage = firstHumanMessage

	return s
}

func LoadMessages(s *Session) error {
	if s.MessagesLoaded {
		return nil
	}

	file, err := os.Open(s.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	s.Messages = make([]*Message, 0)

	for scanner.Scan() {
		line := scanner.Bytes()
		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)

		msgs := ParseJSONLLine(lineCopy)
		for _, msg := range msgs {
			if msg != nil {
				s.Messages = append(s.Messages, msg)
			}
		}
	}

	s.MessagesLoaded = true
	return scanner.Err()
}

func Search(sessions []*Session, query string) []*Session {
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
