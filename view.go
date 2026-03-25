/*
Rendering logic for the three-panel TUI layout.
Handles the sessions list, chat history, and file changes panels.
*/
package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if m.mode == ModeHelp {
		return m.renderHelp()
	}

	var b strings.Builder

	b.WriteString(m.renderTopBar())
	b.WriteString("\n")

	panelsHeight := m.height - 4
	totalWidth := m.width - 2
	leftWidth := totalWidth * 30 / 100
	rightWidth := totalWidth * 25 / 100
	centerWidth := totalWidth - leftWidth - rightWidth

	left := m.renderSessionList(leftWidth, panelsHeight)
	center := m.renderChatView(centerWidth, panelsHeight)
	right := m.renderFilesPanel(rightWidth, panelsHeight)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, center, right))
	b.WriteString("\n")
	b.WriteString(m.renderBottomBar())

	return b.String()
}

// --- Top/Bottom Bars ---

func (m Model) renderTopBar() string {
	title := m.styles.Bold.Foreground(ColorPrimary).Render("Claude Code Manager")

	var parts []string
	parts = append(parts, title)

	if m.sessionList != nil {
		stats := fmt.Sprintf(" | %d sessions | %d projects | %s",
			len(m.sessionList.Sessions),
			m.sessionList.ProjectCount,
			FormatBytes(m.sessionList.TotalSize))
		parts = append(parts, m.styles.StatusText.Render(stats))
	}

	if m.projectFilter != "" {
		parts = append(parts, m.styles.Highlight.Render(" | Project: "+m.projectFilter))
	}
	if m.searchQuery != "" {
		parts = append(parts, m.styles.Highlight.Render(" | Search: "+m.searchQuery))
	}

	return m.styles.TopBar.Width(m.width).Render(strings.Join(parts, ""))
}

func (m Model) renderBottomBar() string {
	var content string

	switch m.mode {
	case ModeSearch:
		content = m.styles.SearchPrompt.Render("/ ") + m.searchInput.View()
	case ModeProjectFilter:
		content = m.styles.Highlight.Render("Select project: up/down to navigate, Enter to select, Esc to cancel")
	case ModeConfirmDelete:
		content = m.styles.Highlight.Render("Delete this session? (y/n)")
	default:
		content = m.styles.HelpText.Render(HelpString(m.keys, m.focusedPanel, false))
	}

	if m.statusMessage != "" {
		content = m.styles.Highlight.Foreground(ColorError).Render(m.statusMessage) + "  " + content
	}

	return m.styles.BottomBar.Width(m.width).Render(content)
}

// --- Sessions Panel ---

func (m Model) renderSessionList(width, height int) string {
	focused := m.focusedPanel == PanelSessions

	title := fmt.Sprintf("Sessions (%d)", len(m.sessions))
	if m.sessionList != nil && len(m.sessions) != len(m.sessionList.Sessions) {
		title = fmt.Sprintf("Sessions (%d/%d)", len(m.sessions), len(m.sessionList.Sessions))
	}

	titleStyle := m.styles.PanelTitle
	if focused {
		titleStyle = m.styles.PanelTitleActive
	}

	var content strings.Builder
	innerHeight := height - 4

	if m.mode == ModeProjectFilter {
		content.WriteString(m.renderProjectFilter(width-4, innerHeight))
	} else if len(m.sessions) == 0 {
		content.WriteString(m.styles.Dim.Render("\n  No sessions found\n"))
		if m.searchQuery != "" || m.projectFilter != "" {
			content.WriteString(m.styles.Dim.Render("  Press Esc to clear filters\n"))
		} else {
			content.WriteString(m.styles.Dim.Render("  Sessions appear after using\n"))
			content.WriteString(m.styles.Dim.Render("  the claude CLI\n"))
		}
	} else {
		visible := m.sessions[m.sessionOffset:]
		if len(visible) > innerHeight {
			visible = visible[:innerHeight]
		}

		for i, session := range visible {
			idx := m.sessionOffset + i
			line := m.renderSessionItem(session, width-4, idx == m.sessionCursor)
			content.WriteString(line)
			content.WriteString("\n")
		}

		if len(m.sessions) > innerHeight {
			info := fmt.Sprintf(" %d-%d of %d ",
				m.sessionOffset+1,
				min(m.sessionOffset+innerHeight, len(m.sessions)),
				len(m.sessions))
			content.WriteString(m.styles.Dim.Render(info))
		}
	}

	panel := m.styles.PanelStyle(focused, width-2, height-2)
	return panel.Render(titleStyle.Render(title) + "\n" + content.String())
}

