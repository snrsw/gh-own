// Package pr provides functionality to handle GitHub pull requests owned by a user.
package pr

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cistatus"
	"github.com/snrsw/gh-own/internal/gh"
)

func SearchPullRequests(client *api.GraphQLClient, username string, teams []string) (*groupedPullRequests, error) {
	results, err := gh.SearchPRs(client, username, teams)
	if err != nil {
		return nil, err
	}

	return &groupedPullRequests{
		Created:         toSearchResult(results.Created),
		Assigned:        toSearchResult(results.Assigned),
		ReviewRequested: toSearchResult(results.ReviewRequested),
		Participated:    toSearchResult(results.Participated),
	}, nil
}

type groupedPullRequests struct {
	Created         gh.SearchResult[pullRequest]
	Assigned        gh.SearchResult[pullRequest]
	ReviewRequested gh.SearchResult[pullRequest]
	Participated    gh.SearchResult[pullRequest]
}

type pullRequest struct {
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
	return pullRequest{
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
