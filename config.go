package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
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
	Chroma string                    `json:"Chroma"`
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

// Load default settings.
func loadDefaultSettings(c *Config) error {
	return json.Unmarshal(defaultConfig, c)
}

// Load user settings.
// todo make it is more robust, do not silent errors from the one side and
// todo do not make a lot of noise / spam from the other...
func loadUserSettings(c *Config) error {
	userHome, _ := os.UserHomeDir() // ignore rare cases when user home is undefined
	userConfig := filepath.Join(userHome, ".config/rhttp/config.json")
	if _, err := os.Stat(userConfig); err != nil {
		if errors.Is(err, os.ErrNotExist) { // it's ok if config is missed
			return nil
		} else {
			return err
		}
	}

	return loadJSON(userConfig, c)
}
func loadOverrideSettings(path string, c *Config) error {
	if path == "" {
		return nil // no config path passed
	}

	if strings.HasPrefix(path, "~/") {
		userHome, _ := os.UserHomeDir() // ignore rare cases when user home is undefined
		path = filepath.Join(userHome, path)
	}
	return loadJSON(path, c)
}

func loadJSON(path string, c *Config) error {
	r, err := os.Open(path)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, c)
	if err != nil {
		return err
	}
	return nil
}

// Read config.
func ReadConfig(configPath string) (*Config, error) {
	c := &Config{}

	if err := loadDefaultSettings(c); err != nil {
		return nil, err
	}

	if err := loadUserSettings(c); err != nil {
		return nil, err
	}

	if err := loadOverrideSettings(configPath, c); err != nil {
		return nil, err
	}

	return c, nil
}

// Print default config.
func PrintDefaultConfig() {
	os.Stdout.Write(defaultConfig)
}
