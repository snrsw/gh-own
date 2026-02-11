package gh

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cistatus"
)

func SearchPRs(client *api.GraphQLClient, username string) (*PRSearchResult, error) {
	if username == "" {
		return &PRSearchResult{}, nil
	}

	variables := map[string]interface{}{
		"created":             fmt.Sprintf("is:pr is:open author:%s", username),
		"assigned":            fmt.Sprintf("is:pr is:open assignee:%s", username),
		"participatedUser":    fmt.Sprintf("is:pr is:open (mentions:%s OR commenter:%s)", username, username),
		"reviewRequestedUser": fmt.Sprintf("is:pr is:open review-requested:%s", username),
	}

	var result map[string]json.RawMessage
	if err := client.Do(buildPRSearchQuery(), variables, &result); err != nil {
		return nil, fmt.Errorf("failed to search pull requests: %w", err)
	}

	return parsePRSearchResult(result)
}

func SearchPRsTeams(client *api.GraphQLClient, username string, teams []string) (*PRSearchResult, error) {
	if username == "" {
		return &PRSearchResult{}, nil
	}

	if len(teams) == 0 {
		return &PRSearchResult{}, nil
	}

	variables := map[string]interface{}{}
	for i, team := range teams {
		variables[fmt.Sprintf("reviewRequestedTeam%d", i)] = fmt.Sprintf("is:pr is:open team-review-requested:%s", team)
		variables[fmt.Sprintf("participatedTeam%d", i)] = fmt.Sprintf("is:pr is:open team:%s", team)
	}

	var result map[string]json.RawMessage
	if err := client.Do(buildPRSearchQueryTeams(teams), variables, &result); err != nil {
		return nil, fmt.Errorf("failed to search pull requests: %w", err)
	}

	return parsePRSearchResult(result)
}

type PRSearchResult struct {
	Created         []PRSearchNode
	Assigned        []PRSearchNode
	Participated    []PRSearchNode
	ReviewRequested []PRSearchNode
}

func MergeSearchResults(a, b *PRSearchResult) *PRSearchResult {
	merged := &PRSearchResult{
		Created:         append(a.Created, b.Created...),
		Assigned:        append(a.Assigned, b.Assigned...),
		Participated:    append(a.Participated, b.Participated...),
		ReviewRequested: append(a.ReviewRequested, b.ReviewRequested...),
	}
	return merged
}

func buildPRSearchQuery() string {
	vars := []string{"created", "assigned", "participatedUser", "reviewRequestedUser"}

	params := ""
	body := ""
	for _, v := range vars {
		params += fmt.Sprintf("$%s: String!", v)
		body += fmt.Sprintf("  %s: search(query: $%s, type: ISSUE, first: 50) %s\n", v, v, prSearchFragment)
	}

	return fmt.Sprintf("query(%s) {\n%s}", params, body)
}

func buildPRSearchQueryTeams(teams []string) string {
	vars := []string{}

	for i := range teams {
		vars = append(vars, fmt.Sprintf("reviewRequestedTeam%d", i))
		vars = append(vars, fmt.Sprintf("participatedTeam%d", i))
	}

	params := ""
	body := ""
	for _, v := range vars {
		params += fmt.Sprintf("$%s: String!", v)
		body += fmt.Sprintf("  %s: search(query: $%s, type: ISSUE, first: 50) %s\n", v, v, prSearchFragment)
	}

	return fmt.Sprintf("query(%s) {\n%s}", params, body)
}

const prSearchFragment = `{
    nodes {
      ... on PullRequest {
        number
        title
        url
        isDraft
        updatedAt
        createdAt
        author { login }
        repository { nameWithOwner }
        commits(last: 1) {
          nodes {
            commit {
              statusCheckRollup { state }
            }
          }
        }
      }
    }
  }`

func parsePRSearchResult(raw map[string]json.RawMessage) (*PRSearchResult, error) {
	parsed := make(map[string][]PRSearchNode)
	for key, data := range raw {
		var result struct {
			Nodes []prSearchRawNode `json:"nodes"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse PR search result for %s: %w", key, err)
		}
		parsed[key] = parsePRSearchNodes(result.Nodes)
	}

	var paticipated, reviewRequested []PRSearchNode
	for key, nodes := range parsed {
		switch {
		case strings.HasPrefix(key, "participated"):
			paticipated = append(paticipated, nodes...)
		case strings.HasPrefix(key, "reviewRequested"):
			reviewRequested = append(reviewRequested, nodes...)
		}
	}

	return &PRSearchResult{
		Created:         parsed["created"],
		Assigned:        parsed["assigned"],
		Participated:    deduplicatePRNodes(paticipated),
		ReviewRequested: deduplicatePRNodes(reviewRequested),
	}, nil
}

type PRSearchNode struct {
	Number      int
	Title       string
	URL         string
	IsDraft     bool
	UpdatedAt   string
	CreatedAt   string
	StatusState string
	Author      struct {
		Login string
	}
	Repository struct {
		NameWithOwner string
	}
}

func (p *PRSearchNode) CIStatus() cistatus.CIStatus {
	return cistatus.ParseState(p.StatusState)
}

func (p *PRSearchNode) RepositoryURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s", p.Repository.NameWithOwner)
}

type prSearchRawNode struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	IsDraft   bool   `json:"isDraft"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *struct {
					State string `json:"state"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
}

func parsePRSearchNodes(rawNodes []prSearchRawNode) []PRSearchNode {
	nodes := make([]PRSearchNode, 0, len(rawNodes))
	for _, n := range rawNodes {
		if n.Number == 0 {
			continue
		}
		node := PRSearchNode{
			Number:    n.Number,
			Title:     n.Title,
			URL:       n.URL,
			IsDraft:   n.IsDraft,
			UpdatedAt: n.UpdatedAt,
			CreatedAt: n.CreatedAt,
		}
		node.Author.Login = n.Author.Login
		node.Repository.NameWithOwner = n.Repository.NameWithOwner

		if len(n.Commits.Nodes) > 0 && n.Commits.Nodes[0].Commit.StatusCheckRollup != nil {
			node.StatusState = n.Commits.Nodes[0].Commit.StatusCheckRollup.State
		}

		nodes = append(nodes, node)
	}
	return nodes
}

func deduplicatePRNodes(nodes []PRSearchNode) []PRSearchNode {
	seen := make(map[string]bool)
	result := make([]PRSearchNode, 0, len(nodes))
	for _, n := range nodes {
		if seen[n.URL] {
			continue
		}
		seen[n.URL] = true
		result = append(result, n)
	}
	return result
}
