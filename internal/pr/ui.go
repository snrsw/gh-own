// Package pr provides functionality to handle GitHub pull requests owned by a user.
package pr

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/snrsw/gh-own/internal/cistatus"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/reviewstatus"
	"github.com/snrsw/gh-own/internal/ui"
)

// BuildTabs converts grouped pull requests into UI tabs.
func (o *GroupedPullRequests) BuildTabs() []ui.Tab {
	tabs := []ui.Tab{
		ui.NewTab(fmt.Sprintf("Created (%d)", o.Created.TotalCount), ui.CreateList(o.prItems(o.Created))),
		ui.NewTab(fmt.Sprintf("Participated (%d)", o.Participated.TotalCount), ui.CreateList(o.prItems(o.Participated))),
		ui.NewTab(fmt.Sprintf("Assigned (%d)", o.Assigned.TotalCount), ui.CreateList(o.prItems(o.Assigned))),
		ui.NewTab(fmt.Sprintf("Review Requested (%d)", o.ReviewRequested.TotalCount), ui.CreateList(o.prItems(o.ReviewRequested))),
	}

	keys := make([]string, 0, len(o.Custom))
	for k := range o.Custom {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sr := o.Custom[k]
		tabs = append(tabs, ui.NewTab(fmt.Sprintf("%s (%d)", k, sr.TotalCount), ui.CreateList(o.prItems(sr))))
	}

	return tabs
}

func (p pullRequest) toItem() ui.Item {
	var desc string
	if p.LatestActivity.Login != "" {
		desc = fmt.Sprintf(
			"%s opened on %s by %s, %s by %s %s",
			RenderPRNumber(p.Number, p.Draft),
			ui.CreatedOn(p.CreatedAt),
			p.User.Login,
			p.LatestActivity.Kind,
			p.LatestActivity.Login,
			ui.UpdatedAgo(p.LatestActivity.At),
		)
	} else {
		desc = fmt.Sprintf(
			"%s opened on %s by %s, updated %s",
			RenderPRNumber(p.Number, p.Draft),
			ui.CreatedOn(p.CreatedAt),
			p.User.Login,
			ui.UpdatedAgo(p.UpdatedAt),
		)
	}
	titleText := p.Title
	if rs := reviewstatus.RenderReviewStatus(p.ReviewStatus); rs != "" {
		titleText += " " + rs
	}
	titleText += " " + cistatus.RenderCIStatus(p.CIStatus)

	return ui.NewItem(
		p.repositoryFullName(),
		titleText,
		desc,
		p.HTMLURL,
	)
}

func (o *GroupedPullRequests) prItems(prs gh.SearchResult[pullRequest]) []list.Item {
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
