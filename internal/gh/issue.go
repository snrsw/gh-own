package gh

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

func SearchIssues(client *api.GraphQLClient, username string, teams []string) (*issueSearchResult, error) {
	if username == "" {
		return &issueSearchResult{}, nil
	}

	variables := buildIssueSearchVariables(username, teams)

	var result map[string]json.RawMessage
	if err := client.Do(buildIssueSearchQuery(teams), variables, &result); err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	return parseIssueSearchResult(result)
}

type issueSearchResult struct {
	Created      []IssueSearchNode
	Assigned     []IssueSearchNode
	Participated []IssueSearchNode
}

func buildIssueSearchVariables(username string, teams []string) map[string]interface{} {
	vars := map[string]interface{}{
		"created":          fmt.Sprintf("is:issue is:open author:%s", username),
		"assigned":         fmt.Sprintf("is:issue is:open assignee:%s", username),
		"participatedUser": fmt.Sprintf("is:issue is:open involves:%s -author:%s -assignee:%s", username, username, username),
	}

	if len(teams) == 0 {
		return vars
	}

	for i, team := range teams {
		vars[fmt.Sprintf("participatedTeam%d", i)] = fmt.Sprintf("is:issue is:open team:%s", team)
	}

	return vars
}

func buildIssueSearchQuery(teams []string) string {
	vars := []string{
		"created",
		"assigned",
		"participatedUser",
	}

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

func parseIssueSearchResult(raw map[string]json.RawMessage) (*issueSearchResult, error) {
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

	return &issueSearchResult{
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
