package gh

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/config"
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
