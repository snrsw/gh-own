// Package issue provides functionality to handle GitHub issues owned by a user.
package issue

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/ui"
)

func (o *groupedIssues) View() error {
	m := ui.NewModel([]ui.Tab{
		ui.NewTab(fmt.Sprintf("Created (%d)", o.Created.TotalCount), ui.CreateList(o.issueItems(o.Created))),
		ui.NewTab(fmt.Sprintf("Participated (%d)", o.Participated.TotalCount), ui.CreateList(o.issueItems(o.Participated))),
		ui.NewTab(fmt.Sprintf("Assigned (%d)", o.Assigned.TotalCount), ui.CreateList(o.issueItems(o.Assigned))),
	})

	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (i issue) toItem() ui.Item {
	return ui.NewItem(
		i.repositoryFullName(),
		i.Title,
		fmt.Sprintf("#%d opened on %s by %s, updated %s", i.Number, ui.CreatedOn(i.CreatedAt), i.User.Login, ui.UpdatedAgo(i.UpdatedAt)),
		i.HTMLURL,
	)
}

func (o *groupedIssues) issueItems(issues gh.SearchResult[issue]) []list.Item {
	items := make([]list.Item, 0, len(issues.Items))
	for _, issue := range issues.Items {
		items = append(items, issue.toItem())
	}
	return items
}
