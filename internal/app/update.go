package app

import (
	"github.com/subhammahanty235/ccsmanager/internal/session"
	"github.com/subhammahanty235/ccsmanager/internal/ui"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionsLoadedMsg struct{ list *session.SessionList }
type errMsg struct{ err error }
type resumeFinishedMsg struct{}

func scanSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		list, err := session.Scan()
		if err != nil {
			return errMsg{err}
		}
		return sessionsLoadedMsg{list}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = m.width/3 - 4
		return m, nil

	case spinner.TickMsg:
		if m.mode == ModeLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case sessionsLoadedMsg:
		m.mode = ModeNormal
		m.sessionList = msg.list
		m.sessions = msg.list.Sessions
		m.projects = session.GetUniqueProjects(msg.list.Sessions)
		if len(m.sessions) > 0 {
			m.selectSession(0)
		}
		return m, tea.ClearScreen

	case errMsg:
		m.mode = ModeNormal
		m.statusMessage = "Error: " + msg.err.Error()
		return m, nil

	case resumeFinishedMsg:
		return m, tea.Batch(tea.EnterAltScreen, tea.ClearScreen, scanSessionsCmd())

	case tea.KeyMsg:
		if m.mode == ModeLoading {
			if key.Matches(msg, m.Keys.Quit) {
				return m, tea.Quit
			}
			return m, nil
		}
		return m.handleKey(msg)
	}

	if m.mode == ModeSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeHelp:
		m.mode = ModeNormal
		return m, nil

	case ModeConfirmDelete:
		return m.handleConfirmDelete(msg)

	case ModeSearch:
		return m.handleSearch(msg)

	case ModeProjectFilter:
		return m.handleProjectFilter(msg)
	}

	switch {
	case key.Matches(msg, m.Keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.Keys.Help):
		m.mode = ModeHelp
		return m, nil

	case key.Matches(msg, m.Keys.Search):
		m.mode = ModeSearch
		m.searchInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, m.Keys.Filter):
		m.mode = ModeProjectFilter
		m.projectCursor = 0
		return m, nil

	case key.Matches(msg, m.Keys.Escape):
		if m.searchQuery != "" || m.projectFilter != "" {
			m.searchQuery = ""
			m.projectFilter = ""
			m.searchInput.SetValue("")
			m.applyFilters()
		}
		return m, nil

	case key.Matches(msg, m.Keys.NextPanel):
		m.focusedPanel = (m.focusedPanel + 1) % 3
		return m, nil

	case key.Matches(msg, m.Keys.PrevPanel):
		m.focusedPanel = (m.focusedPanel + 2) % 3
		return m, nil

	case key.Matches(msg, m.Keys.Resume):
		if m.selectedSession != nil {
			cmd := session.ResumeCommand(m.selectedSession)
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return resumeFinishedMsg{}
			})
		}
		return m, nil

	case key.Matches(msg, m.Keys.Delete):
		if m.selectedSession != nil && m.focusedPanel == ui.PanelSessions {
			m.mode = ModeConfirmDelete
		}
		return m, nil
	}

	switch m.focusedPanel {
	case ui.PanelSessions:
		return m.handleSessionNav(msg)
	case ui.PanelChat:
		return m.handleChatNav(msg)
	case ui.PanelFiles:
		return m.handleFilesNav(msg)
	}

	return m, nil
}

func (m Model) handleConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Confirm):
		if m.selectedSession != nil {
			session.Delete(m.selectedSession)
			m.mode = ModeNormal
			return m, scanSessionsCmd()
		}
	case key.Matches(msg, m.Keys.Cancel), key.Matches(msg, m.Keys.Escape):
		m.mode = ModeNormal
	}
	return m, nil
}

func (m Model) handleSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Escape):
		m.mode = ModeNormal
		m.searchInput.SetValue("")
		m.searchQuery = ""
		m.applyFilters()

	case key.Matches(msg, m.Keys.Select):
		m.mode = ModeNormal
		m.searchQuery = m.searchInput.Value()
		m.applyFilters()
		m.searchInput.Blur()

	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.searchQuery = m.searchInput.Value()
		m.applyFilters()
		return m, cmd
	}
	return m, nil
}

func (m Model) handleProjectFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Escape):
		m.mode = ModeNormal

	case key.Matches(msg, m.Keys.Up):
		if m.projectCursor > 0 {
			m.projectCursor--
		}

	case key.Matches(msg, m.Keys.Down):
		if m.projectCursor < len(m.projects) {
			m.projectCursor++
		}

	case key.Matches(msg, m.Keys.Select):
		if m.projectCursor == 0 {
			m.projectFilter = ""
		} else if m.projectCursor <= len(m.projects) {
			m.projectFilter = m.projects[m.projectCursor-1]
		}
		m.applyFilters()
		m.mode = ModeNormal
	}
	return m, nil
}

