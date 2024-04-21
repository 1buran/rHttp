package main

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"time"
)

const (
	textShiftLimit = 5

	statusInfo int = iota
	statusWarning
	statusError
)

var (
	statusProtoHttp2, statusProtoHttps, statusProtoInsecure, statusDefaultIndEmoji string

	statusBarStyle, statusNugget, statusBadge, statusBadgeError, statusBadgeOk, statusBadgeWarning,
	reqCountStyle, resTimeStyle, statusText, statusTextInfo, statusTextError,
	statusTextWarning, indicatorStyle lipgloss.Style
)

// A status bar state.
type StatusBar struct {
	status        int
	text          string
	textOffset    int
	screenWidth   int
	reqScheme     string
	reqCount      int
	resTime       time.Duration
	resStatusCode int
	resProto      string
	resProtoMajor int
}

type StatusBarTickMsg time.Time

func StatusBarDoTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return StatusBarTickMsg(t)
	})
}

func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	switch msg.(type) {
	case StatusBarTickMsg:
		s.applyTimeChanges()
		return s, StatusBarDoTick()
	}
	return s, nil
}

// Apply time changes.
func (s *StatusBar) applyTimeChanges() {
	s.textOffset += textShiftLimit
}

// Get part of creeping text.
func (s *StatusBar) getText(w int) string {
	tw := len(s.text)

	if tw < w {
		return s.text
	}

	if s.textOffset < tw && s.text[s.textOffset] > 127 { // unicode detected
		s.textOffset += 4 // do not break unicode sequence, skip it
	}

	if s.textOffset >= tw {
		s.textOffset = 0
		return s.text[:w]
	}

	tl := s.textOffset + w
	if tl >= tw {
		appendix := w - (tw - s.textOffset)
		if s.text[appendix] > 127 { // unicode detected
			appendix-- // do not break unicode sequence, skip it
		}
		return s.text[s.textOffset:] + " " + s.text[:appendix]
	}

	if s.text[tl-1] > 127 { // unicode detected
		tl-- // do not break unicode sequence, skip it
	}

	return s.text[s.textOffset:tl]
}

// Get status message.
func (s *StatusBar) getStatusText(w int) string {
	var style lipgloss.Style
	// apply style
	switch s.status {
	case statusInfo:
		style = statusTextInfo
	case statusWarning:
		style = statusTextWarning
	case statusError:
		style = statusTextError
	}
	return style.Render(s.getText(w))
}

// Set status.
func (s *StatusBar) setStatus(status int, text string) {
	s.textOffset = 0
	s.status = status
	s.text = time.Now().Format("[15:04:05] ") + text
}

// Info message.
func (s *StatusBar) Info(text string) {
	s.setStatus(statusInfo, text)
}

// Warning message.
func (s *StatusBar) Warning(text string) {
	s.setStatus(statusWarning, text)
}

// Error message.
func (s *StatusBar) Error(text string) {
	s.setStatus(statusError, text)
}

// Increment count of requests.
func (s *StatusBar) IncReqCount() {
	s.reqCount++
}

// Get count of requests.
func (s *StatusBar) GetReqCount() int {
	return s.reqCount
}

// Get last response time.
func (s *StatusBar) GetResTime() string {
	return s.resTime.String()
}

// Set last response time.
func (s *StatusBar) SetResTime(t time.Duration) {
	s.resTime = t
}

// Set last response status code.
func (s *StatusBar) SetResStatusCode(c int) {
	s.resStatusCode = c
}

// Set last response proto.
func (s *StatusBar) SetResProto(major int, proto, scheme string) {
	s.resProtoMajor = major
	s.resProto = proto
	s.reqScheme = scheme
}

// Set count of requests.
func (s *StatusBar) SetReqCount(c int) {
	s.reqCount = c
}

// Set screen width.
func (s *StatusBar) SetScreenWidth(w int) {
	s.screenWidth = w
}

// Get status badge.
func (s *StatusBar) getStatusBadge(text string) (badge string) {
	var style lipgloss.Style

	switch s.status {
	case statusInfo:
		style = statusBadgeOk
	case statusWarning:
		style = statusBadgeWarning
	case statusError:
		style = statusBadgeError
	default:
		style = statusBadge
	}

	return style.Render(text)
}

// Get status logo indicator.
func (s *StatusBar) protoIndicator() (ind string) {
	switch s.resProtoMajor {
	case 1:
		if s.reqScheme == "https" {
			ind = statusProtoHttps + s.resProto
		} else {
			ind = statusProtoInsecure + s.resProto
		}
	case 2:
		ind = statusProtoHttp2 + s.resProto
	default:
		ind = statusDefaultIndEmoji + "rHttp"
	}
	return
}

// Format status bar.
func (s *StatusBar) FormatStatusBar() string {
	w := lipgloss.Width

	status := s.getStatusBadge("STATUS")
	reqCounter := reqCountStyle.Render(strconv.Itoa(s.reqCount))
	resTime := resTimeStyle.Render(s.GetResTime())
	proto := indicatorStyle.Render(s.protoIndicator())

	maxTextWidth := screenWidth - w(status) - w(reqCounter) - w(resTime) - w(proto)
	statusVal := statusText.Copy().Width(maxTextWidth).Render(s.getStatusText(maxTextWidth))
	bar := lipgloss.JoinHorizontal(lipgloss.Top, status, statusVal, reqCounter, resTime, proto)

	return statusBarStyle.Width(screenWidth).Render(bar)
}

func NewStatusBar(conf *Config) StatusBar {
	statusProtoHttp2 = conf.Emoji("statusbarProtoHttp2")
	statusProtoHttps = conf.Emoji("statusbarProtoHttps")
	statusDefaultIndEmoji = conf.Emoji("statusbarDefaultIndicator")
	statusProtoInsecure = conf.Emoji("statusbarProtoInsecure")

	statusBarStyle = lipgloss.NewStyle().
		Foreground(conf.Color("statusbarFg")).
		Background(conf.Color("statusbarBg"))
	statusNugget = lipgloss.NewStyle().Foreground(conf.Color("statusbarNugget")).Padding(0, 1)
	statusBadge = lipgloss.NewStyle().Inherit(statusBarStyle).
		Background(conf.Color("statusbarBadgeBg")).
		Foreground(conf.Color("statusbarBadgeFg")).
		Padding(0, 1).MarginRight(1)

	statusBadgeError = lipgloss.NewStyle().Inherit(statusBadge).
		Background(conf.Color("statusbarBadgeError")).Padding(0, 1)
	statusBadgeOk = lipgloss.NewStyle().Inherit(statusBadge).
		Background(conf.Color("statusbarBadgeOk")).Padding(0, 1)
	statusBadgeWarning = lipgloss.NewStyle().Inherit(statusBadge).
		Background(conf.Color("statusbarBadgeWarning")).Padding(0, 1)

	reqCountStyle = statusNugget.Copy().
		Background(conf.Color("statusbarReqCount")).Align(lipgloss.Right)
	resTimeStyle = statusNugget.Copy().
		Background(conf.Color("statusbarResTime")).Align(lipgloss.Right)

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)
	statusTextInfo = lipgloss.NewStyle().Inherit(statusText)
	statusTextError = lipgloss.NewStyle().Inherit(statusText).
		Foreground(conf.Color("statusbarTextError"))

	statusTextWarning = lipgloss.NewStyle().Inherit(statusText).
		Foreground(conf.Color("statusbarTextWarning"))
	indicatorStyle = statusNugget.Copy().Background(conf.Color("statusbarIndicator"))

	return StatusBar{}
}
