package gh

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/config"
)

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