func (m Model) renderSessionItem(session *Session, width int, selected bool) string {
	style := m.styles.SessionItem
	if selected {
		style = m.styles.SessionItemSelected
	}

	age := GetSessionAge(session.LastActivity)
	dot := m.styles.SessionDotStyle(age).Render("*")

	projectName := session.ProjectName
	if len(projectName) > 15 {
		projectName = projectName[:12] + "..."
	}

	sessionID := session.ID
	if len(sessionID) > 8 {
		sessionID = sessionID[:8]
	}

	line := fmt.Sprintf("%s %s %s %s %d msgs",
		dot,
		m.styles.SessionProject.Render(projectName),
		m.styles.SessionTime.Render(FormatTimeAgo(session.LastActivity)),
		m.styles.SessionID.Render(sessionID),
		session.MessageCount)

	if lipgloss.Width(line) > width {
		line = truncateString(line, width)
	}

	return style.Width(width).Render(line)
}

func (m Model) renderProjectFilter(width, height int) string {
	var b strings.Builder
	b.WriteString("\n")

	prefix := "  "
	if m.projectCursor == 0 {
		prefix = "> "
	}

	allProjects := prefix + "All Projects"
	if m.projectFilter == "" && m.projectCursor == 0 {
		b.WriteString(m.styles.Highlight.Render(allProjects))
	} else {
		b.WriteString(m.styles.Dim.Render(allProjects))
	}
	b.WriteString("\n")

	for i, project := range m.projects {
		prefix = "  "
		if i+1 == m.projectCursor {
			prefix = "> "
		}

		var line string
		if project == m.projectFilter {
			line = m.styles.Highlight.Render(prefix + project)
		} else if i+1 == m.projectCursor {
			line = m.styles.Bold.Render(prefix + project)
		} else {
			line = prefix + project
		}
		b.WriteString(line + "\n")

		if i >= height-3 {
			b.WriteString(m.styles.Dim.Render(fmt.Sprintf("  ... and %d more", len(m.projects)-i-1)))
			break
		}
	}

	return b.String()
}

// --- Chat Panel ---

func (m Model) renderChatView(width, height int) string {
	focused := m.focusedPanel == PanelChat

	title := "Chat"
	if m.selectedSession != nil {
		title = fmt.Sprintf("Chat - %s", m.selectedSession.ProjectName)
	}

	titleStyle := m.styles.PanelTitle
	if focused {
		titleStyle = m.styles.PanelTitleActive
	}

	var content strings.Builder
	innerWidth := width - 4
	innerHeight := height - 4

	if m.selectedSession == nil {
		content.WriteString(m.styles.Dim.Render("\n  Select a session to view\n"))
	} else if !m.selectedSession.MessagesLoaded {
		content.WriteString(m.styles.Dim.Render("\n  Loading messages...\n"))
	} else {
		lines := m.renderMessages(innerWidth)

		start := m.chatScroll
		if start >= len(lines) {
			start = len(lines) - 1
		}
		if start < 0 {
			start = 0
		}

		end := start + innerHeight
		if end > len(lines) {
			end = len(lines)
		}

		content.WriteString(strings.Join(lines[start:end], "\n"))

		if len(lines) > innerHeight {
			pct := 0
			if len(lines)-innerHeight > 0 {
				pct = m.chatScroll * 100 / (len(lines) - innerHeight)
			}
			content.WriteString("\n" + m.styles.Dim.Render(fmt.Sprintf(" [%d%%]", pct)))
		}
	}

	panel := m.styles.PanelStyle(focused, width-2, height-2)
	return panel.Render(titleStyle.Render(title) + "\n" + content.String())
}

func (m Model) renderMessages(width int) []string {
	const maxLinesPerMessage = 10

	var lines []string

	if m.selectedSession == nil || !m.selectedSession.MessagesLoaded {
		return lines
	}

	for _, msg := range m.selectedSession.Messages {
		if msg == nil {
			continue
		}

		switch msg.Type {
		case MessageTypeHuman:
			lines = append(lines, m.styles.HumanPrefix.Render("You > "))
			msgLines := wrapText(msg.Content, width-2)
			if len(msgLines) > maxLinesPerMessage {
				for _, line := range msgLines[:maxLinesPerMessage-1] {
					lines = append(lines, "  "+m.styles.HumanMessage.Render(line))
				}
				lines = append(lines, "  "+m.styles.Dim.Render(fmt.Sprintf("... (%d more lines)", len(msgLines)-maxLinesPerMessage+1)))
			} else {
				for _, line := range msgLines {
					lines = append(lines, "  "+m.styles.HumanMessage.Render(line))
				}
			}
			lines = append(lines, "")

		case MessageTypeAssistant:
			lines = append(lines, m.styles.AssistantPrefix.Render("Claude > "))
			msgLines := wrapText(msg.Content, width-2)
			if len(msgLines) > maxLinesPerMessage {
				for _, line := range msgLines[:maxLinesPerMessage-1] {
					lines = append(lines, "  "+m.styles.AssistantMessage.Render(line))
				}
				lines = append(lines, "  "+m.styles.Dim.Render(fmt.Sprintf("... (%d more lines)", len(msgLines)-maxLinesPerMessage+1)))
			} else {
				for _, line := range msgLines {
					lines = append(lines, "  "+m.styles.AssistantMessage.Render(line))
				}
			}
			lines = append(lines, "")

		case MessageTypeToolUse:
			lines = append(lines, "  "+m.styles.ToolUse.Render(FormatToolUseOneLiner(msg)))

		case MessageTypeSystem:
			if msg.Content != "" {
				lines = append(lines, m.styles.SystemMessage.Render("  [system] "+truncateString(msg.Content, width-12)))
			}
		}
	}

	return lines
}

