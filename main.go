package main

import (
	"log"
	"net/http"
	"net/textproto"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Number of input or ENUM(input type, int)
const (
	header int = iota
	headerVal
	param
	paramVal
	cookie
	cookieVal

	// the last one is the max index of defined constants
	fieldsCount
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

func (kv *orderedKeyVal) setKey(i *textinput.Model) {
	k := i.Value()
	if k != "" {
		_, ok := kv.KeyVal[k]
		if !ok {
			kv.order = append(kv.order, k)
			kv.KeyVal[k] = ""
		}
	}
}

func (kv *orderedKeyVal) setValue(i1, i2 *textinput.Model) {
	kv.KeyVal[i1.Value()] = i2.Value()
	i1.Reset()
	i2.Reset()
}

type request struct {
	method  string
	uri     string
	headers orderedKeyVal
	params  orderedKeyVal
	cookies orderedKeyVal
}

func (r request) String() string {
	var b strings.Builder

	// TODO url encode with query string
	b.WriteString(uriStyle.Render(r.method+" "+r.uri) + "\n")
	slices.Sort(r.headers.order)
	for _, k := range r.headers.order {
		v := r.headers.KeyVal[k]
		b.WriteString(headerStyle.Render(k+": "+v) + "\n")
	}

	// TODO add prams to URL query string: ...
	b.WriteString("params:" + "\n")
	for _, k := range r.params.order {
		v := r.params.KeyVal[k]
		b.WriteString(headerStyle.Render(k+": "+v) + "\n")
	}

	// TODO Set-Cookie: ...
	b.WriteString("cookies:" + "\n")
	for _, k := range r.cookies.order {
		v := r.cookies.KeyVal[k]
		b.WriteString(headerStyle.Render(k+": "+v) + "\n")
	}

	return b.String()
}

func NewReq() *request {
	return &request{
		"GET", "http://example.com", NewOrderedKeVal(), NewOrderedKeVal(), NewOrderedKeVal()}
}
func NewOrderedKeVal() orderedKeyVal {
	return orderedKeyVal{order: []string{}, KeyVal: make(KeyVal)}
}

func correctHeader(i *textinput.Model) {
	// TODO use regexp instead
	h := i.Value()
	h = textproto.TrimString(h)
	h = strings.ReplaceAll(h, " ", "-")
	h = http.CanonicalHeaderKey(h)

	// auto correct header name
	i.SetValue(h)
}

type model struct {
	req *request
	// TODO add inputs for values
	inputs    []textinput.Model
	cursorIdx int    // edit type
	cursorKey string // edit key of type orderedKeyVal store
	focused   int
	hideMenu  bool
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

func paramValidator(s string) error   { return nil }
func cookieValidator(s string) error  { return nil }
func defaultvalidator(s string) error { return nil }

var validators = [fieldsCount]func(string) error{
	// field name input and field value input validators
	headerValidator, defaultvalidator,
	paramValidator, defaultvalidator,
	cookieValidator, defaultvalidator,
}
var prompts = [fieldsCount]string{"Header ", "", "Param  ", "", "Cookie ", ""}
var placeholders = [fieldsCount]string{"X-Auth-Token", "token value", "products_id", "10", "XDEBUG_SESSION", "debugger"}

func NewKeyValInputs(n int) textinput.Model {
	t := textinput.New()
	t.Prompt = prompts[n]
	t.Placeholder = placeholders[n]
	t.Width = 30
	t.TextStyle = textStyle
	t.PromptStyle = promptStyle
	t.Validate = validators[n]
	t.PlaceholderStyle = continueStyle
	return t
}

func initialModel() model {
	var inputs []textinput.Model
	for i := 0; i < fieldsCount; i++ {
		inputs = append(inputs, NewKeyValInputs(i))
	}
	return model{
		req:    NewReq(),
		inputs: inputs,
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
			switch m.focused {
			case header:
				correctHeader(&m.inputs[header])
				m.req.headers.setKey(&m.inputs[header])
			case param:
				m.req.params.setKey(&m.inputs[param])
			case cookie:
				m.req.cookies.setKey(&m.inputs[cookie])
			case headerVal:
				m.req.headers.setValue(&m.inputs[header], &m.inputs[headerVal])
			case paramVal:
				m.req.params.setValue(&m.inputs[param], &m.inputs[paramVal])
			case cookieVal:
				m.req.cookies.setValue(&m.inputs[cookie], &m.inputs[cookieVal])
			}

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

	for i := 0; i < len(m.inputs); i += 2 {
		b.WriteString(m.inputs[i].View() + " " + m.inputs[i+1].View() + "\n")
	}

	b.WriteString("\n" + m.req.String())
	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
