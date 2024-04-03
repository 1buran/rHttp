package main

import (
	"github.com/charmbracelet/lipgloss"
	"strconv"

	"time"
)

const (
	statusInfoEmoji    = "🟢"
	statusWarningEmoji = "🟡"
	statusErrorEmoji   = "🔴"

	yellow = lipgloss.Color("220")

	statusInfo int = iota
	statusWarning
	statusError
)

var (
	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusBadge = lipgloss.NewStyle().Inherit(statusBarStyle).
			Background(lipgloss.Color("#59A8C9")).
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1).
			MarginRight(1)
	statusBadgeError = lipgloss.NewStyle().Inherit(statusBadge).
				Background(lipgloss.Color("#FF5F87")).Padding(0, 1)

	statusBadgeOk = lipgloss.NewStyle().Inherit(statusBadge).
			Background(lipgloss.Color("#2e8048")).Padding(0, 1)
	statusBadgeWarning = lipgloss.NewStyle().Inherit(statusBadge).
				Background(lipgloss.Color("130")).Padding(0, 1)

	reqCountStyle = statusNugget.Copy().
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	resTimeStyle = statusNugget.Copy().
			Background(lipgloss.Color("#C550DF")).
			Align(lipgloss.Right)

	statusText        = lipgloss.NewStyle().Inherit(statusBarStyle)
	statusTextInfo    = lipgloss.NewStyle().Inherit(statusText)
	statusTextError   = lipgloss.NewStyle().Inherit(statusText).Foreground(lipgloss.Color("225"))
	statusTextWarning = lipgloss.NewStyle().Inherit(statusText).Foreground(yellow)
	indicatorStyle    = statusNugget.Copy().Background(lipgloss.Color("#6124DF"))
)

// A status bar state.
type StatusBar struct {
	status        int
	text          string
	reqCount      int
	resTime       time.Duration
	resStatusCode int
	resProto      string
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

// Set status.
func (s *StatusBar) setStatus(status int, text string) {
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

// Set count of requests.
func (s *StatusBar) SetReqCount(c int) {
	s.reqCount = c
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
func (s *StatusBar) getStatusIndicator() (ind string) {
	switch {
	case s.resStatusCode >= 400:
		ind = statusErrorEmoji + " " + s.resProto
	case s.resStatusCode >= 300:
		ind = statusWarningEmoji + " " + s.resProto
	case s.resStatusCode >= 100:
		ind = statusInfoEmoji + " " + s.resProto
	case s.resStatusCode > 0, s.resStatusCode < 0:
		ind = `🤔` + " " + s.resProto
	default:
		ind = `💜 rHttp`
	}
	return
}

// Format status bar.
func (s *StatusBar) FormatStatusBar() string {
	w := lipgloss.Width

	status := s.getStatusBadge("STATUS")
	reqCounter := reqCountStyle.Render(strconv.Itoa(s.reqCount))
	resTime := resTimeStyle.Render(s.GetResTime())

	indicator := indicatorStyle.Render(s.getStatusIndicator())

	statusVal := statusText.Copy().
		Width(screenWidth - w(status) - w(reqCounter) - w(resTime) - w(indicator)).
		Render(s.getStatusText())

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		status,
		statusVal,
		reqCounter,
		resTime,
		indicator,
	)

	return statusBarStyle.Width(screenWidth).Render(bar)
}

func NewStatusBar() *StatusBar {
	return &StatusBar{}
}
