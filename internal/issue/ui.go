// Package issue provides functionality to handle GitHub issues owned by a user.
package issue

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/snrsw/gh-own/internal/ui"
)

func (o *GroupedIssues) View() error {
	m := ui.NewModel([]ui.Tab{
		ui.NewTab("Created", ui.CreateList(o.issueItems(o.Created))),
		ui.NewTab("Participated", ui.CreateList(o.issueItems(o.Participated))),
		ui.NewTab("Assigned", ui.CreateList(o.issueItems(o.Assigned))),
	})

	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (i Issue) toItem() ui.Item {
	return ui.NewItem(
		fmt.Sprintf("%s %s", i.RepositoryFullName(), i.Title),
		fmt.Sprintf("#%d opened on %s by %s, updated %s", i.Number, ui.CreatedOn(i.CreatedAt), i.User.Login, ui.UpdatedAgo(i.UpdatedAt)),
		i.HTMLURL,
	)
}

func (o *GroupedIssues) issueItems(issues []Issue) []list.Item {
	items := make([]list.Item, 0, len(issues))
	for _, issue := range issues {
		items = append(items, issue.toItem())
	}
	return items
}
