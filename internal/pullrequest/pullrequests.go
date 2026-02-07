// Package pullrequest provides functionality to handle GitHub pull requests owned by a user.
package pullrequest

import (
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cistatus"
	"github.com/snrsw/gh-own/internal/gh"
)

type PullRequest struct {
	Number        int               `json:"number"`
	User          gh.User           `json:"user"`
	RepositoryURL string            `json:"repository_url"`
	Title         string            `json:"title"`
	State         string            `json:"state"`
	HTMLURL       string            `json:"html_url"`
	Draft         bool              `json:"draft"`
	UpdatedAt     string            `json:"updated_at"`
	CreatedAt     string            `json:"created_at"`
	CIStatus      cistatus.CIStatus `json:"-"`
}

func (p *PullRequest) RepositoryFullName() string {
	// Format: "https://api.github.com/repos/owner/repo"
	parts := strings.Split(p.RepositoryURL, "/")
	if len(parts) < 5 {
		return ""
	}
	return parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

func FromGraphQL(node gh.PRSearchNode) PullRequest {
	return PullRequest{
		Number:        node.Number,
		User:          gh.User{Login: node.Author.Login},
		RepositoryURL: node.RepositoryURL(),
		Title:         node.Title,
		HTMLURL:       node.URL,
		Draft:         node.IsDraft,
		UpdatedAt:     node.UpdatedAt,
		CreatedAt:     node.CreatedAt,
		CIStatus:      node.CIStatus(),
	}
}

func FromGraphQLNodes(nodes []gh.PRSearchNode) []PullRequest {
	prs := make([]PullRequest, len(nodes))
	for i, node := range nodes {
		prs[i] = FromGraphQL(node)
	}
	return prs
}

type GroupedPullRequests struct {
	Created         gh.SearchResult[PullRequest]
	Assigned        gh.SearchResult[PullRequest]
	ReviewRequested gh.SearchResult[PullRequest]
	Participated    gh.SearchResult[PullRequest]
}

func BuildConditions(username string) []gh.Condition {
	return []gh.Condition{
		{Name: "created", Query: fmt.Sprintf("is:pr is:open author:%s", username)},
		{Name: "assigned", Query: fmt.Sprintf("is:pr is:open assignee:%s", username)},
		{Name: "participated", Query: fmt.Sprintf("is:pr is:open (mentions:%s OR commenter:%s)", username, username)},
		{Name: "review-requested", Query: fmt.Sprintf("is:pr is:open review-requested:%s", username)},
	}
}

func SearchPullRequests(client *api.GraphQLClient, username string) (*GroupedPullRequests, error) {
	conditions := BuildConditions(username)

	results, err := gh.SearchGraphQL(client, conditions)
	if err != nil {
		return nil, err
	}

	return &GroupedPullRequests{
		Created:         toSearchResult(results["created"]),
		Assigned:        toSearchResult(results["assigned"]),
		Participated:    toSearchResult(results["participated"]),
		ReviewRequested: toSearchResult(results["review-requested"]),
	}, nil
}

func toSearchResult(nodes []gh.PRSearchNode) gh.SearchResult[PullRequest] {
	prs := FromGraphQLNodes(nodes)
	return gh.SearchResult[PullRequest]{
		TotalCount: len(prs),
		Items:      prs,
	}
}
