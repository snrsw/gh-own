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

var defaultPRQueries = map[string]string{
	"drafts":           "is:pr is:open draft:true {owner}",
	"needsAction":      "is:pr is:open draft:false review:changes-requested {owner}",
	"readyToMerge":     "is:pr is:open draft:false review:approved {owner}",
	"waiting":          "is:pr is:open draft:false -review:approved -review:changes-requested {owner}",
	"participatedUser": "is:pr is:open involves:{user} -author:{user} -assignee:{user} -review-requested:{user}",
	"reviewRequested":  "is:pr is:open review-requested:{user}",
}

var defaultIssueQueries = map[string]string{
	"created":          "is:issue is:open author:{user}",
	"assigned":         "is:issue is:open assignee:{user}",
	"participatedUser": "is:issue is:open involves:{user} -author:{user} -assignee:{user}",
}

func DefaultPRKeys() map[string]bool {
	keys := make(map[string]bool, len(defaultPRQueries))
	for k := range defaultPRQueries {
		keys[k] = true
	}
	return keys
}

func DefaultPRQueries() map[string]string {
	return copyMap(defaultPRQueries)
}

func DefaultIssueKeys() map[string]bool {
	keys := make(map[string]bool, len(defaultIssueQueries))
	for k := range defaultIssueQueries {
		keys[k] = true
	}
	return keys
}

func DefaultIssueQueries() map[string]string {
	return copyMap(defaultIssueQueries)
}

func copyMap(m map[string]string) map[string]string {
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func MergePRQueries(override map[string]string) map[string]string {
	return mergeQueries(defaultPRQueries, override)
}

func MergeIssueQueries(override map[string]string) map[string]string {
	return mergeQueries(defaultIssueQueries, override)
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

// ownerAssignedSuffix is appended to a query key when an {owner} query is
// expanded into its assignee variant. The suffixed key keeps the original key
// as a prefix so it still routes into the same tab bucket (see parsePRSearchResult).
const ownerAssignedSuffix = "Assigned"

// ResolveQueries substitutes placeholders with the authenticated username.
//
// {user} is replaced verbatim. {owner} is structural: it expands a single query
// into two — one matching author:{user} and one matching assignee:{user} — since
// GitHub search cannot match author OR assignee in a single query. Both expanded
// queries share the original key as a prefix so they merge into the same tab.
func ResolveQueries(queries map[string]string, username string) map[string]string {
	resolved := make(map[string]string, len(queries))
	for key, query := range queries {
		q := strings.ReplaceAll(query, "{user}", username)
		if strings.Contains(q, "{owner}") {
			resolved[key] = strings.ReplaceAll(q, "{owner}", "author:"+username)
			resolved[key+ownerAssignedSuffix] = strings.ReplaceAll(q, "{owner}", "assignee:"+username)
			continue
		}
		resolved[key] = q
	}
	return resolved
}

// PRSearchEntries builds the search queries for the pr command: defaults merged
// with the user's overrides, placeholders resolved, scoped to org when given,
// and ordered newest first.
func PRSearchEntries(override map[string]string, username, org string) map[string]string {
	return searchEntries(MergePRQueries(override), username, org)
}

// IssueSearchEntries builds the search queries for the issue command. See
// PRSearchEntries.
func IssueSearchEntries(override map[string]string, username, org string) map[string]string {
	return searchEntries(MergeIssueQueries(override), username, org)
}

func searchEntries(queries map[string]string, username, org string) map[string]string {
	return EnsureSort(AppendOrg(ResolveQueries(queries, username), org))
}

// updatedDescQualifier orders search results by last update, newest first.
// Queries are capped at a fixed page size, so without it the window GitHub
// returns is the most *relevant* results rather than the most recent ones.
const updatedDescQualifier = "sort:updated-desc"

// EnsureSort appends the newest-first qualifier to every query that does not
// already carry a sort of its own, so a user-supplied "sort:" always wins.
func EnsureSort(queries map[string]string) map[string]string {
	result := make(map[string]string, len(queries))
	for key, query := range queries {
		if hasSortQualifier(query) {
			result[key] = query
			continue
		}
		result[key] = query + " " + updatedDescQualifier
	}
	return result
}

func hasSortQualifier(query string) bool {
	for _, field := range strings.Fields(query) {
		if strings.HasPrefix(field, "sort:") {
			return true
		}
	}
	return false
}

func AppendOrg(queries map[string]string, org string) map[string]string {
	if org == "" {
		return queries
	}
	result := make(map[string]string, len(queries))
	for key, query := range queries {
		result[key] = query + " org:" + org
	}
	return result
}
