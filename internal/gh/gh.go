// Package gh provides GitHub API client helpers.
package gh

import (
	"fmt"

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
