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

	repo  := ansi.Truncate(item.repoName,    m.Width(), "…")
	title := ansi.Truncate(item.titleText,   m.Width(), "…")
	desc  := ansi.Truncate(item.description, m.Width(), "…")

	isSelected := index == m.Index() && m.FilterState() != list.Filtering

	if isSelected {
		repo  = d.Styles.SelectedTitle.Render(repo)
		title = d.Styles.SelectedTitle.Render(ansi.Strip(title)) + item.titleSuffix
		desc  = d.Styles.SelectedDesc.Render(desc)
	} else {
		repo  = d.repoNameStyle.Render(repo)
		title = d.Styles.NormalTitle.Render(title) + item.titleSuffix
		desc  = d.Styles.NormalDesc.Render(desc)
	}

	if d.ShowDescription {
		fmt.Fprintf(w, "%s\n%s\n%s", repo, title, desc)
	} else {
		fmt.Fprintf(w, "%s\n%s", repo, title)
	}
}

func newGithubDelegate() githubDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(colorSecondary)

	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(colorTitle)

	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(colorMuted)

	d.SetHeight(3)

	return githubDelegate{
		DefaultDelegate: d,
		repoNameStyle: lipgloss.NewStyle().
			Foreground(colorSecondary),
	}
}
