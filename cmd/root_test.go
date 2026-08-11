package cmd

import (
	"slices"
	"testing"

	"github.com/snrsw/gh-own/internal/config"
)

// setFlags points the persistent flag variables at the given values for the
// duration of one test, restoring them afterwards.
func setFlags(t *testing.T, wantOrg string, wantExcludeAuthors []string, wantNoBots bool) {
	t.Helper()
	prevOrg, prevAuthors, prevNoBots := org, excludeAuthors, noBots
	t.Cleanup(func() {
		org, excludeAuthors, noBots = prevOrg, prevAuthors, prevNoBots
	})
	org, excludeAuthors, noBots = wantOrg, wantExcludeAuthors, wantNoBots
}

func TestSearchFilters_NoFlagsNoConfig(t *testing.T) {
	setFlags(t, "", nil, false)

	got := searchFilters(config.Config{})

	if got.Org != "" {
		t.Errorf("Org = %q, want empty", got.Org)
	}
	if len(got.ExcludeAuthors) != 0 {
		t.Errorf("ExcludeAuthors = %v, want empty", got.ExcludeAuthors)
	}
}

func TestSearchFilters_CombinesFlagsAndConfig(t *testing.T) {
	setFlags(t, "my-org", []string{"noisy-bot"}, true)

	got := searchFilters(config.Config{
		Exclude: config.ExcludeConfig{Authors: []string{"my-release-bot"}},
	})

	if got.Org != "my-org" {
		t.Errorf("Org = %q, want %q", got.Org, "my-org")
	}
	want := append(config.DefaultBotAuthors(), "my-release-bot", "noisy-bot")
	if !slices.Equal(got.ExcludeAuthors, want) {
		t.Errorf("ExcludeAuthors = %v, want %v", got.ExcludeAuthors, want)
	}
}

func TestSearchFilters_ConfigBotsWithoutFlag(t *testing.T) {
	setFlags(t, "", nil, false)

	got := searchFilters(config.Config{Exclude: config.ExcludeConfig{Bots: true}})

	if !slices.Equal(got.ExcludeAuthors, config.DefaultBotAuthors()) {
		t.Errorf("ExcludeAuthors = %v, want %v", got.ExcludeAuthors, config.DefaultBotAuthors())
	}
}

func TestPersistentFlagsRegistered(t *testing.T) {
	for _, name := range []string{"org", "exclude-author", "no-bots", "demo", "debug"} {
		if rootCmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("persistent flag %q is not registered", name)
		}
	}
}
