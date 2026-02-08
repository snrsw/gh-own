package gh

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
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

const issueSearchQuery = `
query($created: String!, $assigned: String!, $participated: String!) {
  created: search(query: $created, type: ISSUE, first: 50) {
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
  assigned: search(query: $assigned, type: ISSUE, first: 50) {
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
  participated: search(query: $participated, type: ISSUE, first: 50) {
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
}
`

func buildIssueSearchVariables(username string) map[string]interface{} {
	return map[string]interface{}{
		"created":      fmt.Sprintf("is:issue is:open author:%s", username),
		"assigned":     fmt.Sprintf("is:issue is:open assignee:%s", username),
		"participated": fmt.Sprintf("is:issue is:open involves:%s -author:%s -assignee:%s", username, username, username),
	}
}

type issueSearchResult struct {
	Created struct {
		Nodes []issueSearchRawNode `json:"nodes"`
	} `json:"created"`
	Assigned struct {
		Nodes []issueSearchRawNode `json:"nodes"`
	} `json:"assigned"`
	Participated struct {
		Nodes []issueSearchRawNode `json:"nodes"`
	} `json:"participated"`
}

type IssueSearchResult struct {
	Created      []IssueSearchNode
	Assigned     []IssueSearchNode
	Participated []IssueSearchNode
}

func SearchIssues(client *api.GraphQLClient, username string) (*IssueSearchResult, error) {
	if username == "" {
		return &IssueSearchResult{}, nil
	}

	variables := buildIssueSearchVariables(username)

	var result issueSearchResult
	if err := client.Do(issueSearchQuery, variables, &result); err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	return &IssueSearchResult{
		Created:      parseIssueSearchNodes(result.Created.Nodes),
		Assigned:     parseIssueSearchNodes(result.Assigned.Nodes),
		Participated: parseIssueSearchNodes(result.Participated.Nodes),
	}, nil
}
