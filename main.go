package main

import (
	"io"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Number of input or ENUM(input type, int)
const (
	host int = iota
	proto
	method
	urlPath

	header
	headerVal

	param
	paramVal

	cookie
	cookieVal

	form
	formVal

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
	lightPink     = lipgloss.Color("225")
	yellow        = lipgloss.Color("220")

	// screen terminal dimensions
	width       = 96
	columnWidth = 30
)

var (
	promptStyle      = lipgloss.NewStyle().Foreground(hotPink).Bold(true)
	textStyle        = lipgloss.NewStyle().Foreground(purple)
	textValueStyle   = lipgloss.NewStyle().Foreground(brightPurple)
	continueStyle    = lipgloss.NewStyle().Foreground(darkGray)
	uriStyle         = lipgloss.NewStyle().Foreground(hotPink)
	headerStyle      = textStyle
	headerValueStyle = lipgloss.NewStyle().Foreground(brightPurple)
	urlStyle         = lipgloss.NewStyle().Foreground(brightPurple2).Bold(true)
	bodyStyle        = lipgloss.NewStyle().Foreground(lightPink)

	titleStyle = lipgloss.NewStyle().Foreground(lightBlue).
			Bold(true).BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1).Width(60).AlignHorizontal(lipgloss.Center)
)