func (m Model) handleSessionNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visibleHeight := m.SessionListHeight()
	prevSession := m.selectedSession
	needsClear := false

	switch {
	case key.Matches(msg, m.Keys.Up):
		if m.sessionCursor > 0 {
			m.sessionCursor--
			if m.sessionCursor < m.sessionOffset {
				m.sessionOffset = m.sessionCursor
			}
			if m.sessionCursor >= 0 && m.sessionCursor < len(m.sessions) {
				s := m.sessions[m.sessionCursor]
				if s != m.selectedSession {
					m.selectedSession = s
					session.LoadMessages(s)
					m.fileTree = nil
					m.fileLines = nil
					m.toolUsage = nil
					if s.MessagesLoaded {
						changes := session.ExtractFileChanges(s.Messages)
						m.fileTree = ui.BuildFileTree(changes)
						m.fileLines = ui.RenderFileTree(m.fileTree, 100)
						m.toolUsage = session.CountToolUsage(s.Messages)
					}
					m.chatScroll = 0
					m.filesScroll = 0
					needsClear = true
				}
			}
		}

	case key.Matches(msg, m.Keys.Down):
		if m.sessionCursor < len(m.sessions)-1 {
			m.sessionCursor++
			if m.sessionCursor >= m.sessionOffset+visibleHeight {
				m.sessionOffset = m.sessionCursor - visibleHeight + 1
			}
			if m.sessionCursor >= 0 && m.sessionCursor < len(m.sessions) {
				s := m.sessions[m.sessionCursor]
				if s != m.selectedSession {
					m.selectedSession = s
					session.LoadMessages(s)
					m.fileTree = nil
					m.fileLines = nil
					m.toolUsage = nil
					if s.MessagesLoaded {
						changes := session.ExtractFileChanges(s.Messages)
						m.fileTree = ui.BuildFileTree(changes)
						m.fileLines = ui.RenderFileTree(m.fileTree, 100)
						m.toolUsage = session.CountToolUsage(s.Messages)
					}
					m.chatScroll = 0
					m.filesScroll = 0
					needsClear = true
				}
			}
		}

	case key.Matches(msg, m.Keys.Top):
		m.sessionCursor = 0
		m.sessionOffset = 0
		if len(m.sessions) > 0 {
			s := m.sessions[0]
			if s != m.selectedSession {
				m.selectedSession = s
				session.LoadMessages(s)
				m.fileTree = nil
				m.fileLines = nil
				m.toolUsage = nil
				if s.MessagesLoaded {
					changes := session.ExtractFileChanges(s.Messages)
					m.fileTree = ui.BuildFileTree(changes)
					m.fileLines = ui.RenderFileTree(m.fileTree, 100)
					m.toolUsage = session.CountToolUsage(s.Messages)
				}
				m.chatScroll = 0
				m.filesScroll = 0
				needsClear = true
			}
		}

	case key.Matches(msg, m.Keys.Bottom):
		if len(m.sessions) > 0 {
			m.sessionCursor = len(m.sessions) - 1
			if m.sessionCursor >= visibleHeight {
				m.sessionOffset = m.sessionCursor - visibleHeight + 1
			}
			s := m.sessions[m.sessionCursor]
			if s != m.selectedSession {
				m.selectedSession = s
				session.LoadMessages(s)
				m.fileTree = nil
				m.fileLines = nil
				m.toolUsage = nil
				if s.MessagesLoaded {
					changes := session.ExtractFileChanges(s.Messages)
					m.fileTree = ui.BuildFileTree(changes)
					m.fileLines = ui.RenderFileTree(m.fileTree, 100)
					m.toolUsage = session.CountToolUsage(s.Messages)
				}
				m.chatScroll = 0
				m.filesScroll = 0
				needsClear = true
			}
		}

	case key.Matches(msg, m.Keys.Select):
		if m.sessionCursor < len(m.sessions) {
			s := m.sessions[m.sessionCursor]
			if s != m.selectedSession {
				m.selectedSession = s
				session.LoadMessages(s)
				m.fileTree = nil
				m.fileLines = nil
				m.toolUsage = nil
				if s.MessagesLoaded {
					changes := session.ExtractFileChanges(s.Messages)
					m.fileTree = ui.BuildFileTree(changes)
					m.fileLines = ui.RenderFileTree(m.fileTree, 100)
					m.toolUsage = session.CountToolUsage(s.Messages)
				}
				m.chatScroll = 0
				m.filesScroll = 0
				needsClear = true
			}
			m.focusedPanel = ui.PanelChat
		}

	case key.Matches(msg, m.Keys.PageDown):
		m.sessionCursor += visibleHeight
		if m.sessionCursor >= len(m.sessions) {
			m.sessionCursor = len(m.sessions) - 1
		}
		if m.sessionCursor >= m.sessionOffset+visibleHeight {
			m.sessionOffset = m.sessionCursor - visibleHeight + 1
		}
		if m.sessionCursor >= 0 && m.sessionCursor < len(m.sessions) {
			s := m.sessions[m.sessionCursor]
			if s != m.selectedSession {
				m.selectedSession = s
				session.LoadMessages(s)
				m.fileTree = nil
				m.fileLines = nil
				m.toolUsage = nil
				if s.MessagesLoaded {
					changes := session.ExtractFileChanges(s.Messages)
					m.fileTree = ui.BuildFileTree(changes)
					m.fileLines = ui.RenderFileTree(m.fileTree, 100)
					m.toolUsage = session.CountToolUsage(s.Messages)
				}
				m.chatScroll = 0
				m.filesScroll = 0
				needsClear = true
			}
		}

	case key.Matches(msg, m.Keys.PageUp):
		m.sessionCursor -= visibleHeight
		if m.sessionCursor < 0 {
			m.sessionCursor = 0
		}
		if m.sessionCursor < m.sessionOffset {
			m.sessionOffset = m.sessionCursor
		}
		if m.sessionCursor >= 0 && m.sessionCursor < len(m.sessions) {
			s := m.sessions[m.sessionCursor]
			if s != m.selectedSession {
				m.selectedSession = s
				session.LoadMessages(s)
				m.fileTree = nil
				m.fileLines = nil
				m.toolUsage = nil
				if s.MessagesLoaded {
					changes := session.ExtractFileChanges(s.Messages)
					m.fileTree = ui.BuildFileTree(changes)
					m.fileLines = ui.RenderFileTree(m.fileTree, 100)
					m.toolUsage = session.CountToolUsage(s.Messages)
				}
				m.chatScroll = 0
				m.filesScroll = 0
				needsClear = true
			}
		}
	}

	_ = prevSession // suppress unused warning
	if needsClear {
		return m, tea.ClearScreen
	}
	return m, nil
}

