// Package gh provides GitHub API client helpers.
package gh

import (
	"fmt"
	"net/url"

	"github.com/cli/go-gh/v2/pkg/api"
)

type User struct {
	Login string `json:"login"`
}

// CurrentUser returns a GraphQL client and the current user's login.
func CurrentUser() (*api.GraphQLClient, string, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, "", err
	}

	var query struct {
		Viewer struct {
			Login string
		}
	}
	if err := client.Query("viewer", &query, nil); err != nil {
		return nil, "", fmt.Errorf("failed to get current user: %w", err)
	}

	return client, query.Viewer.Login, nil
}

// CurrentUserREST returns a REST client and the current user's login.
func CurrentUserREST() (*api.RESTClient, string, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, "", err
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := client.Get("user", &user); err != nil {
		return nil, "", fmt.Errorf("failed to get current user: %w", err)
	}

	return client, user.Login, nil
}

type SearchResult[T any] struct {
	TotalCount int `json:"total_count"`
	Items      []T `json:"items"`
}

type Condition struct {
	Name  string
	Query string
}

type searchResult[T any] struct {
	name  string
	query string
	data  SearchResult[T]
	err   error
}

func Search[T any](client *api.RESTClient, conditions []Condition) (map[string]SearchResult[T], error) {
	if len(conditions) == 0 {
		return make(map[string]SearchResult[T]), nil
	}

	ch := make(chan searchResult[T], len(conditions))
	for _, cond := range conditions {
		go func(c Condition) {
			path := fmt.Sprintf(
				"search/issues?q=%s&sort=updated&order=desc&per_page=50&advanced_search=true",
				url.QueryEscape(c.Query),
			)

			var r SearchResult[T]
			err := client.Get(path, &r)
			ch <- searchResult[T]{name: c.Name, query: c.Query, data: r, err: err}
		}(cond)
	}

	var allResults []searchResult[T]
	for range conditions {
		allResults = append(allResults, <-ch)
	}

	results := make(map[string]SearchResult[T], len(conditions))
	for _, res := range allResults {
		if res.err != nil {
			return nil, fmt.Errorf("failed to search for query '%s': %w", res.query, res.err)
		}
		results[res.name] = res.data
	}

	return results, nil
}
