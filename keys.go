/*
Keyboard bindings and help text for TUI navigation and actions.
*/
package main

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	HalfUp   key.Binding
	HalfDown key.Binding
	Top      key.Binding
	Bottom   key.Binding

	NextPanel key.Binding
	PrevPanel key.Binding

	Select  key.Binding
	Resume  key.Binding
	Delete  key.Binding
	Search  key.Binding
	Filter  key.Binding
	Help    key.Binding
	Quit    key.Binding
	Escape  key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("j", "down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
		HalfUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("C-u", "half up")),
		HalfDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("C-d", "half down")),
		Top:      key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g", "top")),
		Bottom:   key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "bottom")),

		NextPanel: key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "next panel")),
		PrevPanel: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-Tab", "prev panel")),

		Select:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "select")),
		Resume:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "resume")),
		Delete:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Search:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Filter:  key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "filter project")),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Escape:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "back")),
		Confirm: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm")),
		Cancel:  key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "cancel")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Search, k.Resume, k.Quit, k.Help}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.HalfUp, k.HalfDown, k.Top, k.Bottom},
		{k.NextPanel, k.PrevPanel},
		{k.Select, k.Resume, k.Delete},
		{k.Search, k.Filter, k.Help, k.Quit},
	}
}

func HelpString(keys KeyMap, focused Panel, searchActive bool) string {
	if searchActive {
		return "Enter: search | Esc: cancel"
	}

	switch focused {
	case PanelSessions:
		return "j/k: nav | Enter: select | r: resume | /: search | p: project | ?: help | q: quit"
	case PanelChat:
		return "j/k: scroll | C-u/C-d: half page | g/G: top/bottom | r: resume | Tab: switch | q: quit"
	case PanelFiles:
		return "j/k: scroll | Tab: switch | q: quit"
	default:
		return "Tab: switch panels | ?: help | q: quit"
	}
}

func FullHelpText() string {
	return `
+-------------------------------------------------------------+
|                    Claude Code Manager                      |
|                      Keyboard Shortcuts                     |
+-------------------------------------------------------------+
|  Navigation                                                 |
|    j/k           Move up/down                               |
|    PgUp/PgDn     Page up/down                               |
|    Ctrl+u/d      Half page up/down                          |
|    g/G           Jump to top/bottom                         |
|    Tab/S-Tab     Switch between panels                      |
|                                                             |
|  Session Actions                                            |
|    Enter         Select/expand session                      |
|    r             Resume session in Claude CLI               |
|    d             Delete session (with confirmation)         |
|                                                             |
|  Search & Filter                                            |
|    /             Search sessions                            |
|    p             Filter by project                          |
|    Esc           Clear search/filter                        |
|                                                             |
|  General                                                    |
|    ?             Toggle this help                           |
|    q / Ctrl+c    Quit                                       |
+-------------------------------------------------------------+

                    Press any key to close
`
}
