package gh

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/config"

	"github.com/snrsw/gh-own/internal/cache"
)

type mockTransport struct {
	handler func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.handler(req)
}

func newTestRESTClient(t *testing.T, transport http.RoundTripper) *api.RESTClient {
	t.Helper()
	client, err := api.NewRESTClient(api.ClientOptions{
		AuthToken: "test-token",
		Transport: transport,
	})
	if err != nil {
		t.Fatalf("failed to create test REST client: %v", err)
	}
	return client
}

func TestParseTeamSlugs(t *testing.T) {
	teams := []teamResponse{
		{Slug: "team-a"},
		{Slug: "team-b"},
	}
	teams[0].Organization.Login = "my-org"
	teams[1].Organization.Login = "other-org"

	slugs := parseTeamSlugs(teams)

	expected := []string{"my-org/team-a", "other-org/team-b"}
	if len(slugs) != len(expected) {
		t.Fatalf("parseTeamSlugs returned %d slugs, want %d", len(slugs), len(expected))
	}
	for i, want := range expected {
		if slugs[i] != want {
			t.Errorf("slugs[%d] = %q, want %q", i, slugs[i], want)
		}
	}
}

func TestParseTeamSlugs_EmptyInput(t *testing.T) {
	slugs := parseTeamSlugs([]teamResponse{})

	if slugs == nil {
		t.Fatal("parseTeamSlugs returned nil, want empty slice")
	}
	if len(slugs) != 0 {
		t.Errorf("parseTeamSlugs returned %d slugs, want 0", len(slugs))
	}
}

func TestGetTeamSlugs_ErrorOnAPIFailure(t *testing.T) {
	transport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("network error")
		},
	}
	client := newTestRESTClient(t, transport)

	_, err := GetTeamSlugs(client)
	if err == nil {
		t.Fatal("expected error on API failure, got nil")
	}
}

func TestSearchResult_GenericTypes(t *testing.T) {
	// Verify the generic SearchResult works with custom types
	type TestItem struct {
		ID   int
		Name string
	}

	result := SearchResult[TestItem]{
		TotalCount: 2,
		Items: []TestItem{
			{ID: 1, Name: "first"},
			{ID: 2, Name: "second"},
		},
	}

	if result.TotalCount != 2 {
		t.Errorf("SearchResult.TotalCount = %d, want %d", result.TotalCount, 2)
	}

	if len(result.Items) != 2 {
		t.Errorf("len(SearchResult.Items) = %d, want %d", len(result.Items), 2)
	}

	if result.Items[0].ID != 1 || result.Items[0].Name != "first" {
		t.Errorf("SearchResult.Items[0] = %+v, want {ID: 1, Name: first}", result.Items[0])
	}
}

func TestLoginFromConfig_ErrorWhenUserMissing(t *testing.T) {
	cfg := config.ReadFromString("hosts:\n  github.com:\n    git_protocol: ssh\n")
	_, err := loginFromConfig(cfg, "github.com")
	if err == nil {
		t.Fatal("expected error when user key is missing, got nil")
	}
}

func TestLoginFromConfig_ReadsUsername(t *testing.T) {
	cfg := config.ReadFromString("hosts:\n  github.com:\n    user: testuser\n")
	login, err := loginFromConfig(cfg, "github.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if login != "testuser" {
		t.Errorf("loginFromConfig() = %q, want %q", login, "testuser")
	}
}

func TestGetTeamSlugsWithCache_CacheHit(t *testing.T) {
	tmpDir := t.TempDir()
	store := cache.NewStoreWithPath(filepath.Join(tmpDir, "teams.json"))

	// Write cache with teams
	expectedTeams := []string{"org/team-a", "org/team-b"}
	if err := store.WriteTeams(expectedTeams); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Create client that should NOT be called
	apiCallCount := 0
	transport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			apiCallCount++
			t.Error("API should not be called when cache is hit")
			return nil, fmt.Errorf("unexpected API call")
		},
	}
	client := newTestRESTClient(t, transport)

	// Call with cache enabled (1 hour TTL)
	teams, err := GetTeamSlugsWithCache(client, store, 1*time.Hour)
	if err != nil {
		t.Fatalf("GetTeamSlugsWithCache() error: %v", err)
	}

	if apiCallCount != 0 {
		t.Errorf("API was called %d times, want 0 (cache hit)", apiCallCount)
	}

	if len(teams) != 2 {
		t.Errorf("got %d teams, want 2", len(teams))
	}
	if teams[0] != "org/team-a" || teams[1] != "org/team-b" {
		t.Errorf("got teams %v, want %v", teams, expectedTeams)
	}
}

