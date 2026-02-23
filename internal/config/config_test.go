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
