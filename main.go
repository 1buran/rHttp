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

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
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
	lightPink     = lipgloss.Color("225")
	darkGray      = lipgloss.Color("#767676")
	purple        = lipgloss.Color("141")
	brightPurple  = lipgloss.Color("183")
	brightPurple2 = lipgloss.Color("189")
	lightBlue     = lipgloss.Color("12")
)

var (
	screenWidth  = 100
	screenHeight = 50
	offsetShift  = 5
	baseStyle    = lipgloss.NewStyle().Width(screenWidth)

	promptStyle      = lipgloss.NewStyle().Foreground(hotPink).Bold(true)
	textStyle        = lipgloss.NewStyle().Foreground(purple)
	textValueStyle   = lipgloss.NewStyle().Foreground(brightPurple)
	continueStyle    = lipgloss.NewStyle().Foreground(darkGray)
	uriStyle         = lipgloss.NewStyle().Foreground(hotPink)
	headerStyle      = textStyle
	headerValueStyle = lipgloss.NewStyle().Foreground(brightPurple)
	urlStyle         = lipgloss.NewStyle().Inherit(baseStyle).
				Foreground(brightPurple2).
				Bold(true).Padding(0, 1)

	bodyStyle  = lipgloss.NewStyle().Inherit(baseStyle).Foreground(lightPink)
	titleStyle = lipgloss.NewStyle().Foreground(lightBlue).
			Bold(true).BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1).Width(60).AlignHorizontal(lipgloss.Center)
)

func newReqest() (r *http.Request) {
	r, _ = http.NewRequest("GET", "http://localhost", nil)
	return
}

