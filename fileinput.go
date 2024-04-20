package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"os"
)

const (
	ReadMode = iota
	WriteMode
)

type FileInput struct {
	id      int
	mode    int
	widget  textinput.Model
	visible bool
}

type OpenFileReadMsg struct{ Id int }
type OpenFileWriteMsg struct{ Id int }

func (f *FileInput) OpenFile() tea.Cmd {
	switch f.mode {
	case ReadMode:
		return func() tea.Msg {
			return OpenFileReadMsg{f.id}
		}
	case WriteMode:
		return func() tea.Msg {
			return OpenFileWriteMsg{f.id}
		}
	}
	return nil
}

func (f *FileInput) SetTitle(s string) {
	f.widget.Prompt = s
}

func (f *FileInput) SetVisible() {
	f.visible = true
}

func (f *FileInput) Hide() {
	f.widget.Blur()
	f.visible = false
}

func (f *FileInput) IsVisible() bool {
	return f.visible
}

func (f *FileInput) Focus() {
	f.widget.Focus()
}

func (f *FileInput) Value() string {
	return f.widget.Value()
}

func (f *FileInput) Reset() {
	f.widget.Reset()
}

func (f *FileInput) Blur() {
	f.widget.Blur()
}

func (f FileInput) View() string {
	return f.widget.View()
}

func NewFileInput(id, mode int, title, placeholder string, colors ...lipgloss.Color) FileInput {
	w := textinput.New()
	w.Prompt = title
	w.Placeholder = placeholder
	w.Width = 25
	w.PromptStyle = lipgloss.NewStyle().Foreground(colors[0]).Bold(true)
	w.PlaceholderStyle = lipgloss.NewStyle().Foreground(colors[1])
	w.TextStyle = lipgloss.NewStyle().Foreground(colors[2])
	return FileInput{id: id, mode: mode, widget: w}
}

type FileInputReader struct {
	Id     int
	Reader io.ReadCloser
	Error  error
	Path   string
}

type FileInputWriter struct {
	Id     int
	Writer io.WriteCloser
	Error  error
	Path   string
}

func (f FileInput) Update(msg tea.Msg) (FileInput, tea.Cmd) {
	var c1, c2 tea.Cmd
	switch msg := msg.(type) {
	case OpenFileReadMsg:
		if msg.Id != f.id {
			return f, nil
		}
		p := f.widget.Value()
		r, err := os.Open(p)
		c2 = func() tea.Msg {
			return FileInputReader{f.id, r, err, p}
		}
		f.Hide()
	case OpenFileWriteMsg:
		if msg.Id != f.id {
			return f, nil
		}
		p := f.widget.Value()
		w, err := os.Create(p)
		c2 = func() tea.Msg {
			return FileInputWriter{f.id, w, err, p}
		}
		f.Hide()
	}
	f.widget, c1 = f.widget.Update(msg)
	return f, tea.Batch(c1, c2)
}
