/*
Package ui provides styling and visual components for the TUI.
*/
package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#00ff88")
	ColorSecondary = lipgloss.Color("#00ccff")
	ColorWarning   = lipgloss.Color("#ffcc00")
	ColorHighlight = lipgloss.Color("#ff66cc")
	ColorDim       = lipgloss.Color("#555555")
	ColorText      = lipgloss.Color("#ffffff")
	ColorSubtle    = lipgloss.Color("#888888")
	ColorError     = lipgloss.Color("#ff5555")
	ColorSuccess   = lipgloss.Color("#00ff88")
	ColorEdited    = lipgloss.Color("#ffcc00")
)

type Styles struct {
	App       lipgloss.Style
	TopBar    lipgloss.Style
	BottomBar lipgloss.Style

	StatusText lipgloss.Style
	HelpText   lipgloss.Style

	PanelBorder       lipgloss.Style
	PanelBorderActive lipgloss.Style
	PanelTitle        lipgloss.Style
	PanelTitleActive  lipgloss.Style

	SessionItem         lipgloss.Style
	SessionItemSelected lipgloss.Style
	SessionProject      lipgloss.Style
	SessionID           lipgloss.Style
	SessionTime         lipgloss.Style
	SessionCount        lipgloss.Style
	DotRecent           lipgloss.Style
	DotWeek             lipgloss.Style
	DotOld              lipgloss.Style

	HumanPrefix      lipgloss.Style
	HumanMessage     lipgloss.Style
	AssistantPrefix  lipgloss.Style
	AssistantMessage lipgloss.Style
	ToolUse          lipgloss.Style
	SystemMessage    lipgloss.Style

	FileCreated lipgloss.Style
	FileEdited  lipgloss.Style
	FileDeleted lipgloss.Style
	FileRead    lipgloss.Style
	DirName     lipgloss.Style
	TreeBranch  lipgloss.Style

	SearchInput       lipgloss.Style
	SearchPrompt      lipgloss.Style
	SearchPlaceholder lipgloss.Style

	Highlight lipgloss.Style
	Dim       lipgloss.Style
	Bold      lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle(),

		TopBar: lipgloss.NewStyle().
			Foreground(ColorText).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1),

		BottomBar: lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1),

		StatusText: lipgloss.NewStyle().Foreground(ColorSubtle),
		HelpText:   lipgloss.NewStyle().Foreground(ColorDim),

		PanelBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDim),

		PanelBorderActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary),

		PanelTitle: lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Bold(true).
			Padding(0, 1),

		PanelTitleActive: lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1),

		SessionItem: lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1),

		SessionItemSelected: lipgloss.NewStyle().
			Foreground(ColorText).
			Background(lipgloss.Color("#2a2a2a")).
			Padding(0, 1),

		SessionProject: lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true),
		SessionID:      lipgloss.NewStyle().Foreground(ColorDim),
		SessionTime:    lipgloss.NewStyle().Foreground(ColorSubtle),
		SessionCount:   lipgloss.NewStyle().Foreground(ColorDim),

		DotRecent: lipgloss.NewStyle().Foreground(ColorSuccess),
		DotWeek:   lipgloss.NewStyle().Foreground(ColorWarning),
		DotOld:    lipgloss.NewStyle().Foreground(ColorDim),

		HumanPrefix:      lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true),
		HumanMessage:     lipgloss.NewStyle().Foreground(ColorText),
		AssistantPrefix:  lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true),
		AssistantMessage: lipgloss.NewStyle().Foreground(ColorText),
		ToolUse:          lipgloss.NewStyle().Foreground(ColorDim).Italic(true),
		SystemMessage:    lipgloss.NewStyle().Foreground(ColorSubtle).Italic(true),

		FileCreated: lipgloss.NewStyle().Foreground(ColorSuccess),
		FileEdited:  lipgloss.NewStyle().Foreground(ColorEdited),
		FileDeleted: lipgloss.NewStyle().Foreground(ColorError),
		FileRead:    lipgloss.NewStyle().Foreground(ColorDim),
		DirName:     lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true),
		TreeBranch:  lipgloss.NewStyle().Foreground(ColorDim),

		SearchInput: lipgloss.NewStyle().
			Foreground(ColorText).
			Background(lipgloss.Color("#2a2a2a")).
			Padding(0, 1),

		SearchPrompt:      lipgloss.NewStyle().Foreground(ColorWarning).Bold(true),
		SearchPlaceholder: lipgloss.NewStyle().Foreground(ColorDim).Italic(true),

		Highlight: lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(ColorDim),
		Bold:      lipgloss.NewStyle().Bold(true),
	}
}

func (s Styles) PanelStyle(focused bool, width, height int) lipgloss.Style {
	if focused {
		return s.PanelBorderActive.Width(width).Height(height)
	}
	return s.PanelBorder.Width(width).Height(height)
}

func (s Styles) SessionDotStyle(age int) lipgloss.Style {
	switch age {
	case 0:
		return s.DotRecent
	case 1:
		return s.DotWeek
	default:
		return s.DotOld
	}
}
