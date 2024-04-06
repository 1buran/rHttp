package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	hint    []string // suggestions
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

func (f *FileInput) Blur() {
	f.widget.Blur()
}

var fpathPrev string

func (f *FileInput) updateSuggestions() ([]string, error) {
	fpath := f.widget.Value()
	if fpath == fpathPrev {
		return nil, errors.New("path is not changed")
	}
	fpathPrev = fpath
	inputDirName, inputFileName := filepath.Split(fpath)
	dirEntries, err := os.ReadDir(inputDirName)
	if err != nil {
		return nil, err
	}
	suggestions := []string{}
	for _, i := range dirEntries {
		s := i.Name()
		if i.IsDir() {
			s += "/"
		}

		if strings.Contains(s, inputFileName) {
			suggestions = append(suggestions, s)
		}

		if len(suggestions) == 10 {
			break
		}
	}
	return suggestions, nil
}

var suggestionStyle lipgloss.Style = lipgloss.NewStyle().
	Height(5).Width(35).Foreground(lipgloss.Color("247"))

func (f *FileInput) getSuggestions() string {
	return suggestionStyle.Render(f.hint...)
}

func (f FileInput) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, f.widget.View(), f.getSuggestions())
}

func NewFileInput(id, mode int, title, placeholder string) FileInput {
	w := textinput.New()
	w.Prompt = title
	w.Placeholder = placeholder
	w.Width = 25
	w.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("177")).Bold(true)
	w.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	w.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("219"))
	return FileInput{id: id, mode: mode, widget: w}
}

type FileInputReader struct {
	Id     int
	Reader io.Reader
	Error  error
	Path   string
}

type FileInputWriter struct {
	Id     int
	Writer io.Writer
	Error  error
	Path   string
}

type FileInputTickMsg time.Time

func FileInputDoTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return FileInputTickMsg(t)
	})
}

func (f FileInput) Update(msg tea.Msg) (FileInput, tea.Cmd) {
	var c1, c2 tea.Cmd
	switch msg := msg.(type) {
	case FileInputTickMsg:
		if f.visible {
			s, err := f.updateSuggestions()
			if err == nil {
				f.hint = s
			}
			return f, FileInputDoTick()
		}
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
