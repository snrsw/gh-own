// Package pr provides functionality to handle GitHub pull requests owned by a user.
package pr

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/snrsw/gh-own/internal/cistatus"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/ui"
)

func (o *GroupedPullRequests) View() error {
	m := ui.NewModel([]ui.Tab{
		ui.NewTab(fmt.Sprintf("Created (%d)", o.Created.TotalCount), ui.CreateList(o.prItems(o.Created))),
		ui.NewTab(fmt.Sprintf("Participated (%d)", o.Participated.TotalCount), ui.CreateList(o.prItems(o.Participated))),
		ui.NewTab(fmt.Sprintf("Assigned (%d)", o.Assigned.TotalCount), ui.CreateList(o.prItems(o.Assigned))),
		ui.NewTab(fmt.Sprintf("Review Requested (%d)", o.ReviewRequested.TotalCount), ui.CreateList(o.prItems(o.ReviewRequested))),
	})

	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (p PullRequest) toItem() ui.Item {
	return ui.NewItem(
		p.RepositoryFullName(),
		fmt.Sprintf("%s %s", p.Title, cistatus.RenderCIStatus(p.CIStatus)),
		fmt.Sprintf(
			"%s opened on %s by %s, updated %s",
			RenderPRNumber(p.Number, p.Draft),
			ui.CreatedOn(p.CreatedAt),
			p.User.Login,
			ui.UpdatedAgo(p.UpdatedAt),
		),
		p.HTMLURL,
	)
}

func (o *GroupedPullRequests) prItems(prs gh.SearchResult[PullRequest]) []list.Item {
	items := make([]list.Item, 0, len(prs.Items))
	for _, pr := range prs.Items {
		items = append(items, pr.toItem())
	}
	return items
}

func RenderPRNumber(n int, draft bool) string {
	s := fmt.Sprintf("#%d", n)
	if draft {
		return numberDraftStyle.Render(s)
	}
	return numberStyle.Render(s)
}

var (
	numberStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#0969DA")) // GitHub blue
	numberDraftStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6E7781")) // GitHub gray
)
