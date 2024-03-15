package main

import (
	"github.com/charmbracelet/lipgloss"

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
				Background(lipgloss.Color("#E3BC68")).Padding(0, 1)

	encodingStyle = statusNugget.Copy().
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	statusText        = lipgloss.NewStyle().Inherit(statusBarStyle)
	statusTextInfo    = lipgloss.NewStyle().Inherit(statusText)
	statusTextError   = lipgloss.NewStyle().Inherit(statusText).Foreground(lipgloss.Color("225"))
	statusTextWarning = lipgloss.NewStyle().Inherit(statusText).Foreground(yellow)
	indicatorStyle    = statusNugget.Copy().Background(lipgloss.Color("#6124DF"))
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

// Set status.
func (s *StatusBar) setStatus(status int, text string) {
	s.status = status
	s.text = time.Now().Format("[15:04:05] ") + text
}

// Increment count of request.
func (s *StatusBar) incReqCount() {
	s.reqCount++
}

// Get count of executed requests.
func (s *StatusBar) getReqCount() int {
	return s.reqCount
}

// Get status logo indicator.
func getStatusIndicator(resStatusCode int, proto string) (ind string) {
	switch {
	case resStatusCode >= 400:
		ind = statusErrorEmoji + " " + proto
	case resStatusCode >= 300:
		ind = statusWarningEmoji + " " + proto
	case resStatusCode >= 100:
		ind = statusInfoEmoji + " " + proto
	case resStatusCode > 0, resStatusCode < 0:
		ind = `🤔` + " " + proto
	default:
		ind = `💜 rHttp`
	}
	return
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