// --- Files Panel ---

func (m Model) renderFilesPanel(width, height int) string {
	focused := m.focusedPanel == PanelFiles

	fileCount := 0
	if m.fileTree != nil {
		fileCount = CountFiles(m.fileTree)
	}
	title := fmt.Sprintf("Files (%d)", fileCount)

	titleStyle := m.styles.PanelTitle
	if focused {
		titleStyle = m.styles.PanelTitleActive
	}

	var content strings.Builder
	innerHeight := height - 6

	if m.selectedSession == nil {
		content.WriteString(m.styles.Dim.Render("\n  No session selected\n"))
	} else if len(m.fileLines) == 0 {
		content.WriteString(m.styles.Dim.Render("\n  No file changes\n"))
	} else {
		start := m.filesScroll
		if start >= len(m.fileLines) {
			start = len(m.fileLines) - 1
		}
		if start < 0 {
			start = 0
		}

		end := start + innerHeight
		if end > len(m.fileLines) {
			end = len(m.fileLines)
		}

		for _, line := range m.fileLines[start:end] {
			content.WriteString(m.colorFileTreeLine(line, width-4) + "\n")
		}
	}

	if len(m.toolUsage) > 0 {
		content.WriteString("\n")
		content.WriteString(m.styles.Dim.Render(strings.Repeat("-", width-6) + "\n"))

		var toolParts []string
		for tool, count := range m.toolUsage {
			toolParts = append(toolParts, fmt.Sprintf("%s: %d", tool, count))
		}
		summary := strings.Join(toolParts, " | ")
		if len(summary) > width-4 {
			summary = summary[:width-7] + "..."
		}
		content.WriteString(m.styles.Dim.Render(summary))
	}

	panel := m.styles.PanelStyle(focused, width-2, height-2)
	return panel.Render(titleStyle.Render(title) + "\n" + content.String())
}

func (m Model) colorFileTreeLine(line string, width int) string {
	if strings.Contains(line, "[created]") {
		line = strings.Replace(line, "[created]", m.styles.FileCreated.Render("[created]"), 1)
	}
	if strings.Contains(line, "[edited") {
		start := strings.Index(line, "[edited")
		if start >= 0 {
			end := strings.Index(line[start:], "]") + start + 1
			edited := line[start:end]
			line = line[:start] + m.styles.FileEdited.Render(edited) + line[end:]
		}
	}
	if strings.Contains(line, "[deleted]") {
		line = strings.Replace(line, "[deleted]", m.styles.FileDeleted.Render("[deleted]"), 1)
	}

	for _, branch := range []string{"├──", "└──", "│", "───"} {
		if strings.Contains(line, branch) {
			line = strings.Replace(line, branch, m.styles.TreeBranch.Render(branch), 1)
		}
	}

	if len(line) > width {
		line = line[:width-3] + "..."
	}

	return line
}

// --- Help Overlay ---

func (m Model) renderHelp() string {
	helpText := FullHelpText()

	helpWidth := 65
	helpHeight := strings.Count(helpText, "\n") + 1

	x := (m.width - helpWidth) / 2
	y := (m.height - helpHeight) / 2

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	var b strings.Builder

	for i := 0; i < y; i++ {
		b.WriteString("\n")
	}

	for _, line := range strings.Split(helpText, "\n") {
		b.WriteString(strings.Repeat(" ", x))
		b.WriteString(m.styles.Bold.Foreground(ColorPrimary).Render(line))
		b.WriteString("\n")
	}

	return b.String()
}

// --- Helpers ---

func wrapText(text string, width int) []string {
	if width <= 0 {
		width = 40
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var current strings.Builder
	for _, word := range words {
		if current.Len() == 0 {
			current.WriteString(word)
		} else if current.Len()+1+len(word) <= width {
			current.WriteString(" ")
			current.WriteString(word)
		} else {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
		}
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}

func truncateString(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 3 {
		return s[:width]
	}
	return s[:width-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
