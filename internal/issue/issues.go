// Package issue provides functionality to handle GitHub issues owned by a user.
package issue

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/gh"
)

type Issue struct {
	Number        int     `json:"number"`
	User          gh.User `json:"user"`
	RepositoryURL string  `json:"repository_url"`
	Title         string  `json:"title"`
	State         string  `json:"state"`
	HTMLURL       string  `json:"html_url"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
}

func (i *Issue) RepositoryFullName() string {
	// Format: "https://api.github.com/repos/owner/repo"
	parts := strings.Split(i.RepositoryURL, "/")
	if len(parts) < 5 {
		return ""
	}
	return parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

type GroupedIssues struct {
	Created      []Issue
	Assigned     []Issue
	Participated []Issue
}

func issueFromNode(node gh.IssueSearchNode) Issue {
	return Issue{
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

func issuesFromNodes(nodes []gh.IssueSearchNode) []Issue {
	issues := make([]Issue, len(nodes))
	for i, node := range nodes {
		issues[i] = issueFromNode(node)
	}
	return issues
}

func SearchIssues(client *api.GraphQLClient, username string) (*GroupedIssues, error) {
	results, err := gh.SearchIssuesGraphQL(client, username)
	if err != nil {
		return nil, err
	}

	return &GroupedIssues{
		Created:      issuesFromNodes(results["created"]),
		Assigned:     issuesFromNodes(results["assigned"]),
		Participated: issuesFromNodes(results["participated"]),
	}, nil
}
