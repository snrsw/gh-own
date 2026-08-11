package config

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultPRKeys_ReturnsKnownKeys(t *testing.T) {
	want := map[string]bool{
		"drafts":           true,
		"needsAction":      true,
		"readyToMerge":     true,
		"waiting":          true,
		"participatedUser": true,
		"reviewRequested":  true,
	}

	got := DefaultPRKeys()

	if len(got) != len(want) {
		t.Fatalf("DefaultPRKeys() has %d keys, want %d", len(got), len(want))
	}
	for key := range want {
		if !got[key] {
			t.Errorf("DefaultPRKeys() missing key %q", key)
		}
	}
}

func TestDefaultPRQueries_ContainsExpectedKeys(t *testing.T) {
	expectedKeys := []string{
		"drafts", "needsAction", "readyToMerge", "waiting",
		"participatedUser", "reviewRequested",
	}

	if len(DefaultPRQueries()) != len(expectedKeys) {
		t.Fatalf("DefaultPRQueries() has %d keys, want %d", len(DefaultPRQueries()), len(expectedKeys))
	}

	for _, key := range expectedKeys {
		query, ok := DefaultPRQueries()[key]
		if !ok {
			t.Errorf("DefaultPRQueries() missing key %q", key)
			continue
		}
		// State tabs key on {owner}; relationship tabs key on {user}.
		if !strings.Contains(query, "{user}") && !strings.Contains(query, "{owner}") {
			t.Errorf("DefaultPRQueries()[%q] = %q, want it to contain {user} or {owner}", key, query)
		}
	}
}

func TestDefaultIssueKeys_ReturnsKnownKeys(t *testing.T) {
	want := map[string]bool{
		"created":          true,
		"assigned":         true,
		"participatedUser": true,
	}

	got := DefaultIssueKeys()

	if len(got) != len(want) {
		t.Fatalf("DefaultIssueKeys() has %d keys, want %d", len(got), len(want))
	}
	for key := range want {
		if !got[key] {
			t.Errorf("DefaultIssueKeys() missing key %q", key)
		}
	}
}

func TestDefaultIssueQueries_ContainsExpectedKeys(t *testing.T) {
	expectedKeys := []string{"created", "assigned", "participatedUser"}

	if len(DefaultIssueQueries()) != len(expectedKeys) {
		t.Fatalf("DefaultIssueQueries() has %d keys, want %d", len(DefaultIssueQueries()), len(expectedKeys))
	}

	for _, key := range expectedKeys {
		query, ok := DefaultIssueQueries()[key]
		if !ok {
			t.Errorf("DefaultIssueQueries() missing key %q", key)
			continue
		}
		if !strings.Contains(query, "{user}") {
			t.Errorf("DefaultIssueQueries()[%q] = %q, want it to contain {user}", key, query)
		}
	}
}

func TestResolveQueries_ReplacesUserPlaceholder(t *testing.T) {
	queries := map[string]string{
		"created": "is:pr is:open author:{user}",
	}

	resolved := ResolveQueries(queries, "octocat")

	want := "is:pr is:open author:octocat"
	if got := resolved["created"]; got != want {
		t.Errorf("resolved[created] = %q, want %q", got, want)
	}
}

func TestResolveQueries_MultipleOccurrences(t *testing.T) {
	queries := map[string]string{
		"participated": "is:pr is:open involves:{user} -author:{user}",
	}

	resolved := ResolveQueries(queries, "octocat")

	want := "is:pr is:open involves:octocat -author:octocat"
	if got := resolved["participated"]; got != want {
		t.Errorf("resolved[participated] = %q, want %q", got, want)
	}
}

func TestResolveQueries_NoPlaceholder(t *testing.T) {
	queries := map[string]string{
		"custom": "is:pr is:open label:bug",
	}

	resolved := ResolveQueries(queries, "octocat")

	want := "is:pr is:open label:bug"
	if got := resolved["custom"]; got != want {
		t.Errorf("resolved[custom] = %q, want %q", got, want)
	}
}

