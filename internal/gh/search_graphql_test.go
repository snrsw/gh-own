package gh

import (
	"testing"

	"github.com/snrsw/gh-own/internal/cistatus"
)

func TestPRSearchResult_CIStatus(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		expected cistatus.CIStatus
	}{
		{"success", "SUCCESS", cistatus.CIStatusSuccess},
		{"failure", "FAILURE", cistatus.CIStatusFailure},
		{"pending", "PENDING", cistatus.CIStatusPending},
		{"none", "", cistatus.CIStatusNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := PRSearchNode{
				StatusState: tt.state,
			}
			if got := pr.CIStatus(); got != tt.expected {
				t.Errorf("CIStatus() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPRSearchNode_RepositoryURL(t *testing.T) {
	pr := PRSearchNode{
		Repository: struct {
			NameWithOwner string
		}{
			NameWithOwner: "owner/repo",
		},
	}

	expected := "https://api.github.com/repos/owner/repo"
	if got := pr.RepositoryURL(); got != expected {
		t.Errorf("RepositoryURL() = %q, want %q", got, expected)
	}
}

func TestBuildSearchQuery(t *testing.T) {
	query := BuildSearchQuery("is:pr is:open author:user")

	if query == "" {
		t.Error("BuildSearchQuery returned empty string")
	}
}

func TestSearchGraphQL_EmptyConditions(t *testing.T) {
	results, err := SearchGraphQL(nil, []Condition{})

	if err != nil {
		t.Errorf("SearchGraphQL with empty conditions returned error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("SearchGraphQL with empty conditions returned %d results, want 0", len(results))
	}
}

func TestIssueSearchNode_RepositoryURL(t *testing.T) {
	issue := IssueSearchNode{
		Repository: struct {
			NameWithOwner string
		}{
			NameWithOwner: "owner/repo",
		},
	}

	expected := "https://api.github.com/repos/owner/repo"
	if got := issue.RepositoryURL(); got != expected {
		t.Errorf("RepositoryURL() = %q, want %q", got, expected)
	}
}

func TestIssueSearchNode_Fields(t *testing.T) {
	issue := IssueSearchNode{
		Number:    42,
		Title:     "Test Issue",
		URL:       "https://github.com/owner/repo/issues/42",
		State:     "OPEN",
		UpdatedAt: "2024-03-15T10:00:00Z",
		CreatedAt: "2024-03-10T08:00:00Z",
	}
	issue.Author.Login = "testuser"
	issue.Repository.NameWithOwner = "owner/repo"

	if issue.Number != 42 {
		t.Errorf("Number = %d, want 42", issue.Number)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Title = %q, want %q", issue.Title, "Test Issue")
	}
	if issue.URL != "https://github.com/owner/repo/issues/42" {
		t.Errorf("URL = %q, want %q", issue.URL, "https://github.com/owner/repo/issues/42")
	}
	if issue.State != "OPEN" {
		t.Errorf("State = %q, want %q", issue.State, "OPEN")
	}
	if issue.Author.Login != "testuser" {
		t.Errorf("Author.Login = %q, want %q", issue.Author.Login, "testuser")
	}
	if issue.Repository.NameWithOwner != "owner/repo" {
		t.Errorf("Repository.NameWithOwner = %q, want %q", issue.Repository.NameWithOwner, "owner/repo")
	}
}
