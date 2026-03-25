/*
Package app contains the Bubbletea application model, state management,
and update logic for the TUI.
*/
package app

import (
	"time"

	"ccmanager/internal/session"
	"ccmanager/internal/ui"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeProjectFilter
	ModeConfirmDelete
	ModeHelp
	ModeLoading
)

type Model struct {
	width  int
	height int

	sessionList     *session.SessionList
	sessions        []*session.Session
	selectedSession *session.Session
	projects        []string

	focusedPanel  ui.Panel
	sessionCursor int
	sessionOffset int
	chatScroll    int
	filesScroll   int
	projectCursor int

	mode          Mode
	searchInput   textinput.Model
	searchQuery   string
	projectFilter string

	Styles ui.Styles
	Keys   ui.KeyMap

	fileTree  *ui.FileNode
	fileLines []string
	toolUsage map[string]int

	statusMessage string

	// Loading state
	spinner   spinner.Model
	loadStart time.Time
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Search sessions..."
	ti.CharLimit = 100
	ti.Width = 40

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = ui.DefaultStyles().Bold.Foreground(ui.ColorPrimary)

	return Model{
		Styles:       ui.DefaultStyles(),
		Keys:         ui.DefaultKeyMap(),
		searchInput:  ti,
		sessions:     make([]*session.Session, 0),
		projects:     make([]string, 0),
		focusedPanel: ui.PanelSessions,
		mode:         ModeLoading,
		spinner:      sp,
		loadStart:    time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		scanSessionsCmd(),
		tea.EnterAltScreen,
		m.spinner.Tick,
	)
}

// Expose state for view rendering
func (m Model) Width() int                        { return m.width }
func (m Model) Height() int                       { return m.height }
func (m Model) SessionList() *session.SessionList { return m.sessionList }
func (m Model) Sessions() []*session.Session      { return m.sessions }
func (m Model) SelectedSession() *session.Session { return m.selectedSession }
func (m Model) Projects() []string                { return m.projects }
func (m Model) FocusedPanel() ui.Panel            { return m.focusedPanel }
func (m Model) SessionCursor() int                { return m.sessionCursor }
func (m Model) SessionOffset() int                { return m.sessionOffset }
func (m Model) ChatScroll() int                   { return m.chatScroll }
func (m Model) FilesScroll() int                  { return m.filesScroll }
func (m Model) ProjectCursor() int                { return m.projectCursor }
func (m Model) Mode() Mode                        { return m.mode }
func (m Model) SearchInput() textinput.Model      { return m.searchInput }
func (m Model) SearchQuery() string               { return m.searchQuery }
func (m Model) ProjectFilter() string             { return m.projectFilter }
func (m Model) FileTree() *ui.FileNode            { return m.fileTree }
func (m Model) FileLines() []string               { return m.fileLines }
func (m Model) ToolUsage() map[string]int         { return m.toolUsage }
func (m Model) StatusMessage() string             { return m.statusMessage }
func (m Model) Spinner() spinner.Model            { return m.spinner }
func (m Model) LoadStart() time.Time              { return m.loadStart }

func (m Model) SessionListHeight() int { return m.height - 6 }
func (m Model) ChatViewHeight() int    { return m.height - 6 }
func (m Model) FilesViewHeight() int   { return m.height - 8 }

func (m Model) MaxChatScroll() int {
	if m.selectedSession == nil || !m.selectedSession.MessagesLoaded {
		return 0
	}

	totalLines := 0
	for _, msg := range m.selectedSession.Messages {
		if msg == nil {
			continue
		}
		switch msg.Type {
		case session.MessageTypeHuman, session.MessageTypeAssistant:
			totalLines += len(msg.Content)/50 + 2
		case session.MessageTypeToolUse:
			totalLines++
		}
	}

	maxScroll := totalLines - m.ChatViewHeight()
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}
