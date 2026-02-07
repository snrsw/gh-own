package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type githubDelegate struct {
	list.DefaultDelegate
	repoNameStyle lipgloss.Style
}

func (d githubDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Item)
	if !ok {
		return
	}
	if m.Width() <= 0 {
		return
	}

	title := item.Title()
	desc := item.Description()
	title = ansi.Truncate(title, m.Width(), "…")
	desc = ansi.Truncate(desc, m.Width(), "…")

	isSelected := index == m.Index() && m.FilterState() != list.Filtering

	if isSelected {
		title = d.Styles.SelectedTitle.Render(title)
		desc = d.Styles.SelectedDesc.Render(desc)
	} else {
		if item.repoName != "" && len(title) > len(item.repoName) {
			title = d.repoNameStyle.Render(title[:len(item.repoName)]) + d.Styles.NormalTitle.Render(title[len(item.repoName):])
		} else {
			title = d.Styles.NormalTitle.Render(title)
		}
		desc = d.Styles.NormalDesc.Render(desc)
	}

	if d.ShowDescription {
		fmt.Fprintf(w, "%s\n%s", title, desc)
	} else {
		fmt.Fprintf(w, "%s", title)
	}
}

func newGithubDelegate() githubDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(colorSecondary)

	d.Styles.NormalTitle = lipgloss.NewStyle().Bold(true)

	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(colorMuted)

	return githubDelegate{
		DefaultDelegate: d,
		repoNameStyle: lipgloss.NewStyle().
			Foreground(colorRepoName),
	}
}
