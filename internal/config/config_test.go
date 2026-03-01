package config

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultPRQueries_ContainsExpectedKeys(t *testing.T) {
	expectedKeys := []string{"created", "assigned", "participatedUser", "reviewRequested"}

	if len(DefaultPRQueries()) != len(expectedKeys) {
		t.Fatalf("DefaultPRQueries() has %d keys, want %d", len(DefaultPRQueries()), len(expectedKeys))
	}

	for _, key := range expectedKeys {
		query, ok := DefaultPRQueries()[key]
		if !ok {
			t.Errorf("DefaultPRQueries() missing key %q", key)
			continue
		}
		if !strings.Contains(query, "{user}") {
			t.Errorf("DefaultPRQueries()[%q] = %q, want it to contain {user}", key, query)
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
		"created": "is:pr is:open author:{user} label:custom",
	}

	merged := MergePRQueries(override)

	if len(merged) != len(DefaultPRQueries()) {
		t.Fatalf("merged has %d keys, want %d", len(merged), len(DefaultPRQueries()))
	}

	if got := merged["created"]; got != override["created"] {
		t.Errorf("merged[created] = %q, want %q", got, override["created"])
	}

	for _, key := range []string{"assigned", "participatedUser", "reviewRequested"} {
		if got := merged[key]; got != DefaultPRQueries()[key] {
			t.Errorf("merged[%q] = %q, want default %q", key, got, DefaultPRQueries()[key])
		}
	}
}

func TestMergePRQueries_FullOverride(t *testing.T) {
	override := map[string]string{
		"created":          "custom-created",
		"assigned":         "custom-assigned",
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
