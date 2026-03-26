// Package issue provides functionality to handle GitHub issues owned by a user.
package issue

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/ui"
)

// BuildTabs converts grouped issues into UI tabs.
func (o *GroupedIssues) BuildTabs() []ui.Tab {
	tabs := []ui.Tab{
		ui.NewTab(fmt.Sprintf("Created (%d)", o.Created.TotalCount), ui.CreateList(o.issueItems(o.Created))),
		ui.NewTab(fmt.Sprintf("Participated (%d)", o.Participated.TotalCount), ui.CreateList(o.issueItems(o.Participated))),
		ui.NewTab(fmt.Sprintf("Assigned (%d)", o.Assigned.TotalCount), ui.CreateList(o.issueItems(o.Assigned))),
	}

	keys := make([]string, 0, len(o.Custom))
	for k := range o.Custom {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sr := o.Custom[k]
		tabs = append(tabs, ui.NewTab(fmt.Sprintf("%s (%d)", ui.HumanizeTabName(k), sr.TotalCount), ui.CreateList(o.issueItems(sr))))
	}

	return tabs
}

func (i issue) toItem(currentLogin string) ui.Item {
	var desc string
	if i.LatestActivity.Login != "" {
		desc = fmt.Sprintf(
			"opened on %s by %s, %s by %s %s",
			ui.CreatedOn(i.CreatedAt),
			ui.RenderUser(i.User.Login, currentLogin),
			i.LatestActivity.Kind,
			ui.RenderUser(i.LatestActivity.Login, currentLogin),
			ui.UpdatedAgo(i.LatestActivity.At),
		)
	} else {
		desc = fmt.Sprintf(
			"opened on %s by %s, updated %s",
			ui.CreatedOn(i.CreatedAt),
			ui.RenderUser(i.User.Login, currentLogin),
			ui.UpdatedAgo(i.UpdatedAt),
		)
	}
	return ui.NewItem(
		i.repositoryFullName(),
		fmt.Sprintf("#%d %s", i.Number, i.Title),
		desc,
		i.HTMLURL,
	)
}

func (o *GroupedIssues) issueItems(issues gh.SearchResult[issue]) []list.Item {
	items := make([]list.Item, 0, len(issues.Items))
	for _, issue := range issues.Items {
		items = append(items, issue.toItem(o.currentLogin))
	}
	return items
}
