package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Number of input or ENUM(input type, int)
const (
	// Bubbletea inputs: instances of [textinput.Model],
	// in code m.inputs slice contains references to them.
	method = iota
	host
	urlPath

	// Key value pairs.
	header
	headerVal

	param
	paramVal

	cookie
	cookieVal

	form
	formVal

	// The last one is the max index of defined text input,
	// this is abroad between text inputs and checkboxes.
	fieldsCount

	// Custom inputs: instances of [Checkbox].
	// in code m.checkboxes slice contains references to them.
	// Index of checkbox can be calculated in this way:
	//   m.checkboxes[i - fieldsCount - 1]
	https
	autoformat

	// last index
	end
)

// File inputs.
const (
	sessionSave = end + iota + 1
	sessionLoad
)

const (
	hotPink   = lipgloss.Color("69")
	lightPink = lipgloss.Color("225")
	darkGray  = lipgloss.Color("243")
	lightGray = lipgloss.Color("249")

	purple        = lipgloss.Color("141")
	brightPurple  = lipgloss.Color("183")
	brightPurple2 = lipgloss.Color("189")
	lightBlue     = lipgloss.Color("12")
	rose          = lipgloss.Color("177")
	rose2         = lipgloss.Color("219")
)

var (
	screenWidth  = 100
	screenHeight = 50
	offsetShift  = 5
	baseStyle    = lipgloss.NewStyle().Width(screenWidth)

	checkboxOnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).PaddingRight(21)
	checkboxOffStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).PaddingRight(21)

	promptStyle       = lipgloss.NewStyle().Foreground(hotPink).Bold(true)
	promptActiveStyle = lipgloss.NewStyle().Foreground(rose).Bold(true)
	textStyle         = lipgloss.NewStyle().Foreground(purple)
	textValueStyle    = lipgloss.NewStyle().Foreground(brightPurple)

	placeholderStyle       = lipgloss.NewStyle().Foreground(darkGray)
	placeholderActiveStyle = lipgloss.NewStyle().Foreground(lightGray)

	uriStyle         = lipgloss.NewStyle().Foreground(hotPink)
	headerStyle      = textStyle
	headerValueStyle = lipgloss.NewStyle().Foreground(brightPurple)
	urlStyle         = lipgloss.NewStyle().Inherit(baseStyle).
				Foreground(brightPurple2).
				Bold(true).Padding(0, 1)

	bodyStyle = lipgloss.NewStyle().Inherit(baseStyle).Foreground(lightPink)

	pressedKeyStyle = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(rose2),
		lipgloss.NewStyle().Foreground(lightPink).Bold(true),
	}

	titleStyle = lipgloss.NewStyle().Foreground(lightBlue).
			Bold(true).BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1).Width(60).AlignHorizontal(lipgloss.Center)

	sbar = NewStatusBar()
)

func fileinputIndex(i int) (idx int) {
	idx = i - end - 1
	if idx < 0 {
		idx = 0 // first file input
	}
	return
}

func checkboxIndex(i int) (idx int) {
	idx = i - fieldsCount - 1
	if idx < 0 {
		idx = 0 // first checkbox element
	}
	return
}

func newReqest() (r *http.Request) {
	r, _ = http.NewRequest("GET", "http://localhost", nil)
	return
}

// Prepare request before send.
func prepareRequest(r *http.Request) {
	if len(formValues) > 0 {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Body = newReadCloser(formValues.Encode())
	}
}

var redirects []string

func handleRedirect(req *http.Request, via []*http.Request) error {
	redirects = append(redirects, strconv.Itoa(req.Response.StatusCode)+" "+req.URL.String())
	return nil
}

func sendRequest(r *http.Request) (*http.Response, error) {
	redirects = nil
	http_cli := http.Client{Timeout: 2 * time.Second, CheckRedirect: handleRedirect}
	prepareRequest(r)
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
	checkboxes   []Checkbox
	fileInputs   []FileInput
	cursorIdx    int    // edit type
	cursorKey    string // edit key of type orderedKeyVal store
	focused      int
	hideMenu     bool
	resBodyLines []string
	fullScreen   bool
	offset       int
	pressedKey   string
	KeyStroke
}

// Blur all prompts.
func (m *model) blurAllPrompts() {
	for i := range m.inputs {
		m.blurPrompt(i)
	}
	for i := range m.checkboxes {
		m.checkboxes[i].style[2] = promptStyle
	}
	for i := range m.fileInputs {
		m.fileInputs[i].Hide()
	}
}