func TestResolveQueries_OwnerExpandsToAuthorAndAssigned(t *testing.T) {
	queries := map[string]string{
		"drafts": "is:pr is:open draft:true {owner}",
	}

	resolved := ResolveQueries(queries, "octocat")

	if len(resolved) != 2 {
		t.Fatalf("ResolveQueries expanded {owner} into %d entries, want 2", len(resolved))
	}

	wantAuthor := "is:pr is:open draft:true author:octocat"
	if got := resolved["drafts"]; got != wantAuthor {
		t.Errorf("resolved[drafts] = %q, want %q", got, wantAuthor)
	}

	wantAssigned := "is:pr is:open draft:true assignee:octocat"
	if got := resolved["drafts"+ownerAssignedSuffix]; got != wantAssigned {
		t.Errorf("resolved[draftsAssigned] = %q, want %q", got, wantAssigned)
	}
}

func TestMergePRQueries_NilOverride_ReturnsDefaults(t *testing.T) {
	merged := MergePRQueries(nil)

	if len(merged) != len(DefaultPRQueries()) {
		t.Fatalf("MergePRQueries(nil) has %d keys, want %d", len(merged), len(DefaultPRQueries()))
	}

	for key, want := range DefaultPRQueries() {
		if got := merged[key]; got != want {
			t.Errorf("merged[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestMergePRQueries_PartialOverride(t *testing.T) {
	override := map[string]string{
		"reviewRequested": "is:pr is:open review-requested:{user} label:custom",
	}

	merged := MergePRQueries(override)

	if len(merged) != len(DefaultPRQueries()) {
		t.Fatalf("merged has %d keys, want %d", len(merged), len(DefaultPRQueries()))
	}

	if got := merged["reviewRequested"]; got != override["reviewRequested"] {
		t.Errorf("merged[reviewRequested] = %q, want %q", got, override["reviewRequested"])
	}

	for _, key := range []string{"drafts", "needsAction", "readyToMerge", "waiting", "participatedUser"} {
		if got := merged[key]; got != DefaultPRQueries()[key] {
			t.Errorf("merged[%q] = %q, want default %q", key, got, DefaultPRQueries()[key])
		}
	}
}

func TestMergePRQueries_FullOverride(t *testing.T) {
	override := map[string]string{
		"drafts":           "custom-drafts",
		"needsAction":      "custom-needs-action",
		"participatedUser": "custom-participated",
		"reviewRequested":  "custom-review",
	}

	merged := MergePRQueries(override)

	for key, want := range override {
		if got := merged[key]; got != want {
			t.Errorf("merged[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestMergeIssueQueries_NilOverride_ReturnsDefaults(t *testing.T) {
	merged := MergeIssueQueries(nil)

	if len(merged) != len(DefaultIssueQueries()) {
		t.Fatalf("MergeIssueQueries(nil) has %d keys, want %d", len(merged), len(DefaultIssueQueries()))
	}

	for key, want := range DefaultIssueQueries() {
		if got := merged[key]; got != want {
			t.Errorf("merged[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestMergeIssueQueries_PartialOverride(t *testing.T) {
	override := map[string]string{
		"assigned": "is:issue is:open assignee:{user} label:custom",
	}

	merged := MergeIssueQueries(override)

	if len(merged) != len(DefaultIssueQueries()) {
		t.Fatalf("merged has %d keys, want %d", len(merged), len(DefaultIssueQueries()))
	}

	if got := merged["assigned"]; got != override["assigned"] {
		t.Errorf("merged[assigned] = %q, want %q", got, override["assigned"])
	}

	for _, key := range []string{"created", "participatedUser"} {
		if got := merged[key]; got != DefaultIssueQueries()[key] {
			t.Errorf("merged[%q] = %q, want default %q", key, got, DefaultIssueQueries()[key])
		}
	}
}

func TestMergePRQueries_NewKeyOverride(t *testing.T) {
	override := map[string]string{
		"customTab": "is:pr is:open label:needs-triage",
	}

	merged := MergePRQueries(override)

	wantLen := len(DefaultPRQueries()) + 1
	if len(merged) != wantLen {
		t.Fatalf("merged has %d keys, want %d", len(merged), wantLen)
	}

	if got := merged["customTab"]; got != override["customTab"] {
		t.Errorf("merged[customTab] = %q, want %q", got, override["customTab"])
	}

	for key, want := range DefaultPRQueries() {
		if got := merged[key]; got != want {
			t.Errorf("merged[%q] = %q, want default %q", key, got, want)
		}
	}
}

func TestNormalizeKeys(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"participated", "participatedUser"},
		{"review_requested", "reviewRequested"},
		{"created", "created"},
		{"assigned", "assigned"},
		{"participatedUser", "participatedUser"},
		{"reviewRequested", "reviewRequested"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			input := map[string]string{tt.input: "some query"}
			normalized := NormalizeKeys(input)

			if _, ok := normalized[tt.want]; !ok {
				t.Errorf("NormalizeKeys(%q) missing key %q, got keys: %v", tt.input, tt.want, keys(normalized))
			}
		})
	}
}

func keys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func TestLoadFromPath_FileNotFound_ReturnsEmptyConfig(t *testing.T) {
	cfg, err := LoadFromPath("/nonexistent/path/config.yaml")

	if err != nil {
		t.Fatalf("LoadFromPath returned error: %v", err)
	}

	if cfg.PR.Queries != nil {
		t.Errorf("PR.Queries = %v, want nil", cfg.PR.Queries)
	}
	if cfg.Issue.Queries != nil {
		t.Errorf("Issue.Queries = %v, want nil", cfg.Issue.Queries)
	}
}

func TestLoadFromPath_ValidYAML_ParsesPRQueries(t *testing.T) {
	content := `
pr:
  queries:
    created: "is:pr is:open author:{user} label:custom"
    review_requested: "is:pr is:open review-requested:{user} label:urgent"
`
	path := writeTempYAML(t, content)

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath returned error: %v", err)
	}

	if got := cfg.PR.Queries["created"]; got != "is:pr is:open author:{user} label:custom" {
		t.Errorf("PR.Queries[created] = %q, want custom value", got)
	}
	if got := cfg.PR.Queries["reviewRequested"]; got != "is:pr is:open review-requested:{user} label:urgent" {
		t.Errorf("PR.Queries[reviewRequested] = %q, want normalized key with custom value", got)
	}
}

func TestLoadFromPath_ValidYAML_ParsesIssueQueries(t *testing.T) {
	content := `
issue:
  queries:
    participated: "is:issue is:open involves:{user}"
`
	path := writeTempYAML(t, content)

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath returned error: %v", err)
	}

	if got := cfg.Issue.Queries["participatedUser"]; got != "is:issue is:open involves:{user}" {
		t.Errorf("Issue.Queries[participatedUser] = %q, want normalized key with custom value", got)
	}
}

func TestLoadFromPath_InvalidYAML_ReturnsError(t *testing.T) {
	path := writeTempYAML(t, "{{invalid yaml")

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatal("LoadFromPath with invalid YAML should return error")
	}
}

func TestLoadFromPath_PartialConfig_OnlyPR(t *testing.T) {
	content := `
pr:
  queries:
    created: "is:pr is:open author:{user} label:mine"
`
	path := writeTempYAML(t, content)

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath returned error: %v", err)
	}

	if cfg.PR.Queries == nil {
		t.Fatal("PR.Queries should not be nil")
	}
	if cfg.Issue.Queries != nil {
		t.Errorf("Issue.Queries = %v, want nil", cfg.Issue.Queries)
	}
}

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/config.yaml"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp YAML: %v", err)
	}
	return path
}

