package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines every key binding used in the TUI.
// It implements the help.KeyMap interface so the built-in help bubble
// can render contextual shortcut hints.
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	NextTab key.Binding
	PrevTab key.Binding
	Search  key.Binding
	Install key.Binding
	Remove  key.Binding
	Info    key.Binding
	Update  key.Binding
	Select  key.Binding
	Confirm key.Binding
	Cancel  key.Binding
	Refresh key.Binding
	Help    key.Binding
	Quit    key.Binding
	Back    key.Binding
}

// DefaultKeyMap returns a KeyMap populated with the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab", "l"),
			key.WithHelp("tab/l", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab", "h"),
			key.WithHelp("shift+tab/h", "prev tab"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Install: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "install"),
		),
		Remove: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "remove"),
		),
		Info: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "info"),
		),
		Update: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "update"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "cancel"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// ShortHelp returns the key bindings shown in the compact help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Search,
		k.Install,
		k.Help,
		k.Quit,
	}
}

// FullHelp returns the key bindings organized into groups for the
// expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation
		{k.Up, k.Down, k.NextTab, k.PrevTab, k.Back},
		// Actions
		{k.Search, k.Install, k.Remove, k.Update, k.Select},
		// UI
		{k.Info, k.Confirm, k.Cancel, k.Refresh, k.Help, k.Quit},
	}
}
