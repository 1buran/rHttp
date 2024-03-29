package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	off int = iota
	on
	title
)

// A Checkbox, when it state has changed the [CheckboxUpdated] will be emitted,
// use Id to distinguish checkboxes.
type Checkbox struct {
	id    int
	style []lipgloss.Style
	text  []string
	state int
}

type CheckboxUpdated struct {
	Id int
	On bool
}

func NewCheckbox(
	id int, titleText, onText, offText string,
	titleStyle, onStyle, offStyle lipgloss.Style) Checkbox {
	return Checkbox{
		id:    id,
		text:  []string{offText, onText, titleText},
		style: []lipgloss.Style{offStyle, onStyle, titleStyle},
	}
}

func (c Checkbox) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		c.style[title].Render(c.text[title]),
		c.style[c.state].Render(c.text[c.state]),
	)
}

func (c *Checkbox) IsOn() bool {
	return c.state == on
}

func (c *Checkbox) SetOn() {
	c.state = on
}

func (c *Checkbox) SetOff() {
	c.state = off
}

func (c Checkbox) Update(msg tea.Msg) (Checkbox, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeySpace: // toggle checkbox
			if c.state == on {
				c.state = off
			} else {
				c.state = on
			}
		}
	}
	cmd := func() tea.Msg {
		return CheckboxUpdated{Id: c.id, On: c.IsOn()}
	}
	return c, cmd
}
