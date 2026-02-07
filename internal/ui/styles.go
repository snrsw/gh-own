// Package ui provides the user interface components for the application.
package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	colorAccent    = lipgloss.AdaptiveColor{Light: "#0969DA", Dark: "#2F81F7"} // GitHub blue
	colorSecondary = lipgloss.AdaptiveColor{Light: "#57606A", Dark: "#8B949E"} // GitHub gray
	colorMuted     = lipgloss.AdaptiveColor{Light: "#6E7781", Dark: "#6E7681"} // GitHub muted
	colorRepoName  = lipgloss.AdaptiveColor{Light: "#656D76", Dark: "#848D97"} // light gray
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
	l.KeyMap.CursorUp.SetEnabled(false)
	l.KeyMap.CursorDown.SetEnabled(false)
	l.KeyMap.Quit.SetEnabled(false)
	l.KeyMap.ShowFullHelp.SetEnabled(false)
	l.KeyMap.CloseFullHelp.SetEnabled(false)

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch tabs")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
			key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		}
	}
}

var (
	DocStyle    = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	WindowStyle = lipgloss.NewStyle().Align(lipgloss.Left)
)
