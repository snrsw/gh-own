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
		ui.NewTab(fmt.Sprintf("Needs action (%d)", o.NeedsAction.TotalCount), ui.CreateList(o.prItems(o.NeedsAction))),
		ui.NewTab(fmt.Sprintf("Ready to merge (%d)", o.ReadyToMerge.TotalCount), ui.CreateList(o.prItems(o.ReadyToMerge))),
		ui.NewTab(fmt.Sprintf("Review Requested (%d)", o.ReviewRequested.TotalCount), ui.CreateList(o.prItems(o.ReviewRequested))),
		ui.NewTab(fmt.Sprintf("Waiting for review or checks (%d)", o.Waiting.TotalCount), ui.CreateList(o.prItems(o.Waiting))),
		ui.NewTab(fmt.Sprintf("Drafts (%d)", o.Drafts.TotalCount), ui.CreateList(o.prItems(o.Drafts))),
		ui.NewTab(fmt.Sprintf("Participated (%d)", o.Participated.TotalCount), ui.CreateList(o.prItems(o.Participated))),
	}

	keys := make([]string, 0, len(o.Custom))
	for k := range o.Custom {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sr := o.Custom[k]
		tabs = append(tabs, ui.NewTab(fmt.Sprintf("%s (%d)", ui.HumanizeTabName(k), sr.TotalCount), ui.CreateList(o.prItems(sr))))
	}

	return tabs
}

func (p pullRequest) toItem(currentLogin string) ui.Item {
	var desc string
	if activity := p.shownActivity(); activity.Login != "" {
		desc = fmt.Sprintf(
			"%s by %s %s · opened on %s by %s",
			ui.RenderActivityKind(activity.Kind),
			ui.RenderUser(activity.Login, currentLogin),
			ui.UpdatedAgo(activity.At),
			ui.CreatedOn(p.CreatedAt),
			ui.RenderUser(p.User.Login, currentLogin),
		)
	} else {
		desc = fmt.Sprintf(
			"updated %s · opened on %s by %s",
			ui.UpdatedAgo(p.UpdatedAt),
			ui.CreatedOn(p.CreatedAt),
			ui.RenderUser(p.User.Login, currentLogin),
		)
	}
	titleText := RenderPRNumber(p.Number, p.Draft) + " " + p.Title

	suffix := " " + cistatus.RenderCIStatus(p.CIStatus)
	if rs := reviewstatus.RenderReviewStatus(p.ReviewStatus); rs != "" {
		suffix = " " + rs + suffix
	}

	return ui.NewItem(
		p.repositoryFullName(),
		titleText,
		desc,
		p.HTMLURL,
	).WithSuffix(suffix).WithSortAt(p.SortAt)
}

func (o *GroupedPullRequests) prItems(prs gh.SearchResult[pullRequest]) []list.Item {
	items := make([]list.Item, 0, len(prs.Items))
	for _, pr := range prs.Items {
		items = append(items, pr.toItem(o.currentLogin))
	}
	return items
}

func RenderPRNumber(n int, draft bool) string {
	s := fmt.Sprintf("#%d", n)
	if draft {
		return numberDraftStyle.Render("[DRAFT] " + s)
	}
	return numberStyle.Render(s)
}

var (
	numberStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#0969DA")) // GitHub blue
	numberDraftStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6E7781")) // GitHub gray
)
