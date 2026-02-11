// Package issue provides functionality to handle GitHub issues owned by a user.
package issue

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/gh"
)

func SearchIssues(client *api.GraphQLClient, username string, teams []string) (*groupedIssues, error) {
	results, err := gh.SearchIssues(client, username, teams)
	if err != nil {
		return nil, err
	}

	return &groupedIssues{
		Created:      toSearchResult(results.Created),
		Assigned:     toSearchResult(results.Assigned),
		Participated: toSearchResult(results.Participated),
	}, nil
}

type groupedIssues struct {
	Created      gh.SearchResult[issue]
	Assigned     gh.SearchResult[issue]
	Participated gh.SearchResult[issue]
}

type issue struct {
	Number        int     `json:"number"`
	User          gh.User `json:"user"`
	RepositoryURL string  `json:"repository_url"`
	Title         string  `json:"title"`
	State         string  `json:"state"`
	HTMLURL       string  `json:"html_url"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
}

func (i *issue) repositoryFullName() string {
	// Format: "https://api.github.com/repos/owner/repo"
	parts := strings.Split(i.RepositoryURL, "/")
	if len(parts) < 5 {
		return ""
	}
	return parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

func toSearchResult(nodes []gh.IssueSearchNode) gh.SearchResult[issue] {
	issues := fromGraphQLNodes(nodes)
	return gh.SearchResult[issue]{
		TotalCount: len(issues),
		Items:      issues,
	}
}

func fromGraphQLNodes(nodes []gh.IssueSearchNode) []issue {
	issues := make([]issue, len(nodes))
	for i, node := range nodes {
		issues[i] = fromGraphQL(node)
	}
	return issues
}

func fromGraphQL(node gh.IssueSearchNode) issue {
	return issue{
		Number:        node.Number,
		User:          gh.User{Login: node.Author.Login},
		RepositoryURL: node.RepositoryURL(),
		Title:         node.Title,
		State:         node.State,
		HTMLURL:       node.URL,
		UpdatedAt:     node.UpdatedAt,
		CreatedAt:     node.CreatedAt,
	}
}
