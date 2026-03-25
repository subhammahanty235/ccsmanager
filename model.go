/*
Application state and update logic for the Bubbletea TUI framework.
Manages panel focus, session selection, search/filter state, and keyboard input.
*/
package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Panel identifies which panel has focus.
type Panel int

const (
	PanelSessions Panel = iota
	PanelChat
	PanelFiles
)

// Mode represents the current interaction mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeProjectFilter
	ModeConfirmDelete
	ModeHelp
)

// Model holds all application state.
type Model struct {
	width  int
	height int

	sessionList     *SessionList
	sessions        []*Session
	selectedSession *Session
	projects        []string

	focusedPanel  Panel
	sessionCursor int
	sessionOffset int
	chatScroll    int
	filesScroll   int
	projectCursor int

	mode          Mode
	searchInput   textinput.Model
	searchQuery   string
	projectFilter string

	styles Styles
	keys   KeyMap

	fileTree  *FileNode
	fileLines []string
	toolUsage map[string]int

	statusMessage string
}

// NewModel creates an initialized Model ready for use.
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Search sessions..."
	ti.CharLimit = 100
	ti.Width = 40

	return Model{
		styles:       DefaultStyles(),
		keys:         DefaultKeyMap(),
		searchInput:  ti,
		sessions:     make([]*Session, 0),
		projects:     make([]string, 0),
		focusedPanel: PanelSessions,
		mode:         ModeNormal,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(scanSessionsCmd(), tea.EnterAltScreen)
}

// --- Messages ---

type sessionsLoadedMsg struct{ list *SessionList }
type sessionSelectedMsg struct{ session *Session }
type errMsg struct{ err error }
type resumeFinishedMsg struct{}

func scanSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		list, err := ScanSessions()
		if err != nil {
			return errMsg{err}
		}
		return sessionsLoadedMsg{list}
	}
}

// --- Update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = m.width/3 - 4
		return m, nil

	case sessionsLoadedMsg:
		m.sessionList = msg.list
		m.sessions = msg.list.Sessions
		m.projects = GetUniqueProjects(msg.list.Sessions)
		if len(m.sessions) > 0 {
			m.selectSession(0)
		}
		return m, nil

	case sessionSelectedMsg:
		m.selectedSession = msg.session
		m.rebuildFileTree()
		m.chatScroll = 0
		return m, nil

	case errMsg:
		m.statusMessage = "Error: " + msg.err.Error()
		return m, nil

	case resumeFinishedMsg:
		return m, tea.Batch(tea.EnterAltScreen, tea.ClearScreen, scanSessionsCmd())

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.mode == ModeSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// --- Key Handling ---

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Mode-specific handlers
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

	// Normal mode
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Help):
		m.mode = ModeHelp
		return m, nil

	case key.Matches(msg, m.keys.Search):
		m.mode = ModeSearch
		m.searchInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Filter):
		m.mode = ModeProjectFilter
		m.projectCursor = 0
		return m, nil

	case key.Matches(msg, m.keys.Escape):
		if m.searchQuery != "" || m.projectFilter != "" {
			m.searchQuery = ""
			m.projectFilter = ""
			m.searchInput.SetValue("")
			m.applyFilters()
		}
		return m, nil

	case key.Matches(msg, m.keys.NextPanel):
		m.focusedPanel = (m.focusedPanel + 1) % 3
		return m, nil

	case key.Matches(msg, m.keys.PrevPanel):
		m.focusedPanel = (m.focusedPanel + 2) % 3
		return m, nil

	case key.Matches(msg, m.keys.Resume):
		if m.selectedSession != nil {
			cmd := ResumeCommand(m.selectedSession)
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return resumeFinishedMsg{}
			})
		}
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		if m.selectedSession != nil && m.focusedPanel == PanelSessions {
			m.mode = ModeConfirmDelete
		}
		return m, nil
	}

	// Panel-specific navigation
	switch m.focusedPanel {
	case PanelSessions:
		return m.handleSessionNav(msg)
	case PanelChat:
		return m.handleChatNav(msg)
	case PanelFiles:
		return m.handleFilesNav(msg)
	}

	return m, nil
}

func (m Model) handleConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		if m.selectedSession != nil {
			DeleteSession(m.selectedSession)
			m.mode = ModeNormal
			return m, scanSessionsCmd()
		}
	case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
	}
	return m, nil
}

