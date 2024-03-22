package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type KeyMap struct {
	Next, Prev, Quit, Help, Run, FullScreen, PageUp, PageDown, Enter key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Next, k.Prev, k.Enter, k.Run},
		{k.PageDown, k.PageUp, k.FullScreen},
		{k.Help, k.Quit},
	}
}

var keys = KeyMap{
	Next: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next item / autocomplete"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev item"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q", "ctrl+c"),
		key.WithHelp("ctrl+q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "toggle help"),
	),
	Run: key.NewBinding(
		key.WithKeys("ctrl+g"),
		key.WithHelp("ctrl+g", "run request"),
	),
	FullScreen: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "toggle full screen"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "scroll up body of resposne"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "scroll down body of response"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "set value / toggle checkbox"),
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