func TestAppendOrg_EmptyOrg_ReturnsOriginal(t *testing.T) {
	queries := map[string]string{
		"created": "is:pr is:open author:octocat",
	}

	got := AppendOrg(queries, "")

	if got["created"] != queries["created"] {
		t.Errorf("AppendOrg with empty org = %q, want %q", got["created"], queries["created"])
	}
}

func TestAppendOrg_NonEmptyOrg_AppendsQualifier(t *testing.T) {
	queries := map[string]string{
		"created":  "is:pr is:open author:octocat",
		"assigned": "is:pr is:open assignee:octocat",
	}

	got := AppendOrg(queries, "my-org")

	wantCreated := "is:pr is:open author:octocat org:my-org"
	if got["created"] != wantCreated {
		t.Errorf("AppendOrg[created] = %q, want %q", got["created"], wantCreated)
	}

	wantAssigned := "is:pr is:open assignee:octocat org:my-org"
	if got["assigned"] != wantAssigned {
		t.Errorf("AppendOrg[assigned] = %q, want %q", got["assigned"], wantAssigned)
	}
}

func TestDefaultPath_UsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")

	got := DefaultPath()
	want := "/tmp/xdg-test/gh-own/config.yaml"
	if got != want {
		t.Errorf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestDefaultPath_FallsBackToHomeDotConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	got := DefaultPath()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir returned error: %v", err)
	}
	want := home + "/.config/gh-own/config.yaml"
	if got != want {
		t.Errorf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestEnsureSort_AppendsQualifier(t *testing.T) {
	queries := map[string]string{
		"drafts":  "is:pr is:open draft:true author:me",
		"waiting": "is:pr is:open",
	}

	got := EnsureSort(queries)

	want := map[string]string{
		"drafts":  "is:pr is:open draft:true author:me sort:updated-desc",
		"waiting": "is:pr is:open sort:updated-desc",
	}
	for key, wantQuery := range want {
		if got[key] != wantQuery {
			t.Errorf("EnsureSort()[%q] = %q, want %q", key, got[key], wantQuery)
		}
	}
}

func TestEnsureSort_RespectsExistingSort(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "explicit ascending", query: "is:pr is:open sort:updated-asc"},
		{name: "unrelated sort", query: "is:pr is:open sort:reactions"},
		{name: "sort in the middle", query: "is:pr sort:comments is:open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsureSort(map[string]string{"custom": tt.query})
			if got["custom"] != tt.query {
				t.Errorf("EnsureSort()[%q] = %q, want %q", "custom", got["custom"], tt.query)
			}
		})
	}
}