func (m Model) handleSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
		m.searchInput.SetValue("")
		m.searchQuery = ""
		m.applyFilters()

	case key.Matches(msg, m.keys.Select):
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
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal

	case key.Matches(msg, m.keys.Up):
		if m.projectCursor > 0 {
			m.projectCursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.projectCursor < len(m.projects) {
			m.projectCursor++
		}

	case key.Matches(msg, m.keys.Select):
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

// --- Panel Navigation ---

func (m Model) handleSessionNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visibleHeight := m.sessionListHeight()

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.sessionCursor > 0 {
			m.sessionCursor--
			if m.sessionCursor < m.sessionOffset {
				m.sessionOffset = m.sessionCursor
			}
			m.selectSession(m.sessionCursor)
		}

	case key.Matches(msg, m.keys.Down):
		if m.sessionCursor < len(m.sessions)-1 {
			m.sessionCursor++
			if m.sessionCursor >= m.sessionOffset+visibleHeight {
				m.sessionOffset = m.sessionCursor - visibleHeight + 1
			}
			m.selectSession(m.sessionCursor)
		}

	case key.Matches(msg, m.keys.Top):
		m.sessionCursor = 0
		m.sessionOffset = 0
		if len(m.sessions) > 0 {
			m.selectSession(0)
		}

	case key.Matches(msg, m.keys.Bottom):
		if len(m.sessions) > 0 {
			m.sessionCursor = len(m.sessions) - 1
			if m.sessionCursor >= visibleHeight {
				m.sessionOffset = m.sessionCursor - visibleHeight + 1
			}
			m.selectSession(m.sessionCursor)
		}

	case key.Matches(msg, m.keys.Select):
		if m.sessionCursor < len(m.sessions) {
			m.selectSession(m.sessionCursor)
			m.focusedPanel = PanelChat
		}

	case key.Matches(msg, m.keys.PageDown):
		m.sessionCursor += visibleHeight
		if m.sessionCursor >= len(m.sessions) {
			m.sessionCursor = len(m.sessions) - 1
		}
		if m.sessionCursor >= m.sessionOffset+visibleHeight {
			m.sessionOffset = m.sessionCursor - visibleHeight + 1
		}
		if m.sessionCursor >= 0 && m.sessionCursor < len(m.sessions) {
			m.selectSession(m.sessionCursor)
		}

	case key.Matches(msg, m.keys.PageUp):
		m.sessionCursor -= visibleHeight
		if m.sessionCursor < 0 {
			m.sessionCursor = 0
		}
		if m.sessionCursor < m.sessionOffset {
			m.sessionOffset = m.sessionCursor
		}
		m.selectSession(m.sessionCursor)
	}

	return m, nil
}

func (m Model) handleChatNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedSession == nil || !m.selectedSession.MessagesLoaded {
		return m, nil
	}

	maxScroll := m.maxChatScroll()
	visibleHeight := m.chatViewHeight()

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.chatScroll > 0 {
			m.chatScroll--
		}

	case key.Matches(msg, m.keys.Down):
		if m.chatScroll < maxScroll {
			m.chatScroll++
		}

	case key.Matches(msg, m.keys.HalfUp):
		m.chatScroll -= visibleHeight / 2
		if m.chatScroll < 0 {
			m.chatScroll = 0
		}

	case key.Matches(msg, m.keys.HalfDown):
		m.chatScroll += visibleHeight / 2
		if m.chatScroll > maxScroll {
			m.chatScroll = maxScroll
		}

	case key.Matches(msg, m.keys.Top):
		m.chatScroll = 0

	case key.Matches(msg, m.keys.Bottom):
		m.chatScroll = maxScroll

	case key.Matches(msg, m.keys.PageUp):
		m.chatScroll -= visibleHeight
		if m.chatScroll < 0 {
			m.chatScroll = 0
		}

	case key.Matches(msg, m.keys.PageDown):
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

	maxScroll := len(m.fileLines) - m.filesViewHeight()
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.filesScroll > 0 {
			m.filesScroll--
		}

	case key.Matches(msg, m.keys.Down):
		if m.filesScroll < maxScroll {
			m.filesScroll++
		}

	case key.Matches(msg, m.keys.Top):
		m.filesScroll = 0

	case key.Matches(msg, m.keys.Bottom):
		m.filesScroll = maxScroll
	}

	return m, nil
}

// --- State Mutations ---

func (m *Model) selectSession(index int) {
	if index < 0 || index >= len(m.sessions) {
		return
	}

	session := m.sessions[index]
	if session != m.selectedSession {
		m.selectedSession = session
		LoadSessionMessages(session)
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

	changes := ExtractFileChanges(m.selectedSession.Messages)
	m.fileTree = BuildFileTree(changes)
	m.fileLines = RenderFileTree(m.fileTree, 100)
	m.toolUsage = CountToolUsage(m.selectedSession.Messages)
	m.filesScroll = 0
}

func (m *Model) applyFilters() {
	if m.sessionList == nil {
		return
	}

	m.sessions = m.sessionList.Sessions

	if m.projectFilter != "" {
		m.sessions = FilterByProject(m.sessions, m.projectFilter)
	}

	if m.searchQuery != "" {
		m.sessions = SearchSessions(m.sessions, m.searchQuery)
	}

	// Reset cursor bounds
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

// --- Dimension Helpers ---

func (m Model) sessionListHeight() int {
	return m.height - 6
}

func (m Model) chatViewHeight() int {
	return m.height - 6
}

func (m Model) filesViewHeight() int {
	return m.height - 8
}

func (m Model) maxChatScroll() int {
	if m.selectedSession == nil || !m.selectedSession.MessagesLoaded {
		return 0
	}

	totalLines := 0
	for _, msg := range m.selectedSession.Messages {
		if msg == nil {
			continue
		}
		switch msg.Type {
		case MessageTypeHuman, MessageTypeAssistant:
			totalLines += len(msg.Content)/50 + 2
		case MessageTypeToolUse:
			totalLines++
		}
	}

	maxScroll := totalLines - m.chatViewHeight()
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}
