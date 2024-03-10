package main

import (
	// "fmt"
	"log"
	// "net/http"
	"regexp"
	"slices"
	"strings"
	// "github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	header int = iota
	param
	cookie
)

const (
	hotPink  = lipgloss.Color("69")
	darkGray = lipgloss.Color("#767676")
	cyan     = lipgloss.Color("#B6FFFF")
)

var (
	promptStyle   = lipgloss.NewStyle().Foreground(hotPink).Bold(true)
	textStyle     = lipgloss.NewStyle().Foreground(cyan)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
	uriStyle      = lipgloss.NewStyle().Foreground(hotPink)
	headerStyle   = lipgloss.NewStyle().Foreground(cyan)
)

type KeyVal map[string]string
type orderedKeyVal struct {
	order []string
	KeyVal
}

type request struct {
	method  string
	uri     string
	headers orderedKeyVal
	params  KeyVal
	cookies KeyVal
}

func NewOrderedKeVal() orderedKeyVal {
	return orderedKeyVal{order: []string{}, KeyVal: make(KeyVal)}
}

func headerValidator(s string) error {
	// TODO add header validation
	// https://developers.cloudflare.com/rules/transform/request-header-modification/reference/header-format/
	// 	The name of the HTTP request header you want to set or remove can only contain:

	// Alphanumeric characters: a-z, A-Z, and 0-9
	// The following special characters: - and _
	// The value of the HTTP request header you want to set can only contain:

	// Alphanumeric characters: a-z, A-Z, and 0-9
	// The following special characters: _ :;.,\/"'?!(){}[]@<>=-+*#$&`|~^%
	return nil
}

func paramValidator(s string) error  { return nil }
func cookieValidator(s string) error { return nil }

func (r *request) AddHeader(k, v string) {
	// TODO use regexp instead
	k = strings.Trim(k, " \n\t-")
	k = strings.ReplaceAll(k, " ", "-")

	re := regexp.MustCompile(`\b(\w)`)
	k = re.ReplaceAllStringFunc(k, func(s string) string { return strings.ToUpper(s) })
	if k != "" {
		// TODO check if already setted
		r.headers.KeyVal[k] = v
		r.headers.order = append(r.headers.order, k)
	}
}

func (r *request) AddParam(k, v string) {
	// TODO add param validation
	if k != "" {
		r.params[k] = v
	}
}

func (r *request) AddCookie(k, v string) {
	// TODO add param validation
	if k != "" {
		r.cookies[k] = v
	}
}

func (r request) String() string {
	var b strings.Builder

	// TODO url encode with query string
	b.WriteString(uriStyle.Render(r.method+" "+r.uri) + "\n")
	slices.Sort(r.headers.order)
	for _, k := range r.headers.order {
		// TODO Set-Cookie: ...
		v := r.headers.KeyVal[k]
		b.WriteString(headerStyle.Render(k+": "+v) + "\n")
	}

	return b.String()
}

func NewReq() *request {
	return &request{
		"GET",
		"http://example.com",
		NewOrderedKeVal(),
		make(KeyVal),
		make(KeyVal),
	}
}

type model struct {
	req     *request
	actions []string
	cursor  int
	// TODO add inputs for values
	inputs   []textinput.Model
	focused  int
	hideMenu bool
}

// nextInput focuses the next input field
func (m *model) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *model) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}

func initialModel() model {
	var inputs []textinput.Model = make([]textinput.Model, 3)
	inputs[header] = textinput.New()
	inputs[header].Placeholder = "X-Custom-Header-Name"
	inputs[header].Prompt = "Header "
	inputs[header].Width = 30
	inputs[header].TextStyle = textStyle
	inputs[header].PromptStyle = promptStyle
	inputs[header].Validate = headerValidator
	inputs[header].PlaceholderStyle = continueStyle

	inputs[param] = textinput.New()
	inputs[param].Placeholder = "products_id"
	inputs[param].Prompt = "Param  "
	inputs[param].Width = 30
	inputs[param].TextStyle = textStyle
	inputs[param].PromptStyle = promptStyle
	inputs[param].Validate = paramValidator
	inputs[param].PlaceholderStyle = continueStyle

	inputs[cookie] = textinput.New()
	inputs[cookie].Placeholder = "XDEBUG_SESSION"
	inputs[cookie].Prompt = "Cookie "
	inputs[cookie].Width = 30
	inputs[cookie].TextStyle = textStyle
	inputs[cookie].PromptStyle = promptStyle
	inputs[cookie].Validate = cookieValidator
	inputs[cookie].PlaceholderStyle = continueStyle

	return model{
		actions: []string{"add header", "add query string param", "add cookie"},
		req:     NewReq(),
		inputs:  inputs,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyShiftTab:
			m.prevInput()
		case tea.KeyTab:
			m.nextInput()
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			m.req.AddHeader(m.inputs[m.focused].Value(), "Val")
			m.inputs[m.focused].Reset()
			// if m.focused == len(m.inputs)-1 {
			// 	return m, tea.Quit
			// }
			m.nextInput()
		}
		for i := 0; i < len(m.inputs); i++ {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	case error:
		log.Println("error: ", msg)
		return m, nil
	}

	for i := 0; i < len(m.inputs); i++ {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)

}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("Modify http request:\n")
	b.WriteString(m.inputs[header].View() + "\n")
	b.WriteString(m.inputs[param].View() + "\n")
	b.WriteString(m.inputs[cookie].View() + "\n")
	b.WriteString("\n" + m.req.String())
	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