func (m Model) handleChatNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedSession == nil || !m.selectedSession.MessagesLoaded {
		return m, nil
	}

	maxScroll := m.MaxChatScroll()
	visibleHeight := m.ChatViewHeight()

	switch {
	case key.Matches(msg, m.Keys.Up):
		if m.chatScroll > 0 {
			m.chatScroll--
		}

	case key.Matches(msg, m.Keys.Down):
		if m.chatScroll < maxScroll {
			m.chatScroll++
		}

	case key.Matches(msg, m.Keys.HalfUp):
		m.chatScroll -= visibleHeight / 2
		if m.chatScroll < 0 {
			m.chatScroll = 0
		}

	case key.Matches(msg, m.Keys.HalfDown):
		m.chatScroll += visibleHeight / 2
		if m.chatScroll > maxScroll {
			m.chatScroll = maxScroll
		}

	case key.Matches(msg, m.Keys.Top):
		m.chatScroll = 0

	case key.Matches(msg, m.Keys.Bottom):
		m.chatScroll = maxScroll

	case key.Matches(msg, m.Keys.PageUp):
		m.chatScroll -= visibleHeight
		if m.chatScroll < 0 {
			m.chatScroll = 0
		}

	case key.Matches(msg, m.Keys.PageDown):
		m.chatScroll += visibleHeight
		if m.chatScroll > maxScroll {
			m.chatScroll = maxScroll
		}
	}

	return m, nil
}

func (m Model) handleFilesNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.fileLines) == 0 {
		return m, nil
	}

	maxScroll := len(m.fileLines) - m.FilesViewHeight()
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch {
	case key.Matches(msg, m.Keys.Up):
		if m.filesScroll > 0 {
			m.filesScroll--
		}

	case key.Matches(msg, m.Keys.Down):
		if m.filesScroll < maxScroll {
			m.filesScroll++
		}

	case key.Matches(msg, m.Keys.Top):
		m.filesScroll = 0

	case key.Matches(msg, m.Keys.Bottom):
		m.filesScroll = maxScroll
	}

	return m, nil
}

func (m *Model) selectSession(index int) {
	if index < 0 || index >= len(m.sessions) {
		return
	}

	s := m.sessions[index]
	if s != m.selectedSession {
		m.selectedSession = s
		session.LoadMessages(s)
		m.rebuildFileTree()
		m.chatScroll = 0
	}
}

func (m *Model) rebuildFileTree() {
	if m.selectedSession == nil || !m.selectedSession.MessagesLoaded {
		m.fileTree = nil
		m.fileLines = nil
		m.toolUsage = nil
		return
	}

	changes := session.ExtractFileChanges(m.selectedSession.Messages)
	m.fileTree = ui.BuildFileTree(changes)
	m.fileLines = ui.RenderFileTree(m.fileTree, 100)
	m.toolUsage = session.CountToolUsage(m.selectedSession.Messages)
	m.filesScroll = 0
}

func (m *Model) applyFilters() {
	if m.sessionList == nil {
		return
	}

	m.sessions = m.sessionList.Sessions

	if m.projectFilter != "" {
		m.sessions = session.FilterByProject(m.sessions, m.projectFilter)
	}

	if m.searchQuery != "" {
		m.sessions = session.Search(m.sessions, m.searchQuery)
	}

	if m.sessionCursor >= len(m.sessions) {
		m.sessionCursor = len(m.sessions) - 1
	}
	if m.sessionCursor < 0 {
		m.sessionCursor = 0
	}
	m.sessionOffset = 0

	if len(m.sessions) > 0 {
		m.selectSession(m.sessionCursor)
	} else {
		m.selectedSession = nil
		m.fileTree = nil
		m.fileLines = nil
		m.toolUsage = nil
	}
}