func sendRequest(r *http.Request) (*http.Response, error) {
	http_cli := http.Client{Timeout: 2 * time.Second}
	return http_cli.Do(r)
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

func headersPrintf(h http.Header) []string {
	var order, lines []string
	for k := range h {
		order = append(order, k)
	}

	slices.Sort(order)

	// print headers
	for _, name := range order {
		val := strings.Join(h[name], ", ")
		nameRendered := headerStyle.Padding(0, 1).Render(name + ":")
		lines = append(
			lines, lipgloss.JoinHorizontal(
				lipgloss.Top,
				nameRendered,
				headerValueStyle.Padding(0, 1).
					Width(screenWidth-lipgloss.Width(nameRendered)).
					Render(val),
			),
		)
	}
	return lines
}

// The model is a state of app
type model struct {
	req          *http.Request
	res          *http.Response
	inputs       []textinput.Model
	cursorIdx    int    // edit type
	cursorKey    string // edit key of type orderedKeyVal store
	focused      int
	hideMenu     bool
	resBodyLines []string
	fullScreen   bool
	offset       int
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
	return m.res != nil
}

// Clear response artefacts.
func (m *model) clearRespArtefacts() {
	m.res = nil
	m.resBodyLines = nil
	m.offset = 0
}

// Get page of response.
func (m *model) getRespPageLines() []string {
	limit := screenHeight - usedScreenLines - 2 // available screen lines for display of res body
	end := m.offset + limit
	if end > len(m.resBodyLines)-1 {
		return m.resBodyLines[m.offset:]
	}
	return m.resBodyLines[m.offset:end]
}

func (m *model) setReqHeader() {
	correctHeader(&m.inputs[header])
	name := m.inputs[header].Value()
	val := m.inputs[headerVal].Value()
	if name != "" && val != "" {
		m.req.Header.Set(name, val)
		m.inputs[header].Reset()
		m.inputs[headerVal].Reset()
	}
}

func (m *model) setReqParam() {
	v, _ := url.ParseQuery(m.req.URL.RawQuery)
	name := m.inputs[param].Value()
	val := m.inputs[paramVal].Value()
	if name != "" && val != "" {
		v.Set(name, val)
		m.req.URL.RawQuery = v.Encode()
		m.inputs[param].Reset()
		m.inputs[paramVal].Reset()
	}
}

func (m *model) setReqCookie() {
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

func (m *model) setReqForm() {
	name := m.inputs[form].Value()
	val := m.inputs[formVal].Value()
	if name != "" && val != "" {
		formValues.Add(name, val)
		m.req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		m.req.Body = newReadCloser(formValues.Encode())
		m.inputs[form].Reset()
		m.inputs[formVal].Reset()
	}
}

func (m *model) setReqMethod() {
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

func (m *model) setReqProto() {
	val := m.inputs[proto].Value()
	if val != "0" && val != "1" {
		m.inputs[proto].SetValue("1")
		m.inputs[proto].CursorEnd()
	} else {
		m.req.Proto = "HTTP/1." + val
	}
}

func (m *model) setReqUrlPath() {
	val := m.inputs[urlPath].Value()
	m.req.URL.Path = val
}

func (m *model) setReqHost() {
	val := m.inputs[host].Value()
	m.req.URL.Host = val

}

func (m *model) restoreReqMethod() {
	m.inputs[method].SetValue(m.req.Method)
	m.inputs[method].CursorEnd()
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
		req:    newReqest(),
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

// Match lexer against given content type of http response.
func matchContentTypeTolexer(ct string) chroma.Lexer {
	for _, l := range lexers.GlobalLexerRegistry.Lexers {
		for _, mt := range l.Config().MimeTypes {
			if strings.Contains(ct, mt) {
				return l
			}
		}
	}

	return nil
}

func formatRespBody(ct, s string) []string {
	var content strings.Builder

	// huge one line splitter
	lp := lipgloss.NewStyle().Width(screenWidth).Padding(0, 1)
	s = lp.Render(s)

	lexer := matchContentTypeTolexer(ct)
	if lexer == nil {
		// detect lang
		lexer = lexers.Analyse(s)
	}
	lexer = chroma.Coalesce(lexer)

	// pick a style
	style := styles.Get("catppuccin-mocha")
	if style == nil {
		style = styles.Fallback
	}

	// pick a formatter
	formatter := formatters.Get("terminal16m")
	iterator, err := lexer.Tokenise(nil, s)
	if err != nil {
		// tea.Println(err)
		panic(err)
	}

	err = formatter.Format(&content, style, iterator)
	if err != nil {
		// tea.Println(err)
		panic(err)
	}

	return strings.Split(content.String(), "\n")
}

// Timer is a data container for some payload + time started.
type Timer struct {
	start   time.Time
	payload tea.Msg
}

// New message with timer.
func NewMessageWithTimer(payload any) Timer {
	return Timer{time.Now(), payload}
}

// Elapsed time from start of timer.
func (t *Timer) elapsedTime() time.Duration {
	return time.Since(t.start)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case Timer:
		m.reqTime = msg.elapsedTime()
		cmd := func() tea.Msg {
			return msg.payload
		}
		return m, cmd

	case *http.Response:
		defer msg.Body.Close()
		buf, _ := io.ReadAll(msg.Body)
		m.res = msg
		m.resBodyLines = formatRespBody(m.res.Header.Get("content-type"), string(buf))
		m.setStatus(statusInfo, "request is executed, response taken")

	case tea.WindowSizeMsg:
		m.setStatus(
			statusInfo,
			"detected screen size: "+strconv.Itoa(msg.Width)+" x "+strconv.Itoa(msg.Height))
		screenWidth = msg.Width
		screenHeight = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyPgDown:
			availableScreenLines := screenHeight - usedScreenLines - 2
			if m.offset+offsetShift+availableScreenLines <= len(m.resBodyLines) {
				m.offset += offsetShift
			} else {
				// decrease offset to take one last page in full size of screen lines
				m.offset += len(m.resBodyLines) - availableScreenLines - m.offset
				if m.offset < 0 {
					m.offset = 0
				}
			}
		case tea.KeyPgUp:
			if m.offset-offsetShift >= 0 {
				m.offset -= offsetShift
			} else {
				m.offset = 0
			}
		case tea.KeyCtrlF:
			if m.fullScreen {
				m.fullScreen = false
				return m, tea.ExitAltScreen
			}
			m.fullScreen = true
			return m, tea.EnterAltScreen
		case tea.KeyCtrlG:
			m.setStatus(statusInfo, "sending request...")
			m.clearRespArtefacts()
			m.incReqCount()
			cmd := func() tea.Msg {
				r, err := sendRequest(m.req)
				if err != nil {
					return NewMessageWithTimer(err)
				}
				return NewMessageWithTimer(r)
			}
			return m, cmd
		case tea.KeyShiftTab:
			m.prevInput()
		case tea.KeyTab:
			switch m.focused {
			case proto:
				m.setReqProto()
			case method:
				m.setReqMethod()
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyCtrlQ:
			return m, tea.Quit
		case tea.KeyEnter:
			switch m.focused {
			case header, headerVal:
				m.setReqHeader()
			case param, paramVal:
				m.setReqParam()
			case cookie, cookieVal:
				m.setReqCookie()
			case form, formVal:
				m.setReqForm()
			case method:
				m.restoreReqMethod() // disallow changing the value by enter
			case urlPath:
				m.setReqUrlPath()
			case host:
				m.setReqHost()
			case proto:
				m.setReqProto()
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
		m.clearRespArtefacts()
		return m, tea.ClearScreen
	}

	for i := 0; i < len(m.inputs); i++ {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)

}

// Format status bar.
func (m *model) formatStatusBar() string {
	w := lipgloss.Width

	var resStatusCode int
	if m.reqIsExecuted() {
		resStatusCode = m.res.StatusCode
	}

	status := m.getStatusBadge("STATUS")
	reqCounter := reqCountStyle.Render(strconv.Itoa(m.getReqCount()))
	reqTime := reqTimeStyle.Render(m.getReqTime())

	indicator := indicatorStyle.Render(
		getStatusIndicator(resStatusCode, m.req.Proto))

	statusVal := statusText.Copy().
		Width(screenWidth - w(status) - w(reqCounter) - w(reqTime) - w(indicator)).
		Render(m.getStatusText())

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		status,
		statusVal,
		reqCounter,
		reqTime,
		indicator,
	)

	return statusBarStyle.Width(screenWidth).Render(bar)
}

var usedScreenLines int

func (m model) View() string {
	var lines []string

	// print prompts
	var prompts []string
	for i := 0; i < len(m.inputs)-1; i += 2 {
		prompts = append(
			prompts,
			lipgloss.JoinHorizontal(lipgloss.Top, " ", m.inputs[i].View(), m.inputs[i+1].View()))
	}
	lines = append(lines, prompts...)
	lines = append(lines, "") // one more empty line between this and next render

	// print result URL
	reqUrl := urlStyle.Render(m.req.Proto + " " + m.req.Method + " " + m.req.URL.String())
	lines = append(lines, reqUrl)

	// print headers
	lines = append(lines, headersPrintf(m.req.Header)...)
	lines = append(lines, "") // one more empty line between this and next render

	// print body
	if m.req.Body != nil {
		lines = append(lines, bodyStyle.Render(formValues.Encode()))
		lines = append(lines, "") // one more empty line between this and next render
	}

	// print response
	if m.reqIsExecuted() {
		resUrl := urlStyle.Render(m.res.Proto + " " + m.res.Status)
		lines = append(lines, resUrl)
		lines = append(lines, "") // one more empty line between this and next render

		// print headers
		lines = append(lines, headersPrintf(m.res.Header)...)
		lines = append(lines, "") // one more empty line between this and next render

		// TODO..
		// if m.res.Header["Content-Type"] == "application/json" {
		// } else {
		// 	b.WriteString("\n" + string(m.resBody))
		// }

		// print body
		usedScreenLines = len(lines)
		lines = append(lines, m.getRespPageLines()...)
		lines = append(lines, "") // one more empty line between this and next render
	}

	// add status bar
	lines = append(lines, m.formatStatusBar())

	// write all lines to output
	return lipgloss.JoinVertical(lipgloss.Top, lines...)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
