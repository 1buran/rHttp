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
		ind = statusErrorEmoji
	case resStatusCode >= 300:
		ind = statusWarningEmoji
	case resStatusCode >= 100:
		ind = statusInfoEmoji
	case resStatusCode > 0, resStatusCode < 0:
		ind = `🤔`
	default:
		ind = `💜 rHttp`
	}
	ind += " " + proto
	return
}
