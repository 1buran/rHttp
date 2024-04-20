package main

import (
	_ "embed"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

//go:embed config.json
var defaultConfig []byte

// Config of rHtttp.
type Config struct {
	Settings `json:"Settings"`
	Theme    `json:"Theme"`
	Warnings []string // todo show warnings to user
}

// Settings: default checkbox state, full screen mode etc.
type Settings struct {
	Timeout      int             `json:"Timeout"`
	MaxRedirects int             `json:"MaxRedirects"`
	Fullscreen   bool            `json:"Fullscreen"`
	Checkboxes   map[string]bool `json:"Checkboxes"`
}

// UI color settings.
type Theme struct {
	Colors map[string]lipgloss.Color `json:"Colors"`
	Emojis map[string]string         `json:"Emojis"`
}

func (c *Config) HasWarnings() bool {
	return len(c.Warnings) > 0
}

func (c *Config) WarningMessage() string {
	return strings.Join(c.Warnings, ", ")
}

// Add error.
func (c *Config) AddWarn(warn string) {
	c.Warnings = append(c.Warnings, warn)
}

// Color of item, write to stderr a message if requested color is not found in theme.
func (c *Config) Color(s string) lipgloss.Color {
	color, ok := c.Colors[s]
	if !ok {
		fallbackColor := lipgloss.Color("11")
		c.AddWarn(`color "` + s + `" not found`)
		return fallbackColor
	}
	return color
}

// Emoji of item, write to stderr a message if requested emoji is not found in theme.
func (c *Config) Emoji(s string) string {
	emoji, ok := c.Emojis[s]
	if !ok {
		fallbackEmoji := "❓"
		c.AddWarn(`emoji "` + s + `" not found`)
		return fallbackEmoji
	}
	return emoji
}

// Read config.
func ReadConfig(filepath string) (*Config, error) {
	c := &Config{}
	err := json.Unmarshal(defaultConfig, c)
	if err != nil {
		return nil, err
	}

	if filepath == "" {
		return c, nil // no filepath passed, return default config
	}

	r, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
