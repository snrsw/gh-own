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

	variables := map[string]interface{}{
		"created":          fmt.Sprintf("is:issue is:open author:%s", username),
		"assigned":         fmt.Sprintf("is:issue is:open assignee:%s", username),
		"participatedUser": fmt.Sprintf("is:issue is:open involves:%s -author:%s -assignee:%s", username, username, username),
	}

	var result map[string]json.RawMessage
	if err := client.Do(buildIssueSearchQuery(), variables, &result); err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	return parseIssueSearchResult(result)
}

func SearchIssuesTeams(client *api.GraphQLClient, username string, teams []string) (*IssueSearchResult, error) {
	if username == "" {
		return &IssueSearchResult{}, nil
	}

	if len(teams) == 0 {
		return &IssueSearchResult{}, nil
	}

	variables := map[string]interface{}{}
	for i, team := range teams {
		variables[fmt.Sprintf("participatedTeam%d", i)] = fmt.Sprintf("is:issue is:open team:%s", team)
	}

	var result map[string]json.RawMessage
	if err := client.Do(buildIssueSearchQueryTeams(teams), variables, &result); err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	return parseIssueSearchResult(result)
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

func buildIssueSearchQuery() string {
	vars := []string{
		"created",
		"assigned",
		"participatedUser",
	}

	params := ""
	body := ""
	for _, v := range vars {
		params += fmt.Sprintf("$%s: String!", v)
		body += fmt.Sprintf("  %s: search(query: $%s, type: ISSUE, first: 50) %s\n", v, v, issueSearchFragment)
	}

	return fmt.Sprintf("query(%s) {\n%s}", params, body)
}

func buildIssueSearchQueryTeams(teams []string) string {
	vars := []string{}

	for i := range teams {
		vars = append(vars, fmt.Sprintf("participatedTeam%d", i))
	}

	params := ""
	body := ""
	for _, v := range vars {
		params += fmt.Sprintf("$%s: String!", v)
		body += fmt.Sprintf("  %s: search(query: $%s, type: ISSUE, first: 50) %s\n", v, v, issueSearchFragment)
	}

	return fmt.Sprintf("query(%s) {\n%s}", params, body)
}

const issueSearchFragment = `{
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
  }`

func parseIssueSearchResult(raw map[string]json.RawMessage) (*IssueSearchResult, error) {
	parsed := make(map[string][]IssueSearchNode)
	for key, data := range raw {
		var result struct {
			Nodes []issueSearchRawNode `json:"nodes"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse issue search result for %s: %w", key, err)
		}
		parsed[key] = parseIssueSearchNodes(result.Nodes)
	}

	var paticipated []IssueSearchNode
	for key, nodes := range parsed {
		switch {
		case strings.HasPrefix(key, "participated"):
			paticipated = append(paticipated, nodes...)
		}
	}

	return &IssueSearchResult{
		Created:      parsed["created"],
		Assigned:     parsed["assigned"],
		Participated: deduplicateIssueNodes(paticipated),
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