func TestEnsureSort_DoesNotMatchSubstring(t *testing.T) {
	// "resort:" is a title term, not a sort qualifier.
	query := "is:pr is:open resort:foo"

	got := EnsureSort(map[string]string{"custom": query})

	want := query + " sort:updated-desc"
	if got["custom"] != want {
		t.Errorf("EnsureSort()[%q] = %q, want %q", "custom", got["custom"], want)
	}
}

func TestEnsureSort_EmptyMap(t *testing.T) {
	got := EnsureSort(map[string]string{})

	if got == nil {
		t.Fatal("EnsureSort() returned nil, want empty map")
	}
	if len(got) != 0 {
		t.Errorf("EnsureSort() has %d keys, want 0", len(got))
	}
}

func TestPRSearchEntries_AppliesFullPipeline(t *testing.T) {
	got := PRSearchEntries(map[string]string{
		"myTab": "is:pr is:open label:mine",
	}, "octocat", Filters{Org: "my-org"})

	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "{owner} expands to the author variant",
			key:  "drafts",
			want: "is:pr is:open draft:true author:octocat org:my-org sort:updated-desc",
		},
		{
			name: "{owner} expands to the assignee variant",
			key:  "draftsAssigned",
			want: "is:pr is:open draft:true assignee:octocat org:my-org sort:updated-desc",
		},
		{
			name: "{user} is substituted",
			key:  "reviewRequested",
			want: "is:pr is:open review-requested:octocat org:my-org sort:updated-desc",
		},
		{
			name: "custom tabs get the same treatment",
			key:  "myTab",
			want: "is:pr is:open label:mine org:my-org sort:updated-desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got[tt.key] != tt.want {
				t.Errorf("PRSearchEntries()[%q] = %q, want %q", tt.key, got[tt.key], tt.want)
			}
		})
	}
}

func TestPRSearchEntries_KeepsUserSort(t *testing.T) {
	got := PRSearchEntries(map[string]string{
		"waiting": "is:pr is:open sort:created-asc",
	}, "octocat", Filters{})

	want := "is:pr is:open sort:created-asc"
	if got["waiting"] != want {
		t.Errorf("PRSearchEntries()[%q] = %q, want %q", "waiting", got["waiting"], want)
	}
}

func TestIssueSearchEntries_AppliesFullPipeline(t *testing.T) {
	got := IssueSearchEntries(nil, "octocat", Filters{})

	want := "is:issue is:open author:octocat sort:updated-desc"
	if got["created"] != want {
		t.Errorf("IssueSearchEntries()[%q] = %q, want %q", "created", got["created"], want)
	}
}
