// Package gh provides GitHub API client helpers.
package gh

import (
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/config"
)

type User struct {
	Login string `json:"login"`
}

func CurrentLogin() (string, error) {
	cfg, err := config.Read(nil)
	if err != nil {
		return "", fmt.Errorf("failed to read gh config: %w", err)
	}
	host, _ := auth.DefaultHost()
	return loginFromConfig(cfg, host)
}

func loginFromConfig(cfg *config.Config, host string) (string, error) {
	login, err := cfg.Get([]string{"hosts", host, "user"})
	if err != nil {
		return "", fmt.Errorf("failed to get user for host %s: %w", host, err)
	}
	return login, nil
}

type SearchResult[T any] struct {
	TotalCount int `json:"total_count"`
	Items      []T `json:"items"`
}

type teamResponse struct {
	Slug         string `json:"slug"`
	Organization struct {
		Login string `json:"login"`
	} `json:"organization"`
}

func GetTeamSlugs(client *api.RESTClient) ([]string, error) {
	var teams []teamResponse
	if err := client.Get("user/teams?per_page=50", &teams); err != nil {
		return nil, fmt.Errorf("failed to fetch teams: %w", err)
	}
	return parseTeamSlugs(teams), nil
}

func parseTeamSlugs(teams []teamResponse) []string {
	slugs := make([]string, len(teams))
	for i, t := range teams {
		slugs[i] = t.Organization.Login + "/" + t.Slug
	}
	return slugs
}

func searchOne[T any](
	client *api.GraphQLClient,
	gql string,
	search string,
	parse func(json.RawMessage) ([]T, error),
) ([]T, error) {
	vars := map[string]interface{}{"q": search}

	var raw map[string]json.RawMessage
	if err := client.Do(gql, vars, &raw); err != nil {
		return nil, err
	}
	return parse(raw["result"])
}

func Search[T any](
	client *api.GraphQLClient,
	gql string,
	entries map[string]string,
	parse func(json.RawMessage) ([]T, error),
) (map[string][]T, error) {
	type result struct {
		key   string
		nodes []T
		err   error
	}

	ch := make(chan result, len(entries))
	for key, search := range entries {
		go func(key, search string) {
			nodes, err := searchOne(client, gql, search, parse)
			ch <- result{key: key, nodes: nodes, err: err}
		}(key, search)
	}

	merged := make(map[string][]T, len(entries))
	for range entries {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		merged[r.key] = r.nodes
	}
	return merged, nil
}
