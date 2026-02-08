// Package pullrequest provides functionality to handle GitHub pull requests owned by a user.
package pullrequest

import (
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

func SearchPullRequests(client *api.GraphQLClient, username string) (*GroupedPullRequests, error) {
	results, err := gh.SearchPRs(client, username)
	if err != nil {
		return nil, err
	}

	return &GroupedPullRequests{
		Created:         toSearchResult(results.Created),
		Assigned:        toSearchResult(results.Assigned),
		ReviewRequested: toSearchResult(results.ReviewRequested),
		Participated:    toSearchResult(results.Participated),
	}, nil
}

func toSearchResult(nodes []gh.PRSearchNode) gh.SearchResult[PullRequest] {
	prs := FromGraphQLNodes(nodes)
	return gh.SearchResult[PullRequest]{
		TotalCount: len(prs),
		Items:      prs,
	}
}
