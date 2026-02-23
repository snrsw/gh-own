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
