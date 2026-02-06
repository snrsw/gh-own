// Package pullrequest provides functionality to handle GitHub pull requests owned by a user.
package pullrequest

import (
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/gh"
)

type PullRequest struct {
	Number        int     `json:"number"`
	User          gh.User `json:"user"`
	RepositoryURL string  `json:"repository_url"`
	Title         string  `json:"title"`
	State         string  `json:"state"`
	HTMLURL       string  `json:"html_url"`
	Draft         bool    `json:"draft"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
}

func (p *PullRequest) RepositoryFullName() string {
	// Format: "https://api.github.com/repos/owner/repo"
	parts := strings.Split(p.RepositoryURL, "/")
	if len(parts) < 5 {
		return ""
	}
	return parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

type GroupedPullRequests struct {
	Created         gh.SearchResult[PullRequest]
	Assigned        gh.SearchResult[PullRequest]
	ReviewRequested gh.SearchResult[PullRequest]
	Participated    gh.SearchResult[PullRequest]
}

func SearchPullRequests(client *api.RESTClient, username string) (*GroupedPullRequests, error) {
	conditions := []gh.Condition{
		{Name: "created", Query: fmt.Sprintf("is:pr is:open author:%s", username)},
		{Name: "assigned", Query: fmt.Sprintf("is:pr is:open assignee:%s", username)},
		{Name: "participated", Query: fmt.Sprintf("is:pr is:open (mentions:%s OR commenter:%s)", username, username)},
		{Name: "review-requested", Query: fmt.Sprintf("is:pr is:open review-requested:%s", username)},
	}

	results, err := gh.Search[PullRequest](client, conditions)
	if err != nil {
		return nil, err
	}

	return &GroupedPullRequests{
		Created:         results["created"],
		Assigned:        results["assigned"],
		Participated:    results["participated"],
		ReviewRequested: results["review-requested"],
	}, nil
}
