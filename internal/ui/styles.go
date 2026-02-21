// Package ui provides the user interface components for the application.
package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	colorAccent    = lipgloss.AdaptiveColor{Light: "#0969DA", Dark: "#2F81F7"} // GitHub blue
	colorSecondary = lipgloss.AdaptiveColor{Light: "#57606A", Dark: "#8B949E"} // GitHub gray
	colorTitle     = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#D1D5DB"} // slightly light
	colorMuted     = lipgloss.AdaptiveColor{Light: "#6E7781", Dark: "#6E7681"} // GitHub muted
)

func GithubTabStyles() (active, inactive lipgloss.Style) {
	active = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(colorAccent)

	inactive = lipgloss.NewStyle().
		Foreground(colorSecondary).
		Padding(0, 1)

	return
}

func configureHelp(l *list.Model) {
	l.SetShowHelp(false)
}

var (
	helpKeyStyle  = lipgloss.NewStyle().Foreground(colorSecondary)
	helpDescStyle = lipgloss.NewStyle().Foreground(colorMuted)
	helpSepStyle  = lipgloss.NewStyle().Foreground(colorMuted)
)

func helpView() string {
	entries := []struct{ key, desc string }{
		{"/", "filter"},
		{"r", "refresh"},
		{"tab", "switch tabs"},
		{"enter", "open"},
		{"ctrl+c", "quit"},
	}

	var parts []string
	for _, e := range entries {
		parts = append(parts, helpKeyStyle.Render(e.key)+" "+helpDescStyle.Render(e.desc))
	}

	sep := helpSepStyle.Render(" • ")
	return strings.Join(parts, sep)
}

var (
	DocStyle    = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	WindowStyle = lipgloss.NewStyle().Align(lipgloss.Left)
)
