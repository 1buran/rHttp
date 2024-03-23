package main

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// A Select, when it state has changed the [SelectUpdated] will be emitted,
// use Id to distinguish selects.
type Select struct {
	title        string
	titleStyle   lipgloss.Style
	options      []string
	optionsStyle lipgloss.Style
	selected     int
}

func NewSelect(title string, titleStyle lipgloss.Style, options []string, optionsStyle lipgloss.Style) Select {
	return Select{title, titleStyle, options, optionsStyle, 0}
}

func (s *Select) Value() string {
	if s.selected < len(s.options)-1 {
		return s.options[s.selected]
	}
	return ""
}
func (s *Select) NextOption() {
	if len(s.options) > 0 {
		s.selected = (s.selected + 1) % len(s.options)
	}
}

func (s *Select) PrevOption() {
	s.selected--
	if s.selected < 0 {
		if len(s.options) > 0 {
			s.selected = len(s.options) - 1
		} else {
			s.selected = 0
		}
	}
}

func (s *Select) AddOption(o string) {
	if slices.Contains(s.options, o) {
		return
	}
	s.options = append(s.options, o)
}

func (s *Select) DelOption(o string) {
	s.options = slices.DeleteFunc(s.options, func(s string) bool { return s == o })
}

func (s *Select) SetTitle(t string) {
	s.title = t
}

func (s Select) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		s.titleStyle.Render(s.title),
		s.optionsStyle.Render(s.Value()),
	)
}

func (s Select) Update(msg tea.Msg) (Select, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyDown:
			s.NextOption()
		case tea.KeyUp:
			s.PrevOption()
		case tea.KeyEnter: // option is selected
			cmd = func() tea.Msg {
				return msg
			}
		}
	}
	return s, cmd
}