func TestGetTeamSlugsWithCache_CacheMiss(t *testing.T) {
	tmpDir := t.TempDir()
	store := cache.NewStoreWithPath(filepath.Join(tmpDir, "teams.json"))

	// Create client that returns teams
	apiCallCount := 0
	transport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			apiCallCount++
			if !strings.HasSuffix(req.URL.Path, "/user/teams") {
				t.Errorf("unexpected request path: %s", req.URL.Path)
			}

			teams := []teamResponse{
				{Slug: "team-x"},
				{Slug: "team-y"},
			}
			teams[0].Organization.Login = "my-org"
			teams[1].Organization.Login = "my-org"

			body, err := json.Marshal(teams)
			if err != nil {
				return nil, err
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(body))),
				Header:     make(http.Header),
			}, nil
		},
	}
	client := newTestRESTClient(t, transport)

	// Call with cache enabled - should hit API since cache is empty
	teams, err := GetTeamSlugsWithCache(client, store, 1*time.Hour)
	if err != nil {
		t.Fatalf("GetTeamSlugsWithCache() error: %v", err)
	}

	if apiCallCount != 1 {
		t.Errorf("API was called %d times, want 1 (cache miss)", apiCallCount)
	}

	expectedTeams := []string{"my-org/team-x", "my-org/team-y"}
	if len(teams) != 2 {
		t.Errorf("got %d teams, want 2", len(teams))
	} else if teams[0] != expectedTeams[0] || teams[1] != expectedTeams[1] { //nolint:gosec
		t.Errorf("got teams %v, want %v", teams, expectedTeams)
	}

	// Verify cache was written
	cachedTeams, err := store.ReadTeams(1 * time.Hour)
	if err != nil {
		t.Fatalf("failed to read cache after write: %v", err)
	}
	if cachedTeams == nil {
		t.Fatal("cache was not written")
	}
	if len(cachedTeams) != 2 {
		t.Errorf("cached %d teams, want 2", len(cachedTeams))
	}
}

func TestGetTeamSlugsWithCache_CacheExpired(t *testing.T) {
	tmpDir := t.TempDir()
	store := cache.NewStoreWithPath(filepath.Join(tmpDir, "teams.json"))

	// Write cache with old timestamp
	oldTeams := []string{"org/old-team"}
	if err := store.WriteTeams(oldTeams); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Manually update cache timestamp to be expired
	cachePath := filepath.Join(tmpDir, "teams.json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache: %v", err)
	}
	var teamCache cache.TeamCache
	err = json.Unmarshal(data, &teamCache)
	if err != nil {
		t.Fatalf("failed to unmarshal cache: %v", err)
	}
	teamCache.CachedAt = time.Now().Add(-7 * time.Hour) // Older than 6 hour TTL
	data, err = json.Marshal(teamCache)
	if err != nil {
		t.Fatalf("failed to marshal cache: %v", err)
	}
	err = os.WriteFile(cachePath, data, 0600)
	if err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Create client that returns new teams
	apiCallCount := 0
	transport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			apiCallCount++
			teams := []teamResponse{
				{Slug: "new-team"},
			}
			teams[0].Organization.Login = "new-org"

			body, marshalErr := json.Marshal(teams)
			if marshalErr != nil {
				return nil, marshalErr
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(body))),
				Header:     make(http.Header),
			}, nil
		},
	}
	client := newTestRESTClient(t, transport)

	// Call with 6 hour TTL - should hit API since cache is expired
	teams, err := GetTeamSlugsWithCache(client, store, 6*time.Hour)
	if err != nil {
		t.Fatalf("GetTeamSlugsWithCache() error: %v", err)
	}

	if apiCallCount != 1 {
		t.Errorf("API was called %d times, want 1 (cache expired)", apiCallCount)
	}

	if len(teams) != 1 || teams[0] != "new-org/new-team" {
		t.Errorf("got teams %v, want [new-org/new-team]", teams)
	}
}
