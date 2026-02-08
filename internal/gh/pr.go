package gh

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cistatus"
)

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

const prSearchQuery = `
query($created: String!, $assigned: String!, $participated: String!, $reviewRequested: String!) {
  created: search(query: $created, type: ISSUE, first: 50) {
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
  }
  assigned: search(query: $assigned, type: ISSUE, first: 50) {
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
  }
  participated: search(query: $participated, type: ISSUE, first: 50) {
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
  }
  reviewRequested: search(query: $reviewRequested, type: ISSUE, first: 50) {
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
  }
}
`

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

func buildPRSearchVariables(username string) map[string]interface{} {
	return map[string]interface{}{
		"created":         fmt.Sprintf("is:pr is:open author:%s", username),
		"assigned":        fmt.Sprintf("is:pr is:open assignee:%s", username),
		"participated":    fmt.Sprintf("is:pr is:open (mentions:%s OR commenter:%s)", username, username),
		"reviewRequested": fmt.Sprintf("is:pr is:open review-requested:%s", username),
	}
}

type prBatchedSearchResult struct {
	Created struct {
		Nodes []prSearchRawNode `json:"nodes"`
	} `json:"created"`
	Assigned struct {
		Nodes []prSearchRawNode `json:"nodes"`
	} `json:"assigned"`
	Participated struct {
		Nodes []prSearchRawNode `json:"nodes"`
	} `json:"participated"`
	ReviewRequested struct {
		Nodes []prSearchRawNode `json:"nodes"`
	} `json:"reviewRequested"`
}

type PRSearchResult struct {
	Created         []PRSearchNode
	Assigned        []PRSearchNode
	Participated    []PRSearchNode
	ReviewRequested []PRSearchNode
}

func SearchPRs(client *api.GraphQLClient, username string) (*PRSearchResult, error) {
	if username == "" {
		return &PRSearchResult{}, nil
	}

	variables := buildPRSearchVariables(username)

	var result prBatchedSearchResult
	if err := client.Do(prSearchQuery, variables, &result); err != nil {
		return nil, fmt.Errorf("failed to search pull requests: %w", err)
	}

	return &PRSearchResult{
		Created:         parsePRSearchNodes(result.Created.Nodes),
		Assigned:        parsePRSearchNodes(result.Assigned.Nodes),
		Participated:    parsePRSearchNodes(result.Participated.Nodes),
		ReviewRequested: parsePRSearchNodes(result.ReviewRequested.Nodes),
	}, nil
}