// Blur prompt.
func (m *model) blurPrompt(i int) {
	p := i
	switch i {
	case headerVal, paramVal, cookieVal, formVal:
		p = i - 1
	}
	if i < fieldsCount {
		m.inputs[i].PlaceholderStyle = placeholderStyle
		m.inputs[p].PromptStyle = promptStyle
		m.inputs[i].Blur()
	} else {
		idx := checkboxIndex(i)
		m.checkboxes[idx].style[2] = promptStyle
	}
}

// Focus prompt.
func (m *model) focusPrompt(i int) {
	n := i
	switch i {
	case headerVal, paramVal, cookieVal, formVal:
		n = i - 1
	}
	if i < fieldsCount {
		m.inputs[i].PlaceholderStyle = placeholderActiveStyle
		m.inputs[n].PromptStyle = promptActiveStyle
		m.inputs[i].Focus()
		m.inputs[i].CursorEnd()
	} else {
		idx := checkboxIndex(i)
		m.checkboxes[idx].style[2] = promptActiveStyle
	}

}

// nextInput focuses the next input field
func (m *model) nextInput() {
	switch m.focused {
	case method:
		m.setReqMethod()
	}

	m.blurPrompt(m.focused)
	if m.focused+1 == fieldsCount {
		m.focused = fieldsCount + 1 // crossed abroad, go to checkboxes
	} else {
		m.focused = (m.focused + 1) % end
	}
	m.focusPrompt(m.focused)
}

