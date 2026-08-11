package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PR      CommandConfig `yaml:"pr"`
	Issue   CommandConfig `yaml:"issue"`
	Exclude ExcludeConfig `yaml:"exclude"`
}

type CommandConfig struct {
	Queries map[string]string `yaml:"queries"`
}

// ExcludeConfig names the authors whose pull requests and issues are dropped
// from every tab of both commands.
type ExcludeConfig struct {
	// Bots turns on the built-in automation preset, see DefaultBotAuthors.
	Bots bool `yaml:"bots"`
	// Authors are additional logins to drop, in either the "renovate[bot]" or
	// the "app/renovate" spelling.
	Authors []string `yaml:"authors"`
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

// Filters carries the narrowing options that apply to every search query,
// whether the query came from the defaults, from the user's config, or from the
// team searches in the gh package.
type Filters struct {
	// Org restricts results to a single organization. Empty means no restriction.
	Org string
	// ExcludeAuthors drops items written by these logins. Empty means no exclusion.
	ExcludeAuthors []string
}

// PRSearchEntries builds the search queries for the pr command: defaults merged
// with the user's overrides, placeholders resolved, narrowed by filters, and
// ordered newest first.
func PRSearchEntries(override map[string]string, username string, filters Filters) map[string]string {
	return searchEntries(MergePRQueries(override), username, filters)
}

// IssueSearchEntries builds the search queries for the issue command. See
// PRSearchEntries.
func IssueSearchEntries(override map[string]string, username string, filters Filters) map[string]string {
	return searchEntries(MergeIssueQueries(override), username, filters)
}

func searchEntries(queries map[string]string, username string, filters Filters) map[string]string {
	return ApplyFilters(ResolveQueries(queries, username), filters)
}

// ApplyFilters narrows every query by the given filters and guarantees a sort
// qualifier. Both the configurable user queries and the fixed team queries in
// the gh package go through it, so a filter reaches every tab.
func ApplyFilters(queries map[string]string, filters Filters) map[string]string {
	return EnsureSort(AppendOrg(ExcludeAuthors(queries, filters.ExcludeAuthors), filters.Org))
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

// defaultBotAuthors is the preset behind "exclude.bots" and --no-bots: the
// automation accounts that open pull requests on most repositories. They use the
// "app/" spelling because all three are GitHub Apps.
var defaultBotAuthors = []string{"app/dependabot", "app/renovate", "app/github-actions"}

// DefaultBotAuthors returns the built-in automation preset.
func DefaultBotAuthors() []string {
	return slices.Clone(defaultBotAuthors)
}

// ResolveExcludeAuthors folds the three sources of exclusions — the built-in
// preset, the config file, and the command line — into one ordered list.
//
// The sources union rather than override, since every one of them is a request
// to remove noise. Logins that name the same account in different spellings
// collapse to the first one seen.
func ResolveExcludeAuthors(cfg ExcludeConfig, flagAuthors []string, noBots bool) []string {
	var sources []string
	if cfg.Bots || noBots {
		sources = append(sources, defaultBotAuthors...)
	}
	sources = append(sources, cfg.Authors...)
	sources = append(sources, flagAuthors...)

	var authors []string
	seen := make(map[string]bool, len(sources))
	for _, author := range sources {
		author = strings.TrimSpace(author)
		key := canonicalAuthor(author)
		if author == "" || seen[key] {
			continue
		}
		seen[key] = true
		authors = append(authors, author)
	}
	return authors
}

// ExcludeAuthors appends a negated author qualifier for every excluded login,
// so the items never reach us and never consume a slot in the fixed-size page
// GitHub returns.
//
// A query that already names the author — positively or negatively, in any
// spelling — is left untouched, mirroring how EnsureSort defers to a
// user-supplied "sort:". This keeps a deliberate bot tab working while a global
// exclusion is in force.
func ExcludeAuthors(queries map[string]string, authors []string) map[string]string {
	if len(authors) == 0 {
		return queries
	}
	result := make(map[string]string, len(queries))
	for key, query := range queries {
		result[key] = excludeAuthorsFrom(query, authors)
	}
	return result
}

func excludeAuthorsFrom(query string, authors []string) string {
	for _, author := range authors {
		author = strings.TrimSpace(author)
		if author == "" || namesAuthor(query, author) {
			continue
		}
		query += " -author:" + author
	}
	return query
}

// namesAuthor reports whether the query already constrains the author to the
// given login. Only whole "author:" and "-author:" fields count, so a login
// appearing under another qualifier — "review-requested:renovate[bot]", say —
// does not suppress the exclusion.
func namesAuthor(query, author string) bool {
	for _, field := range strings.Fields(query) {
		login, ok := strings.CutPrefix(field, "-author:")
		if !ok {
			login, ok = strings.CutPrefix(field, "author:")
		}
		if ok && canonicalAuthor(login) == canonicalAuthor(author) {
			return true
		}
	}
	return false
}

// canonicalAuthor reduces the spellings GitHub accepts for one bot account to a
// single form: the search qualifier "app/renovate" and the API login
// "renovate[bot]" both canonicalize to "renovate".
func canonicalAuthor(login string) string {
	login = strings.ToLower(login)
	login = strings.TrimPrefix(login, "app/")
	return strings.TrimSuffix(login, "[bot]")
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
