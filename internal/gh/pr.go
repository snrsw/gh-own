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

	entries := map[string]string{
		"created":          fmt.Sprintf("is:pr is:open author:%s", username),
		"assigned":         fmt.Sprintf("is:pr is:open assignee:%s", username),
		"participatedUser": fmt.Sprintf("is:pr is:open involves:%s -author:%s -assignee:%s -review-requested:%s", username, username, username, username),
		"reviewRequested":  fmt.Sprintf("is:pr is:open review-requested:%s", username),
	}

	raw, err := Search(client, prSearchQuery, entries, parsePRSearchJSON)
	if err != nil {
		return nil, err
	}

	return parsePRSearchResult(raw)
}

func SearchPRsTeams(client *api.GraphQLClient, username string, teams []string) (*PRSearchResult, error) {
	if username == "" {
		return &PRSearchResult{}, nil
	}

	if len(teams) == 0 {
		return &PRSearchResult{}, nil
	}

	entries := make(map[string]string, len(teams))
	for i, team := range teams {
		entries[fmt.Sprintf("participatedTeam%d", i)] = fmt.Sprintf("is:pr is:open team:%s", team)
	}

	raw, err := Search(client, prSearchQuery, entries, parsePRSearchJSON)
	if err != nil {
		return nil, err
	}

	return parsePRSearchResult(raw)
}

type PRSearchResult struct {
	Created         []PRSearchNode
	Assigned        []PRSearchNode
	Participated    []PRSearchNode
	ReviewRequested []PRSearchNode
}

func MergeSearchPRsResults(a, b *PRSearchResult) *PRSearchResult {
	merged := &PRSearchResult{
		Created:         append(a.Created, b.Created...),
		Assigned:        append(a.Assigned, b.Assigned...),
		Participated:    append(a.Participated, b.Participated...),
		ReviewRequested: append(a.ReviewRequested, b.ReviewRequested...),
	}
	return merged
}

func parsePRSearchJSON(data json.RawMessage) ([]PRSearchNode, error) {
	var sr struct {
		Nodes []prSearchRawNode `json:"nodes"`
	}
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, err
	}
	return parsePRSearchNodes(sr.Nodes), nil
}

const prSearchQuery = `query($q: String!) {
	result: search(query: $q, type: ISSUE, first: 50) {
		nodes {
			... on PullRequest {
				number
				title
				url
				isDraft
				updatedAt
				createdAt
				reviewDecision
				author { login }
				repository { nameWithOwner }
				commits(last: 1) {
					nodes {
						commit {
							statusCheckRollup { state }
							committedDate
							author { user { login } }
						}
					}
				}
				comments(last: 1) {
					nodes { author { login } createdAt }
				}
				reviews(last: 1) {
					nodes { author { login } submittedAt state }
				}
			}
		}
	}
}`

func parsePRSearchResult(parsed map[string][]PRSearchNode) (*PRSearchResult, error) {
	var participated []PRSearchNode
	for key, nodes := range parsed {
		switch {
		case strings.HasPrefix(key, "participated"):
			participated = append(participated, nodes...)
		}
	}

	return &PRSearchResult{
		Created:         parsed["created"],
		Assigned:        parsed["assigned"],
		Participated:    deduplicatePRNodes(participated),
		ReviewRequested: parsed["reviewRequested"],
	}, nil
}

type PRSearchNode struct {
	Number         int
	Title          string
	URL            string
	IsDraft        bool
	UpdatedAt      string
	CreatedAt      string
	StatusState    string
	ReviewDecision string
	LatestActivity LatestActivity
	Author         struct {
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
	Number         int    `json:"number"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	IsDraft        bool   `json:"isDraft"`
	UpdatedAt      string `json:"updatedAt"`
	CreatedAt      string `json:"createdAt"`
	ReviewDecision string `json:"reviewDecision"`
	Author         struct {
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
				CommittedDate string `json:"committedDate"`
				Author        struct {
					User *struct {
						Login string `json:"login"`
					} `json:"user"`
				} `json:"author"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
	Comments struct {
		Nodes []struct {
			Author    struct{ Login string `json:"login"` } `json:"author"`
			CreatedAt string                                `json:"createdAt"`
		} `json:"nodes"`
	} `json:"comments"`
	Reviews struct {
		Nodes []struct {
			Author      struct{ Login string `json:"login"` } `json:"author"`
			SubmittedAt string                                `json:"submittedAt"`
			State       string                                `json:"state"`
		} `json:"nodes"`
	} `json:"reviews"`
}

func parsePRSearchNodes(rawNodes []prSearchRawNode) []PRSearchNode {
	nodes := make([]PRSearchNode, 0, len(rawNodes))
	for _, n := range rawNodes {
		if n.Number == 0 {
			continue
		}
		node := PRSearchNode{
			Number:         n.Number,
			Title:          n.Title,
			URL:            n.URL,
			IsDraft:        n.IsDraft,
			UpdatedAt:      n.UpdatedAt,
			CreatedAt:      n.CreatedAt,
			ReviewDecision: n.ReviewDecision,
		}
		node.Author.Login = n.Author.Login
		node.Repository.NameWithOwner = n.Repository.NameWithOwner

		if len(n.Commits.Nodes) > 0 && n.Commits.Nodes[0].Commit.StatusCheckRollup != nil {
			node.StatusState = n.Commits.Nodes[0].Commit.StatusCheckRollup.State
		}

		var commentLogin, commentAt string
		if len(n.Comments.Nodes) > 0 {
			commentLogin = n.Comments.Nodes[0].Author.Login
			commentAt = n.Comments.Nodes[0].CreatedAt
		}
		var reviewLogin, reviewAt, reviewState string
		if len(n.Reviews.Nodes) > 0 {
			reviewLogin = n.Reviews.Nodes[0].Author.Login
			reviewAt = n.Reviews.Nodes[0].SubmittedAt
			reviewState = n.Reviews.Nodes[0].State
		}
		var pushLogin, pushAt string
		if len(n.Commits.Nodes) > 0 && n.Commits.Nodes[0].Commit.Author.User != nil {
			pushLogin = n.Commits.Nodes[0].Commit.Author.User.Login
			pushAt = n.Commits.Nodes[0].Commit.CommittedDate
		}
		node.LatestActivity = NewLatestActivity(commentLogin, commentAt, reviewLogin, reviewAt, reviewState, pushLogin, pushAt)

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