func NewReq() (r *http.Request) {
	r, _ = http.NewRequest("GET", "http://localhost", nil)
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

const (
	statusInfoEmoji    = "🟢"
	statusWarningEmoji = "🟡"
	statusErrorEmoji   = "🔴"
)

const (
	statusInfo int = iota
	statusWarning
	statusError
)

// A status bar state.
type StatusBar struct {
	status   int
	text     string
	reqCount int
}

// Get status message.
func (s *StatusBar) getStatusText() (t string) {
	var style lipgloss.Style

	switch s.status {
	case statusInfo:
		style = statusTextInfo
	case statusWarning:
		style = statusTextWarning
	case statusError:
		style = statusTextError
	}

	return style.Render(s.text)
}

// The model is a state of app
type model struct {
	req       *http.Request
	res       *http.Response
	inputs    []textinput.Model
	cursorIdx int    // edit type
	cursorKey string // edit key of type orderedKeyVal store
	focused   int
	hideMenu  bool
	resBody   []byte

	StatusBar
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

// Request is executed.
func (m *model) reqIsExecuted() bool {
	return m.res != nil && m.res.StatusCode > 0
}

// Set status.
func (m *model) setStatus(s int, t string) {
	m.StatusBar.status = s
	m.StatusBar.text = time.Now().Format("[15:04:05] ") + t
}

// Increment count of request.
func (m *model) incReqCount() {
	m.StatusBar.reqCount++
}

// Get count of executed requests.
func (m *model) getReqCount() int {
	return m.StatusBar.reqCount
}

// Get status logo indicator
func (m *model) getStatusIndicator() (ind string) {
	ind = `💜 rHttp`
	if m.reqIsExecuted() {
		switch {
		case m.res.StatusCode >= 400:
			ind = statusErrorEmoji
		case m.res.StatusCode >= 300:
			ind = statusWarningEmoji
		case m.res.StatusCode >= 100:
			ind = statusInfoEmoji
		default:
			ind = `🤔`
		}
		ind += " " + m.req.Proto
	}
	return
}

var (
	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	encodingStyle = statusNugget.Copy().
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	statusText        = lipgloss.NewStyle().Inherit(statusBarStyle)
	statusTextInfo    = lipgloss.NewStyle().Inherit(statusText)
	statusTextError   = lipgloss.NewStyle().Inherit(statusText).Foreground(lightPink)
	statusTextWarning = lipgloss.NewStyle().Inherit(statusText).Foreground(yellow)
	indicatorStyle    = statusNugget.Copy().Background(lipgloss.Color("#6124DF"))
)

// Format status bar.
func (m *model) formatStatusBar() string {
	w := lipgloss.Width

	status := statusStyle.Render("STATUS")
	reqCount := encodingStyle.Render(strconv.Itoa(m.getReqCount()))
	indicator := indicatorStyle.Render(m.getStatusIndicator())
	statusVal := statusText.Copy().
		Width(width - w(status) - w(reqCount) - w(indicator)).
		Render(m.getStatusText())

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		status,
		statusVal,
		reqCount,
		indicator,
	)

	return statusBarStyle.Width(width).Render(bar)
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

var allowedMethods = []string{
	"GET", "POST", "PUT", "PATCH", "HEAD", "DELETE", "OPTIONS", "PROPFIND", "SEARCH",
	"TRACE", "PATCH", "PUT", "CONNECT",
}

var prompts = [fieldsCount]string{
	"Host   ", "HTTP/1.", "Method ", "Path  ",
	"Header ", "", "Param  ", "", "Cookie ", "", "Form   ", ""}
var placeholders = [fieldsCount]string{
	"example.com", "1", "GET", "/",
	"X-Auth-Token", "token value", "products_id", "10",
	"XDEBUG_SESSION", "debugger", "login", "user"}

func NewKeyValInputs(n int) textinput.Model {
	t := textinput.New()
	t.Prompt = prompts[n]
	t.Placeholder = placeholders[n]
	t.Width = 30
	t.PromptStyle = promptStyle
	t.PlaceholderStyle = continueStyle
	t.TextStyle = textValueStyle

	// set defaults input text
	switch n {
	case proto:
		t.SetValue("1")
		t.SetSuggestions([]string{"0", "1"})
		t.ShowSuggestions = true
		t.Width = 1
		t.CharLimit = 1
	case host:
		t.SetValue("localhost")
	case method:
		t.SetValue("GET")
		t.SetSuggestions(allowedMethods)
		t.ShowSuggestions = true
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

var formValues = make(url.Values)

type readCloser struct {
	strings.Reader
}

func (rc *readCloser) Close() error      { return nil }
func newReadCloser(s string) *readCloser { return &readCloser{*strings.NewReader(s)} }

func eraseIfError(t textinput.Model) {
	if t.Err != nil {
		t.Reset()
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case *http.Response:
		defer msg.Body.Close()
		m.resBody, _ = io.ReadAll(msg.Body)
		m.res = msg
		m.setStatus(statusInfo, "request is executed, response taken")

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlG:
			m.setStatus(statusInfo, "sending request...")
			m.incReqCount()
			cmd := func() tea.Msg {
				r, err := sendRequest(m.req)
				if err != nil {
					return err
				}
				return r
			}
			return m, cmd
		case tea.KeyShiftTab:
			m.prevInput()
		case tea.KeyTab:
			switch m.focused {
			case proto:
				val := m.inputs[proto].Value()
				if val != "0" && val != "1" {
					m.inputs[proto].SetValue("1")
					m.inputs[proto].CursorEnd()
				} else {
					m.req.Proto = "HTTP/1." + val
				}
			case method:
				r := regexp.MustCompile(`(?i)\b` + m.inputs[method].Value())
				if r.MatchString(strings.Join(allowedMethods, " ")) {
					m.inputs[method].SetValue(m.inputs[method].CurrentSuggestion())
					m.req.Method = m.inputs[method].Value()
					m.inputs[method].SetCursor(len(m.req.Method))
				} else { // not allowed HTTP method, reset to last saved
					m.inputs[method].SetValue(m.req.Method)
					m.inputs[method].CursorEnd()
				}
			}
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
			case form, formVal:
				name := m.inputs[form].Value()
				val := m.inputs[formVal].Value()
				if name != "" && val != "" {
					formValues.Add(name, val)
					m.req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					m.req.Body = newReadCloser(formValues.Encode())
					m.inputs[form].Reset()
					m.inputs[formVal].Reset()
				}
			case method:
				// disallow changing the value by enter
				m.inputs[method].SetValue(m.req.Method)
				m.inputs[method].CursorEnd()
			case urlPath:
				val := m.inputs[urlPath].Value()
				m.req.URL.Path = val
			case host:
				val := m.inputs[host].Value()
				m.req.URL.Host = val
			case proto:
				val := m.inputs[proto].Value()
				if val != "0" && val != "1" {
					m.inputs[proto].SetValue("1")
					m.inputs[proto].CursorEnd()
				}
				m.req.Proto = "HTTP/1." + val
			}

			// after handling enter is done, go to next input..
			m.nextInput()
		}
		for i := 0; i < len(m.inputs); i++ {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	case error:
		m.setStatus(statusError, msg.Error())
		return m, nil
	}

	for i := 0; i < len(m.inputs); i++ {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)

}

func headersPrintf(o io.StringWriter, h http.Header) {
	var order []string
	for k := range h {
		order = append(order, k)
	}

	slices.Sort(order)

	// print headers
	for _, name := range order {
		v := h[name]
		o.WriteString(headerStyle.Render(name+": ") + headerValueStyle.Render(strings.Join(v, ", ")) + "\n")
	}
}

func (m model) View() string {
	var b strings.Builder

	// print prompts
	for i := 0; i < len(m.inputs); i += 2 {
		b.WriteString(m.inputs[i].View() + " " + m.inputs[i+1].View() + "\n")
	}

	// print result URL
	b.WriteString(
		"\n" + urlStyle.Render(
			m.req.Proto+" "+m.req.Method+" "+m.req.URL.String()) + "\n")

	headersPrintf(&b, m.req.Header)

	// print body
	if m.req.Body != nil {
		b.WriteString("\n" + bodyStyle.Render(formValues.Encode()))
	}

	// print response
	if m.reqIsExecuted() {
		b.WriteString(
			"\n" + urlStyle.Render(m.res.Proto+" "+m.res.Status) + "\n")

		headersPrintf(&b, m.res.Header)

		// TODO..
		// if m.res.Header["Content-Type"] == "application/json" {
		// } else {
		// 	b.WriteString("\n" + string(m.resBody))
		// }

		b.WriteString("\n" + bodyStyle.Render(string(m.resBody)))

	}

	// add status bar
	b.WriteString("\n" + m.formatStatusBar())

	return b.String()
}

func sendRequest(r *http.Request) (*http.Response, error) {
	http_cli := http.Client{Timeout: 2 * time.Second}
	return http_cli.Do(r)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
