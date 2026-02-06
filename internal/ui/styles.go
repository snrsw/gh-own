// Package ui provides the user interface components for the application.
package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func githubTabColors() (active, inactive lipgloss.AdaptiveColor) {
	return lipgloss.AdaptiveColor{
			Light: "#0969DA", // GitHub blue
			Dark:  "#2F81F7",
		}, lipgloss.AdaptiveColor{
			Light: "#57606A", // GitHub gray
			Dark:  "#8B949E",
		}
}

func GithubTabStyles() (active, inactive lipgloss.Style) {
	activeColor, inactiveColor := githubTabColors()

	active = lipgloss.NewStyle().
		Foreground(activeColor).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(activeColor)

	inactive = lipgloss.NewStyle().
		Foreground(inactiveColor).
		Padding(0, 1)

	return
}

func githubAccentColor() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: "#0969DA", // GitHub blue (light theme)
		Dark:  "#2F81F7", // GitHub blue (dark theme)
	}
}

func GithubDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	accent := githubAccentColor()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(accent).
		Bold(true)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#57606A", Dark: "#8B949E"})

	d.Styles.NormalTitle = lipgloss.NewStyle()

	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#6E7781", Dark: "#6E7681"})

	return d
}

var (
	DocStyle    = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	WindowStyle = lipgloss.NewStyle().Align(lipgloss.Left)
)
