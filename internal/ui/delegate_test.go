package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestRender_ThreeLines(t *testing.T) {
	d := newGithubDelegate()
	d.ShowDescription = true
	item := NewItem("owner/repo", "Fix bug #42 ✓", "updated 2h ago", "https://example.com")
	m := list.New([]list.Item{item}, d, 80, 20)

	var buf bytes.Buffer
	d.Render(&buf, m, 0, item)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("Render() produced %d lines, want 3:\n%q", len(lines), buf.String())
	}
}

func renderLines(t *testing.T, d githubDelegate, m list.Model, index int, item Item) []string {
	t.Helper()
	var buf bytes.Buffer
	d.Render(&buf, m, index, item)
	return strings.Split(buf.String(), "\n")
}

func TestRender_SelectedTitleUsesStrippedContent(t *testing.T) {
	// Force ANSI colors so styles produce distinct output outside a TTY.
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	d := newGithubDelegate()
	d.ShowDescription = true

	// Simulate a PR: pre-styled #N in titleText, pre-styled status marks in titleSuffix.
	preTitledNumber := lipgloss.NewStyle().Foreground(lipgloss.Color("#0969DA")).Render("#7")
	styledTitle := preTitledNumber + " chore: update CI configuration"
	styledSuffix := lipgloss.NewStyle().Foreground(lipgloss.Color("#1a7f37")).Render("✓")
	item := NewItem("owner/repo", styledTitle, "updated 2h ago", "https://example.com").
		WithSuffix(" " + styledSuffix)
	m := list.New([]list.Item{item, item}, d, 80, 20)

	sel := renderLines(t, d, m, 0, item) // selected (index 0 == m.Index())

	// When selected, the title line must equal SelectedTitle applied to the stripped
	// titleText, with the suffix (status mark) appended as-is (preserving its color).
	want := d.Styles.SelectedTitle.Render(ansi.Strip(styledTitle)) + " " + styledSuffix
	if sel[1] != want {
		t.Errorf("selected title line\n got  %q\n want %q", sel[1], want)
	}
}

func TestRender_SelectedLinesChanges(t *testing.T) {
	// Force ANSI colors so styles produce distinct output outside a TTY.
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	d := newGithubDelegate()
	d.ShowDescription = true
	item := NewItem("owner/repo", "Fix bug", "updated 2h ago", "https://example.com")
	m := list.New([]list.Item{item, item}, d, 80, 20)

	sel := renderLines(t, d, m, 0, item)   // selected
	unsel := renderLines(t, d, m, 1, item) // unselected

	if sel[0] == unsel[0] {
		t.Error("repo line should render differently when item is selected")
	}
	if sel[1] == unsel[1] {
		t.Error("title line should render differently when item is selected")
	}
}

func TestRender_TwoLines_NoDescription(t *testing.T) {
	d := newGithubDelegate()
	d.ShowDescription = false
	item := NewItem("owner/repo", "Fix bug #42 ✓", "updated 2h ago", "https://example.com")
	m := list.New([]list.Item{item}, d, 80, 20)

	var buf bytes.Buffer
	d.Render(&buf, m, 0, item)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("Render() produced %d lines, want 2:\n%q", len(lines), buf.String())
	}
}