// prevInput focuses the previous input field
func (m *model) prevInput() {

	m.blurPrompt(m.focused)

	if m.focused-1 == fieldsCount {
		m.focused -= 2 // crossed abroad, go to checkboxes
	} else {
		m.focused--
	}

	// Wrap around
	if m.focused < 0 {
		m.focused = end - 1
	}
	m.focusPrompt(m.focused)
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
func (m *model) getRespPageLines(usedLines int) []string {
	limit := screenHeight - usedLines // available screen lines for display of res body
	end := m.offset + limit
	if end > len(m.resBodyLines)-1 {
		return m.resBodyLines[m.offset:]
	}
	return m.resBodyLines[m.offset:end]
}

func (m *model) delReqHeader() {
	name := m.inputs[header].Value()
	if name != "" {
		m.req.Header.Del(name)
	}
}

func (m *model) setReqHeader() {
	var s1, s2 []string
	correctHeader(&m.inputs[header])
	name := m.inputs[header].Value()
	val := m.inputs[headerVal].Value()
	if name != "" && val != "" {
		m.req.Header.Set(name, val)
		for h, v := range m.req.Header {
			s1 = append(s1, h)
			s2 = append(s1, strings.Join(v, ";"))
		}
		m.inputs[header].SetSuggestions(s1)
		m.inputs[headerVal].SetSuggestions(s2)
		m.inputs[header].Reset()
		m.inputs[headerVal].Reset()
	}
}

func (m *model) delReqParam() {
	v, _ := url.ParseQuery(m.req.URL.RawQuery)
	name := m.inputs[param].Value()
	if name != "" {
		v.Del(name)
		m.req.URL.RawQuery = v.Encode()
	}
}

func (m *model) setReqParam() {
	var s1, s2 []string
	v, _ := url.ParseQuery(m.req.URL.RawQuery)
	name := m.inputs[param].Value()
	val := m.inputs[paramVal].Value()
	if name != "" && val != "" {
		v.Set(name, val)
		for pKey, pVal := range v {
			s1 = append(s1, pKey)
			s2 = append(s2, strings.Join(pVal, ", "))
		}
		m.inputs[param].SetSuggestions(s1)
		m.inputs[paramVal].SetSuggestions(s2)
		m.req.URL.RawQuery = v.Encode()
		m.inputs[param].Reset()
		m.inputs[paramVal].Reset()
	}
}

func (m *model) delReqCookie() {
	cookies := m.req.Cookies()
	m.req.Header.Del("Cookie")
	name := m.inputs[cookie].Value()
	if name != "" {
		for _, c := range cookies {
			if c.Name != name {
				m.req.AddCookie(c)
			}
		}
	}
}

func (m *model) setReqCookie() {
	var s1, s2 []string
	name := m.inputs[cookie].Value()
	val := m.inputs[cookieVal].Value()
	if name != "" && val != "" {
		c, _ := m.req.Cookie(name)
		if c == nil {
			m.req.AddCookie(&http.Cookie{Name: name, Value: val})
		}
		for _, i := range m.req.Cookies() {
			s1 = append(s1, i.Name)
			s2 = append(s2, i.Value)
		}
		m.inputs[cookie].SetSuggestions(s1)
		m.inputs[cookieVal].SetSuggestions(s2)
		m.inputs[cookie].Reset()
		m.inputs[cookieVal].Reset()
	}
}

func (m *model) delReqForm() {
	name := m.inputs[form].Value()
	if name != "" {
		formValues.Del(name)
	}
}

func (m *model) setReqForm() {
	var s1, s2 []string
	name := m.inputs[form].Value()
	val := m.inputs[formVal].Value()
	if name != "" && val != "" {
		if formValues.Get(name) != "" {
			formValues.Set(name, val)
		} else {
			formValues.Add(name, val)
		}
		for k, v := range formValues {
			s1 = append(s1, k)
			s2 = append(s2, strings.Join(v, ", "))
		}
		m.inputs[form].SetSuggestions(s1)
		m.inputs[formVal].SetSuggestions(s2)
		m.req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		m.inputs[form].Reset()
		m.inputs[formVal].Reset()
	}
}

func (m *model) setReqMethod() {
	defer func() {
		recover() // BUG: textinput.(*Model).CurrentSuggestion(...) cause panic first time
	}()
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

func (m *model) setReqUrlPath() {
	val := m.inputs[urlPath].Value()
	m.req.URL.Path = val
}

func (m *model) setReqHost() {
	val := m.inputs[host].Value()
	m.req.URL.Host = val
	m.req.Host = val
}

func (m *model) restoreReqMethod() {
	m.inputs[method].SetValue(m.req.Method)
	m.inputs[method].CursorEnd()
}

func (m *model) setHttps(b bool) {
	if b {
		m.req.URL.Scheme = "https"
	} else {
		m.req.URL.Scheme = "http"
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

var allowedMethods = []string{
	"GET", "POST", "PUT", "PATCH", "HEAD", "DELETE", "OPTIONS", "PROPFIND", "SEARCH",
	"TRACE", "PATCH", "PUT", "CONNECT",
}

var prompts = [fieldsCount]string{
	"Method ", "Host ", "Path   ",
	"Header ", "", "Param  ", "", "Cookie ", "", "Form   ", ""}
var placeholders = [fieldsCount]string{
	"GET", "example.com", "/",
	"X-Auth-Token", "token value", "products_id", "10",
	"XDEBUG_SESSION", "debugger", "login", "user"}

func NewKeyValInputs(n int) textinput.Model {
	t := textinput.New()
	t.Prompt = prompts[n]
	t.Placeholder = placeholders[n]
	t.Width = 25
	t.PromptStyle = promptStyle
	t.PlaceholderStyle = placeholderStyle
	t.TextStyle = textValueStyle
	t.ShowSuggestions = true

	// set defaults input text
	switch n {
	case host:
		t.SetValue("localhost")
	case method:
		t.SetValue("GET")
		t.PromptStyle = promptActiveStyle
		t.SetSuggestions(allowedMethods)
		t.Focus() // start program with first prompt activated
	case urlPath:
		t.Width = 52
	}
	return t
}

func initialModel() model {
	var inputs []textinput.Model
	var checkboxes []Checkbox
	var fileInputs []FileInput

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}
	screenWidth = w
	screenHeight = h

	for i := 0; i < fieldsCount; i++ {
		inputs = append(inputs, NewKeyValInputs(i))
	}

	c1 := NewCheckbox(https, "https  ", "⟨on⟩ ", "⟨off⟩", promptStyle, checkboxOnStyle, checkboxOffStyle)
	c2 := NewCheckbox(autoformat, "Auto format JSON ", "⟨on⟩ ", "⟨off⟩", promptStyle, checkboxOnStyle, checkboxOffStyle)
	checkboxes = append(checkboxes, c1, c2)

	f1 := NewFileInput(sessionSave, WriteMode, "Session save: ", "/home/user/ses.json")
	f2 := NewFileInput(sessionLoad, ReadMode, "Session load: ", "/home/user/ses.json")
	fileInputs = append(fileInputs, f1, f2)

	m := model{
		req:        newReqest(),
		inputs:     inputs,
		checkboxes: checkboxes,
		fileInputs: fileInputs,
		KeyStroke:  NewKeyStroke(),
	}
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, StatusBarDoTick())
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

func autoFormatJSON(s string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(s), "", "\t")
	return out.String()
}

func formatRespBody(ct, s string, autoformat bool) []string {
	var content strings.Builder

	if s == "" {
		return []string{}
	}

	lexer := matchContentTypeTolexer(ct)
	if lexer == nil {
		// detect lang
		lexer = lexers.Analyse(s)
	}
	lexer = chroma.Coalesce(lexer)

	if autoformat && lexer.Config().Name == "JSON" {
		s = autoFormatJSON(s)
	}

	// huge one line splitter
	lp := lipgloss.NewStyle().Width(screenWidth).Padding(0, 1)
	s = lp.Render(s)

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

// Forward message to [Checkbox] handler.
func (m *model) checkboxHandler(msg tea.Msg, i int) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	idx := checkboxIndex(i)
	m.checkboxes[idx], cmd = m.checkboxes[idx].Update(msg)
	return m, cmd
}

// Load session: create and populate request and response from the given file.
func loadSession(m model, r io.Reader) (tea.Model, tea.Cmd) {
	ses, _ := NewSession(
		m.req, m.res, sbar.GetReqCount(),
		sbar.GetResTime(), formValues, m.resBodyLines)
	err := ses.Load(r)
	if err != nil {
		sbar.Error(err.Error())
		return m, nil
	}
	// Update state (request and response and some other stuff)
	// TODO update suggestions and all this to separate function
	sbar.SetReqCount(ses.ReqCount)
	formValues = ses.Request.FormValues

	// Create a new request instance
	m.req = newReqest()

	// req scheme (http or https)
	m.req.URL.Scheme = ses.Request.Scheme
	idx := checkboxIndex(https)
	if m.req.URL.Scheme == "https" {
		m.checkboxes[idx].SetOn()
	} else {
		m.checkboxes[idx].SetOff()
	}

	// req method
	m.req.Method = ses.Request.Method
	m.inputs[method].SetValue(ses.Request.Method)

	// req host
	m.req.Host = ses.Request.Host
	m.req.URL.Host = ses.Request.Host
	m.inputs[host].SetValue(ses.Request.Host)

	// req url path
	m.req.URL.Path = ses.Request.UrlPath
	m.inputs[urlPath].SetValue(ses.Request.UrlPath)

	// req headers
	m.req.Header = ses.Request.Headers

	// req query params
	m.req.URL.RawQuery = ses.Request.RawQuery

	// Create a new response instance
	m.res = &http.Response{Request: m.req}
	m.res.Status = ses.Response.Status
	m.res.Proto = ses.Response.Proto
	m.res.Header = ses.Response.Headers
	m.resBodyLines = ses.Response.BodyLines

	return m, nil
}

var usedScreenLines int

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case FileInputReader:
		if msg.Error != nil {
			sbar.Error(msg.Error.Error())
			return m, nil
		}
		switch msg.Id {
		case sessionLoad:
			sbar.Info("load session from: " + msg.Path)
			m.focused = 0
			m.focusPrompt(0)
			return loadSession(m, msg.Reader)
		}
	case FileInputWriter:
		if msg.Error != nil {
			sbar.Error(msg.Error.Error())
			return m, nil
		}
		ses, _ := NewSession(
			m.req, m.res, sbar.GetReqCount(),
			sbar.GetResTime(), formValues, m.resBodyLines)
		err := ses.Save(msg.Writer)
		if err != nil {
			sbar.Error(err.Error())
		}
		sbar.Info("saved session to: " + msg.Path)
		m.focused = 0
		m.focusPrompt(0)
		return m, nil
	case CheckboxUpdated: // todo: move this to checkboxHandler (Checkbox.Update loop)
		switch msg.Id {
		case https:
			m.setHttps(msg.On)
		}
	case Timer:
		sbar.SetResTime(msg.elapsedTime())
		cmd := func() tea.Msg {
			return msg.payload
		}
		return m, cmd

	case *http.Response:
		defer msg.Body.Close()
		buf, _ := io.ReadAll(msg.Body)
		m.res = msg
		m.resBodyLines = formatRespBody(
			m.res.Header.Get("content-type"), string(buf),
			m.checkboxes[checkboxIndex(autoformat)].IsOn())
		sbar.SetResStatusCode(m.res.StatusCode)
		sbar.SetResProto(m.res.Proto)
		sbar.Info("request is executed, response taken")
		if len(redirects) > 0 {
			sbar.Warning(
				strconv.Itoa(len(redirects)) + " redirects: " + strings.Join(redirects, " → "))
		}

	case tea.WindowSizeMsg:
		sbar.Info(
			"detected screen size: " + strconv.Itoa(msg.Width) + " x " + strconv.Itoa(msg.Height))
		screenWidth = msg.Width
		screenHeight = msg.Height
		sbar.SetScreenWidth(msg.Width)
	case tea.KeyMsg:
		m.pressedKey = msg.String()
		switch {
		case key.Matches(msg, m.keys.PageDown):
			availableScreenLines := screenHeight - usedScreenLines
			if m.offset+offsetShift+availableScreenLines <= len(m.resBodyLines) {
				m.offset += offsetShift
			} else {
				// decrease offset to take one last page in full size of screen lines
				m.offset += len(m.resBodyLines) - availableScreenLines - m.offset
				if m.offset < 0 {
					m.offset = 0
				}
			}
		case key.Matches(msg, m.keys.PageUp):
			if m.offset-offsetShift >= 0 {
				m.offset -= offsetShift
			} else {
				m.offset = 0
			}
		case key.Matches(msg, m.keys.FullScreen):
			if m.fullScreen {
				m.fullScreen = false
				sbar.Info("full screen mode is off")
				return m, tea.ExitAltScreen
			}
			m.fullScreen = true
			sbar.Info("full screen mode is on")
			return m, tea.EnterAltScreen
		case key.Matches(msg, m.keys.Run):
			sbar.Info("sending request...")
			m.clearRespArtefacts()
			sbar.IncReqCount()
			cmd := func() tea.Msg {
				r, err := sendRequest(m.req)
				if err != nil {
					return NewMessageWithTimer(err)
				}
				return NewMessageWithTimer(r)
			}
			return m, cmd
		case key.Matches(msg, m.keys.Prev):
			m.prevInput()
			return m, nil
		case key.Matches(msg, m.keys.Next):
			m.nextInput()
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.SaveSession):
			idx := fileinputIndex(sessionSave)
			if !m.fileInputs[idx].visible {
				m.blurAllPrompts()
				m.fileInputs[idx].SetVisible()
				m.fileInputs[idx].Focus()
				m.focused = sessionSave
			} else {
				m.blurAllPrompts()
				m.fileInputs[idx].Hide()
				m.focused = 0
				m.focusPrompt(0)
			}
			// return m, nil
		case key.Matches(msg, m.keys.LoadSession):
			idx := fileinputIndex(sessionLoad)
			if !m.fileInputs[idx].visible {
				m.blurAllPrompts()
				m.fileInputs[idx].SetVisible()
				m.fileInputs[idx].Focus()
				m.focused = sessionLoad
			} else {
				m.blurAllPrompts()
				m.fileInputs[idx].Hide()
				m.focused = 0
				m.focusPrompt(0)
			}
			// return m, nil
		case key.Matches(msg, m.keys.Delete):
			switch m.focused {
			case header, headerVal:
				m.delReqHeader()
			case param, paramVal:
				m.delReqParam()
			case cookie, cookieVal:
				m.delReqCookie()
			case form, formVal:
				m.delReqForm()
			}
		case key.Matches(msg, m.keys.ToggleCheckbox):
			switch m.focused {
			case https:
				return m.checkboxHandler(msg, https)
			case autoformat:
				return m.checkboxHandler(msg, autoformat)
			}
		case key.Matches(msg, m.keys.Enter):
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
			case sessionSave:
				idx := fileinputIndex(sessionSave)
				return m, m.fileInputs[idx].OpenFile()
			case sessionLoad:
				idx := fileinputIndex(sessionLoad)
				return m, m.fileInputs[idx].OpenFile()
			}

			// after handling enter is done, go to next input..
			m.nextInput()
		}

	case error:
		sbar.Error(msg.Error())
		m.clearRespArtefacts()
		return m, tea.ClearScreen
	}

	// Update text inputs
	for i := 0; i < len(m.inputs); i++ {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	// Update file input widgets
	var c tea.Cmd
	for i := 0; i < len(m.fileInputs); i++ {
		m.fileInputs[i], c = m.fileInputs[i].Update(msg)
		cmds = append(cmds, c)
	}

	// Update status bar
	sbar, c = sbar.Update(msg)
	cmds = append(cmds, c)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// Layout parts
	var (
		prompts, reqHeaders, resHeaders, resBodyLines []string
		reqUrl, resUrl, formValuesEncoded             string
	)

	prompts = append(
		prompts,
		lipgloss.JoinHorizontal(lipgloss.Top, " ", m.inputs[method].View(), m.inputs[host].View()),
		" "+m.inputs[urlPath].View())

	// Text inputs (key value pairs)
	for i := header; i < fieldsCount; i += 2 {
		prompts = append(
			prompts,
			lipgloss.JoinHorizontal(lipgloss.Top, " ", m.inputs[i].View(), m.inputs[i+1].View()))
	}

	// Checkboxes
	prompts = append(
		prompts,
		lipgloss.JoinHorizontal(
			lipgloss.Top, " ",
			m.checkboxes[checkboxIndex(https)].View(),
			m.checkboxes[checkboxIndex(autoformat)].View(),
		),
	)

	// Request URL
	reqUrl = urlStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, m.req.Method, " ", m.req.URL.String()))

	// Request headers
	reqHeaders = headersPrintf(m.req.Header)

	// Request encoded form values
	if len(formValues) > 0 {
		formValuesEncoded = " " + bodyStyle.Render(formValues.Encode())
	}

	// print response
	if m.reqIsExecuted() {

		// Response URL
		resUrl = urlStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Top, m.res.Proto, " ", m.res.Status))

		// Response headers
		resHeaders = headersPrintf(m.res.Header)

		// TODO..
		// if m.res.Header["Content-Type"] == "application/json" {
		// } else {
		// 	b.WriteString("\n" + string(m.resBody))
		// }
	}

	// add status bar
	sbarRendered := sbar.FormatStatusBar()

	leftPanel := lipgloss.JoinVertical(lipgloss.Left, prompts...)
	lW, _ := lipgloss.Size(leftPanel)
	rW := screenWidth - lW
	if rW < 0 {
		rW = 0
	}
	m.help.Width = rW
	if m.pressedKey == " " {
		m.pressedKey = "space"
	}
	rpContent := []string{
		lipgloss.NewStyle().Width(rW).Render(m.help.View(m.keys)),
		lipgloss.NewStyle().Width(rW).Render(
			pressedKeyStyle[0].Render("Pressed key: ") +
				pressedKeyStyle[1].Render(m.pressedKey)),
	}
	for i := 0; i < len(m.fileInputs); i++ {
		if m.fileInputs[i].IsVisible() {
			rpContent = append(rpContent,
				lipgloss.NewStyle().Width(rW).Render(m.fileInputs[i].View()))
		}
	}
	rightPanel := lipgloss.JoinVertical(lipgloss.Center, rpContent...)

	menuRendered := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		rightPanel,
	)

	reqInfo := []string{"", reqUrl}
	reqInfo = append(reqInfo, reqHeaders...)
	if formValuesEncoded != "" {
		reqInfo = append(reqInfo, "", formValuesEncoded)
	}
	reqInfoRendered := lipgloss.JoinVertical(lipgloss.Top, reqInfo...)

	resInfo := []string{"", resUrl}
	resInfo = append(resInfo, resHeaders...)
	resInfo = append(resInfo, "")

	preResInfoRendered := lipgloss.JoinVertical(lipgloss.Top, resInfo...)

	usedScreenLines = lipgloss.Height(menuRendered) +
		lipgloss.Height(reqInfoRendered) +
		lipgloss.Height(preResInfoRendered) +
		1 // +1 line of status bar
	resBodyLines = m.getRespPageLines(usedScreenLines)
	resBodyRendered := lipgloss.JoinVertical(lipgloss.Top, resBodyLines...)
	resInfoRendered := lipgloss.JoinVertical(lipgloss.Top, preResInfoRendered, resBodyRendered)

	// write all lines to output
	return lipgloss.JoinVertical(
		lipgloss.Top, menuRendered, reqInfoRendered, resInfoRendered, sbarRendered,
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
