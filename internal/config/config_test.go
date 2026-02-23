package config

import (
	"strings"
	"testing"
)

func TestDefaultPRQueries_ContainsExpectedKeys(t *testing.T) {
	expectedKeys := []string{"created", "assigned", "participatedUser", "reviewRequested"}

	if len(DefaultPRQueries) != len(expectedKeys) {
		t.Fatalf("DefaultPRQueries has %d keys, want %d", len(DefaultPRQueries), len(expectedKeys))
	}

	for _, key := range expectedKeys {
		query, ok := DefaultPRQueries[key]
		if !ok {
			t.Errorf("DefaultPRQueries missing key %q", key)
			continue
		}
		if !strings.Contains(query, "{user}") {
			t.Errorf("DefaultPRQueries[%q] = %q, want it to contain {user}", key, query)
		}
	}
}

func TestDefaultIssueQueries_ContainsExpectedKeys(t *testing.T) {
	expectedKeys := []string{"created", "assigned", "participatedUser"}

	if len(DefaultIssueQueries) != len(expectedKeys) {
		t.Fatalf("DefaultIssueQueries has %d keys, want %d", len(DefaultIssueQueries), len(expectedKeys))
	}

	for _, key := range expectedKeys {
		query, ok := DefaultIssueQueries[key]
		if !ok {
			t.Errorf("DefaultIssueQueries missing key %q", key)
			continue
		}
		if !strings.Contains(query, "{user}") {
			t.Errorf("DefaultIssueQueries[%q] = %q, want it to contain {user}", key, query)
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

	if len(merged) != len(DefaultPRQueries) {
		t.Fatalf("MergePRQueries(nil) has %d keys, want %d", len(merged), len(DefaultPRQueries))
	}

	for key, want := range DefaultPRQueries {
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

	if len(merged) != len(DefaultPRQueries) {
		t.Fatalf("merged has %d keys, want %d", len(merged), len(DefaultPRQueries))
	}

	if got := merged["created"]; got != override["created"] {
		t.Errorf("merged[created] = %q, want %q", got, override["created"])
	}

	for _, key := range []string{"assigned", "participatedUser", "reviewRequested"} {
		if got := merged[key]; got != DefaultPRQueries[key] {
			t.Errorf("merged[%q] = %q, want default %q", key, got, DefaultPRQueries[key])
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

	if len(merged) != len(DefaultIssueQueries) {
		t.Fatalf("MergeIssueQueries(nil) has %d keys, want %d", len(merged), len(DefaultIssueQueries))
	}

	for key, want := range DefaultIssueQueries {
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

	if len(merged) != len(DefaultIssueQueries) {
		t.Fatalf("merged has %d keys, want %d", len(merged), len(DefaultIssueQueries))
	}

	if got := merged["assigned"]; got != override["assigned"] {
		t.Errorf("merged[assigned] = %q, want %q", got, override["assigned"])
	}

	for _, key := range []string{"created", "participatedUser"} {
		if got := merged[key]; got != DefaultIssueQueries[key] {
			t.Errorf("merged[%q] = %q, want default %q", key, got, DefaultIssueQueries[key])
		}
	}
}
