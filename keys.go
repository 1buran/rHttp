package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type KeyMap struct {
	Next, Prev, Quit, Help, Run, FullScreen, PageUp, PageDown, Up, Down, Enter,
	Delete, Autocomplete, LoadSession, SaveSession, ToggleCheckbox key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Next, k.Prev, k.Enter, k.Run, k.Delete, k.ToggleCheckbox},
		{k.PageDown, k.PageUp, k.Up, k.Down, k.Autocomplete},
		{k.FullScreen, k.Help, k.Quit, k.LoadSession, k.SaveSession},
	}
}

var keys = KeyMap{
	LoadSession: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "load session"),
	),
	SaveSession: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save session"),
	),
	Autocomplete: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "autocomplete"),
	),
	Next: key.NewBinding(
		key.WithKeys("shift+right"),
		key.WithHelp("shift+right", "next item"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+left"),
		key.WithHelp("shift+left", "prev item"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q", "ctrl+c"),
		key.WithHelp("Ctrl+q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("Ctrl+h", "toggle help"),
	),
	Delete: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("Ctrl+d", "delete item"),
	),
	Run: key.NewBinding(
		key.WithKeys("ctrl+g"),
		key.WithHelp("Ctrl+g", "run request"),
	),
	FullScreen: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("Ctrl+f", "toggle full screen"),
	),
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("Up", "scroll up select options"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("Down", "scroll down select options"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("PgUp", "scroll up body of resposne"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("PgDn", "scroll down body of response"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "set value"),
	),
	ToggleCheckbox: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("Space", "toggle checkbox"),
	),
}

// Helper struct for linking together help and key bindings.
type KeyStroke struct {
	keys KeyMap
	help help.Model
}

// Create new instance of [KeyStroke].
func NewKeyStroke() KeyStroke {
	h := help.New()
	h.Styles.FullKey = h.Styles.FullKey.Foreground(lipgloss.Color("219"))
	h.Styles.FullDesc = h.Styles.FullDesc.Foreground(lipgloss.Color("213"))
	h.Styles.ShortKey = h.Styles.FullKey
	h.Styles.ShortDesc = h.Styles.FullDesc
	return KeyStroke{keys: keys, help: h}
}
