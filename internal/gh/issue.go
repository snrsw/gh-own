package gh

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

func SearchIssues(client *api.GraphQLClient, username string) (*IssueSearchResult, error) {
	if username == "" {
		return &IssueSearchResult{}, nil
	}

	entries := map[string]string{
		"created":          fmt.Sprintf("is:issue is:open author:%s", username),
		"assigned":         fmt.Sprintf("is:issue is:open assignee:%s", username),
		"participatedUser": fmt.Sprintf("is:issue is:open involves:%s -author:%s -assignee:%s", username, username, username),
	}

	raw, err := Search(client, issueSearchQuery, entries, parseIssueSearchJSON)
	if err != nil {
		return nil, err
	}

	return parseIssueSearchResult(raw)
}

func SearchIssuesTeams(client *api.GraphQLClient, username string, teams []string) (*IssueSearchResult, error) {
	if username == "" {
		return &IssueSearchResult{}, nil
	}

	if len(teams) == 0 {
		return &IssueSearchResult{}, nil
	}

	entries := map[string]string{}
	for i, team := range teams {
		entries[fmt.Sprintf("participatedTeam%d", i)] = fmt.Sprintf("is:issue is:open team:%s", team)
	}

	raw, err := Search(client, issueSearchQuery, entries, parseIssueSearchJSON)
	if err != nil {
		return nil, err
	}

	return parseIssueSearchResult(raw)
}

type IssueSearchResult struct {
	Created      []IssueSearchNode
	Assigned     []IssueSearchNode
	Participated []IssueSearchNode
}

func MergeSearchIssuesResults(a, b *IssueSearchResult) *IssueSearchResult {
	merged := &IssueSearchResult{
		Created:      append(a.Created, b.Created...),
		Assigned:     append(a.Assigned, b.Assigned...),
		Participated: append(a.Participated, b.Participated...),
	}
	return merged
}

func parseIssueSearchJSON(data json.RawMessage) ([]IssueSearchNode, error) {
	var sr struct {
		Nodes []issueSearchRawNode `json:"nodes"`
	}
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, err
	}
	return parseIssueSearchNodes(sr.Nodes), nil
}

const issueSearchQuery = `query($q: String!) {
	result: search(query: $q, type: ISSUE, first: 50) {
		nodes {
		... on Issue {
			number
			title
			url
			state
			updatedAt
			createdAt
			author { login }
				repository { nameWithOwner }
			}
		}
	}
}`

func parseIssueSearchResult(parsed map[string][]IssueSearchNode) (*IssueSearchResult, error) {
	var participated []IssueSearchNode
	for key, nodes := range parsed {
		switch {
		case strings.HasPrefix(key, "participated"):
			participated = append(participated, nodes...)
		}
	}

	return &IssueSearchResult{
		Created:      parsed["created"],
		Assigned:     parsed["assigned"],
		Participated: deduplicateIssueNodes(participated),
	}, nil
}

type IssueSearchNode struct {
	Number    int
	Title     string
	URL       string
	State     string
	UpdatedAt string
	CreatedAt string
	Author    struct {
		Login string
	}
	Repository struct {
		NameWithOwner string
	}
}

func (i *IssueSearchNode) RepositoryURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s", i.Repository.NameWithOwner)
}

type issueSearchRawNode struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	State     string `json:"state"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
}

func parseIssueSearchNodes(rawNodes []issueSearchRawNode) []IssueSearchNode {
	nodes := make([]IssueSearchNode, 0, len(rawNodes))
	for _, n := range rawNodes {
		if n.Number == 0 {
			continue
		}
		node := IssueSearchNode{
			Number:    n.Number,
			Title:     n.Title,
			URL:       n.URL,
			State:     n.State,
			UpdatedAt: n.UpdatedAt,
			CreatedAt: n.CreatedAt,
		}
		node.Author.Login = n.Author.Login
		node.Repository.NameWithOwner = n.Repository.NameWithOwner
		nodes = append(nodes, node)
	}
	return nodes
}

func deduplicateIssueNodes(nodes []IssueSearchNode) []IssueSearchNode {
	seen := make(map[string]bool)
	result := make([]IssueSearchNode, 0, len(nodes))
	for _, n := range nodes {
		if seen[n.URL] {
			continue
		}
		seen[n.URL] = true
		result = append(result, n)
	}
	return result
}
