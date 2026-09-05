// Package pr provides functionality to handle GitHub pull requests owned by a user.
package pr

import (
	"strings"
	"time"

	"github.com/snrsw/gh-own/internal/cistatus"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/reviewstatus"
)

type GroupedPullRequests struct {
	Drafts          gh.SearchResult[pullRequest]
	NeedsAction     gh.SearchResult[pullRequest]
	ReadyToMerge    gh.SearchResult[pullRequest]
	Waiting         gh.SearchResult[pullRequest]
	ReviewRequested gh.SearchResult[pullRequest]
	Participated    gh.SearchResult[pullRequest]
	Custom          map[string]gh.SearchResult[pullRequest]
	currentLogin    string
}

func NewGroupedPullRequests(ghResult *gh.PRSearchResult, currentLogin string) *GroupedPullRequests {
	ghResult = promoteNeedsAction(ghResult, currentLogin)
	custom := make(map[string]gh.SearchResult[pullRequest], len(ghResult.Custom))
	for k, nodes := range ghResult.Custom {
		custom[k] = toSearchResult(nodes)
	}

	return &GroupedPullRequests{
		Drafts:          toSearchResult(ghResult.Drafts),
		NeedsAction:     toSearchResult(ghResult.NeedsAction),
		ReadyToMerge:    toSearchResult(ghResult.ReadyToMerge),
		Waiting:         toSearchResult(ghResult.Waiting),
		ReviewRequested: toSearchResult(ghResult.ReviewRequested),
		Participated:    toSearchResult(ghResult.Participated),
		Custom:          custom,
		currentLogin:    currentLogin,
	}
}

type pullRequest struct {
	Number         int                       `json:"number"`
	User           gh.User                   `json:"user"`
	RepositoryURL  string                    `json:"repository_url"`
	Title          string                    `json:"title"`
	State          string                    `json:"state"`
	HTMLURL        string                    `json:"html_url"`
	Draft          bool                      `json:"draft"`
	UpdatedAt      string                    `json:"updated_at"`
	CreatedAt      string                    `json:"created_at"`
	CIStatus       cistatus.CIStatus         `json:"-"`
	ReviewStatus   reviewstatus.ReviewStatus `json:"-"`
	LatestActivity gh.LatestActivity         `json:"-"`
	// Attention says why the pull request is waiting on the user, when it is.
	Attention gh.Attention `json:"-"`
	// SortAt is the timestamp the list is ordered by. It matches the time shown
	// on the description line (see toItem).
	SortAt time.Time `json:"-"`
}

// shownActivity is the event the description line reports: what is waiting on
// the user when something is, and the latest activity otherwise.
func (p *pullRequest) shownActivity() gh.LatestActivity {
	if p.Attention.Reason != "" {
		return gh.LatestActivity{Kind: p.Attention.Reason, Login: p.Attention.Login, At: p.Attention.At}
	}
	return p.LatestActivity
}

func (p *pullRequest) repositoryFullName() string {
	// Format: "https://api.github.com/repos/owner/repo"
	parts := strings.Split(p.RepositoryURL, "/")
	if len(parts) < 5 {
		return ""
	}
	return parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

func toSearchResult(nodes []gh.PRSearchNode) gh.SearchResult[pullRequest] {
	prs := fromGraphQLNodes(nodes)
	gh.SortByUpdatedDesc(prs, func(p pullRequest) time.Time { return p.SortAt })
	return gh.SearchResult[pullRequest]{
		TotalCount: len(prs),
		Items:      prs,
	}
}

func fromGraphQLNodes(nodes []gh.PRSearchNode) []pullRequest {
	prs := make([]pullRequest, len(nodes))
	for i, node := range nodes {
		prs[i] = fromGraphQL(node)
	}
	return prs
}

func fromGraphQL(node gh.PRSearchNode) pullRequest {
	p := pullRequest{
		Number:         node.Number,
		User:           gh.User{Login: node.Author.Login},
		RepositoryURL:  node.RepositoryURL(),
		Title:          node.Title,
		HTMLURL:        node.URL,
		Draft:          node.IsDraft,
		UpdatedAt:      node.UpdatedAt,
		CreatedAt:      node.CreatedAt,
		CIStatus:       node.CIStatus(),
		ReviewStatus:   reviewstatus.ParseReviewDecision(node.ReviewDecision),
		LatestActivity: node.LatestActivity,
		Attention:      node.Attention,
	}
	p.SortAt = gh.SortAt(p.shownActivity(), node.UpdatedAt)
	return p
}
