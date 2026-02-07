package gh

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cistatus"
)

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

// PRSearchNode represents a pull request from GraphQL search.
type PRSearchNode struct {
	Number     int
	Title      string
	URL        string
	IsDraft    bool
	UpdatedAt  string
	CreatedAt  string
	StatusState string
	Author     struct {
		Login string
	}
	Repository struct {
		NameWithOwner string
	}
}

// CIStatus returns the CI status parsed from GraphQL statusCheckRollup state.
func (p *PRSearchNode) CIStatus() cistatus.CIStatus {
	return cistatus.ParseState(p.StatusState)
}

// RepositoryURL returns the REST API URL for the repository.
func (p *PRSearchNode) RepositoryURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s", p.Repository.NameWithOwner)
}

const searchQuery = `
query($q: String!) {
  search(query: $q, type: ISSUE, first: 50) {
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

// BuildSearchQuery returns the GraphQL query string.
func BuildSearchQuery(q string) string {
	return searchQuery
}

// PRSearchResult holds the GraphQL search response.
type PRSearchResult struct {
	Search struct {
		Nodes []struct {
			Number     int    `json:"number"`
			Title      string `json:"title"`
			URL        string `json:"url"`
			IsDraft    bool   `json:"isDraft"`
			UpdatedAt  string `json:"updatedAt"`
			CreatedAt  string `json:"createdAt"`
			Author     struct {
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
		} `json:"nodes"`
	} `json:"search"`
}

type graphqlSearchResult struct {
	name  string
	query string
	data  []PRSearchNode
	err   error
}

// SearchGraphQL searches PRs using GraphQL API with parallel queries.
func SearchGraphQL(client *api.GraphQLClient, conditions []Condition) (map[string][]PRSearchNode, error) {
	if len(conditions) == 0 {
		return make(map[string][]PRSearchNode), nil
	}

	ch := make(chan graphqlSearchResult, len(conditions))
	for _, cond := range conditions {
		go func(c Condition) {
			var result PRSearchResult
			variables := map[string]interface{}{
				"q": c.Query,
			}

			err := client.Do(searchQuery, variables, &result)
			if err != nil {
				ch <- graphqlSearchResult{name: c.Name, query: c.Query, err: err}
				return
			}

			nodes := make([]PRSearchNode, 0, len(result.Search.Nodes))
			for _, n := range result.Search.Nodes {
				if n.Number == 0 {
					continue // Skip non-PR nodes
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

				// Extract CI status from commits
				if len(n.Commits.Nodes) > 0 && n.Commits.Nodes[0].Commit.StatusCheckRollup != nil {
					node.StatusState = n.Commits.Nodes[0].Commit.StatusCheckRollup.State
				}

				nodes = append(nodes, node)
			}

			ch <- graphqlSearchResult{name: c.Name, query: c.Query, data: nodes}
		}(cond)
	}

	var allResults []graphqlSearchResult
	for range conditions {
		allResults = append(allResults, <-ch)
	}

	results := make(map[string][]PRSearchNode, len(conditions))
	for _, res := range allResults {
		if res.err != nil {
			return nil, fmt.Errorf("failed to search for query '%s': %w", res.query, res.err)
		}
		results[res.name] = res.data
	}

	return results, nil
}

func SearchIssuesGraphQL(client *api.GraphQLClient, username string) (map[string][]IssueSearchNode, error) {
	if username == "" {
		return make(map[string][]IssueSearchNode), nil
	}

	return make(map[string][]IssueSearchNode), nil
}
