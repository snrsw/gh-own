// Package issue provides functionality to handle GitHub issues owned by a user.
package issue

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/ui"
)

func (o *GroupedIssues) View() error {
	m := ui.NewModel([]ui.Tab{
		ui.NewTab(fmt.Sprintf("Created (%d)", o.Created.TotalCount), ui.CreateList(o.issueItems(o.Created))),
		ui.NewTab(fmt.Sprintf("Participated (%d)", o.Participated.TotalCount), ui.CreateList(o.issueItems(o.Participated))),
		ui.NewTab(fmt.Sprintf("Assigned (%d)", o.Assigned.TotalCount), ui.CreateList(o.issueItems(o.Assigned))),
	})

	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (i issue) toItem() ui.Item {
	var desc string
	if i.LatestActivity.Login != "" {
		desc = fmt.Sprintf(
			"#%d opened on %s by %s, %s by %s %s",
			i.Number,
			ui.CreatedOn(i.CreatedAt),
			i.User.Login,
			i.LatestActivity.Kind,
			i.LatestActivity.Login,
			ui.UpdatedAgo(i.LatestActivity.At),
		)
	} else {
		desc = fmt.Sprintf(
			"#%d opened on %s by %s, updated %s",
			i.Number,
			ui.CreatedOn(i.CreatedAt),
			i.User.Login,
			ui.UpdatedAgo(i.UpdatedAt),
		)
	}
	return ui.NewItem(
		i.repositoryFullName(),
		i.Title,
		desc,
		i.HTMLURL,
	)
}

func (o *GroupedIssues) issueItems(issues gh.SearchResult[issue]) []list.Item {
	items := make([]list.Item, 0, len(issues.Items))
	for _, issue := range issues.Items {
		items = append(items, issue.toItem())
	}
	return items
}
