// Package issue provides functionality to handle GitHub issues owned by a user.
package issue

import (
	"fmt"
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
	Created      gh.SearchResult[Issue]
	Assigned     gh.SearchResult[Issue]
	Participated gh.SearchResult[Issue]
}

func SearchIssues(client *api.RESTClient, username string) (*GroupedIssues, error) {
	conditions := []gh.Condition{
		{Name: "created", Query: fmt.Sprintf("is:issue is:open author:%s", username)},
		{Name: "assigned", Query: fmt.Sprintf("is:issue is:open assignee:%s", username)},
		{Name: "participated", Query: fmt.Sprintf("is:issue is:open (mentions:%s OR commenter:%s)", username, username)},
	}

	results, err := gh.Search[Issue](client, conditions)
	if err != nil {
		return nil, err
	}

	return &GroupedIssues{
		Created:      results["created"],
		Assigned:     results["assigned"],
		Participated: results["participated"],
	}, nil
}
