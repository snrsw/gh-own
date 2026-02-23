package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PR    CommandConfig `yaml:"pr"`
	Issue CommandConfig `yaml:"issue"`
}

type CommandConfig struct {
	Queries map[string]string `yaml:"queries"`
}

func DefaultPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "gh-own", "config.yaml")
}

func LoadFromPath(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	if cfg.PR.Queries != nil {
		cfg.PR.Queries = NormalizeKeys(cfg.PR.Queries)
	}
	if cfg.Issue.Queries != nil {
		cfg.Issue.Queries = NormalizeKeys(cfg.Issue.Queries)
	}

	return cfg, nil
}

var DefaultPRQueries = map[string]string{
	"created":          "is:pr is:open author:{user}",
	"assigned":         "is:pr is:open assignee:{user}",
	"participatedUser": "is:pr is:open involves:{user} -author:{user} -assignee:{user} -review-requested:{user}",
	"reviewRequested":  "is:pr is:open review-requested:{user}",
}

var DefaultIssueQueries = map[string]string{
	"created":          "is:issue is:open author:{user}",
	"assigned":         "is:issue is:open assignee:{user}",
	"participatedUser": "is:issue is:open involves:{user} -author:{user} -assignee:{user}",
}

func MergePRQueries(override map[string]string) map[string]string {
	return mergeQueries(DefaultPRQueries, override)
}

func MergeIssueQueries(override map[string]string) map[string]string {
	return mergeQueries(DefaultIssueQueries, override)
}

func mergeQueries(defaults, override map[string]string) map[string]string {
	merged := make(map[string]string, len(defaults))
	for k, v := range defaults {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}

var keyAliases = map[string]string{
	"participated":     "participatedUser",
	"review_requested": "reviewRequested",
}

func NormalizeKeys(queries map[string]string) map[string]string {
	normalized := make(map[string]string, len(queries))
	for k, v := range queries {
		if alias, ok := keyAliases[k]; ok {
			normalized[alias] = v
		} else {
			normalized[k] = v
		}
	}
	return normalized
}

func ResolveQueries(queries map[string]string, username string) map[string]string {
	resolved := make(map[string]string, len(queries))
	for key, query := range queries {
		resolved[key] = strings.ReplaceAll(query, "{user}", username)
	}
	return resolved
}
