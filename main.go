package main

import (
	"log"
	"net/http"
	"net/textproto"
	"net/url"
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
	hotPink       = lipgloss.Color("69")
	darkGray      = lipgloss.Color("#767676")
	purple        = lipgloss.Color("141")
	brightPurple  = lipgloss.Color("183")
	brightPurple2 = lipgloss.Color("189")
	lightBlue     = lipgloss.Color("12")
)

var (
	promptStyle      = lipgloss.NewStyle().Foreground(hotPink).Bold(true)
	textStyle        = lipgloss.NewStyle().Foreground(purple)
	textValueStyle   = lipgloss.NewStyle().Foreground(brightPurple)
	continueStyle    = lipgloss.NewStyle().Foreground(darkGray)
	uriStyle         = lipgloss.NewStyle().Foreground(hotPink)
	headerStyle      = textStyle
	headerValueStyle = textValueStyle
	urlStyle         = lipgloss.NewStyle().Foreground(brightPurple2).Bold(true)
	titleStyle       = lipgloss.NewStyle().Foreground(lightBlue).
				Bold(true).BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(1).Width(60).AlignHorizontal(lipgloss.Center)
)

func NewReq() (r *http.Request) {
	r, _ = http.NewRequest("GET", "http://example.com", nil)
	return
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
	req *http.Request
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
	t.PromptStyle = promptStyle
	t.Validate = validators[n]
	t.PlaceholderStyle = continueStyle

	switch n {
	case headerVal, paramVal, cookieVal:
		t.TextStyle = textValueStyle
	default:
		t.TextStyle = textStyle
	}

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
			case header, headerVal:
				correctHeader(&m.inputs[header])
				name := m.inputs[header].Value()
				val := m.inputs[headerVal].Value()
				if name != "" && val != "" {
					m.req.Header.Set(name, val)
					m.inputs[header].Reset()
					m.inputs[headerVal].Reset()
				}
			case param, paramVal:
				v, _ := url.ParseQuery(m.req.URL.RawQuery)
				name := m.inputs[param].Value()
				val := m.inputs[paramVal].Value()
				if name != "" && val != "" {
					v.Set(name, val)
					m.req.URL.RawQuery = v.Encode()
					m.inputs[param].Reset()
					m.inputs[paramVal].Reset()
				}
			case cookie, cookieVal:
				name := m.inputs[cookie].Value()
				val := m.inputs[cookieVal].Value()
				if name != "" && val != "" {
					isNew := true
					for _, i := range m.req.Cookies() {
						if i.Name == name && i.Value == val {
							isNew = false
							break
						}
					}
					if isNew {
						m.req.AddCookie(&http.Cookie{Name: name, Value: val})
					}
					m.inputs[cookie].Reset()
					m.inputs[cookieVal].Reset()
				}
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

	b.WriteString(titleStyle.Render("rHTTP v0.0.1") + "\n")

	// print prompts
	for i := 0; i < len(m.inputs); i += 2 {
		b.WriteString(m.inputs[i].View() + " " + m.inputs[i+1].View() + "\n")
	}

	// print result URL
	b.WriteString("\n")
	b.WriteString(
		urlStyle.Render(
			m.req.Proto+" "+m.req.Method+" "+m.req.URL.String()) + "\n")

	var order []string
	for k := range m.req.Header {
		order = append(order, k)
	}

	slices.Sort(order)

	// print headers
	for _, name := range order {
		v := m.req.Header[name]
		b.WriteString(headerStyle.Render(name+": ") + headerValueStyle.Render(strings.Join(v, ", ")) + "\n")
	}
	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
