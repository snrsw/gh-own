package gh

import (
	"strings"
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

func TestSearchIssuesGraphQL_EmptyUsername(t *testing.T) {
	results, err := SearchIssuesGraphQL(nil, "")

	if err != nil {
		t.Errorf("SearchIssuesGraphQL with empty username returned error: %v", err)
	}
	if results == nil {
		t.Error("SearchIssuesGraphQL returned nil map")
	}
	if len(results) != 0 {
		t.Errorf("SearchIssuesGraphQL with empty username returned %d results, want 0", len(results))
	}
}

func TestBuildIssueSearchVariables(t *testing.T) {
	username := "testuser"
	vars := buildIssueSearchVariables(username)

	expectedCreated := "is:issue is:open author:testuser"
	if vars["created"] != expectedCreated {
		t.Errorf("created = %q, want %q", vars["created"], expectedCreated)
	}

	expectedAssigned := "is:issue is:open assignee:testuser"
	if vars["assigned"] != expectedAssigned {
		t.Errorf("assigned = %q, want %q", vars["assigned"], expectedAssigned)
	}

	expectedParticipated := "is:issue is:open involves:testuser -author:testuser -assignee:testuser"
	if vars["participated"] != expectedParticipated {
		t.Errorf("participated = %q, want %q", vars["participated"], expectedParticipated)
	}
}

func TestIssueSearchQuery_ContainsAliases(t *testing.T) {
	query := issueSearchQuery

	if query == "" {
		t.Fatal("issueSearchQuery is empty")
	}

	requiredParts := []string{
		"created: search",
		"assigned: search",
		"participated: search",
		"... on Issue",
	}

	for _, part := range requiredParts {
		if !strings.Contains(query, part) {
			t.Errorf("issueSearchQuery should contain %q", part)
		}
	}
}

func TestParseIssueSearchNodes(t *testing.T) {
	node1 := issueSearchRawNode{
		Number:    1,
		Title:     "Issue One",
		URL:       "https://github.com/owner/repo/issues/1",
		State:     "OPEN",
		UpdatedAt: "2024-03-15T10:00:00Z",
		CreatedAt: "2024-03-10T08:00:00Z",
	}
	node1.Author.Login = "user1"
	node1.Repository.NameWithOwner = "owner/repo"

	node2 := issueSearchRawNode{
		Number:    2,
		Title:     "Issue Two",
		URL:       "https://github.com/owner/repo/issues/2",
		State:     "OPEN",
		UpdatedAt: "2024-03-16T10:00:00Z",
		CreatedAt: "2024-03-11T08:00:00Z",
	}
	node2.Author.Login = "user2"
	node2.Repository.NameWithOwner = "owner/repo"

	rawNodes := []issueSearchRawNode{
		node1,
		{Number: 0}, // Should be skipped
		node2,
	}

	nodes := parseIssueSearchNodes(rawNodes)

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	if nodes[0].Number != 1 {
		t.Errorf("nodes[0].Number = %d, want 1", nodes[0].Number)
	}
	if nodes[0].Title != "Issue One" {
		t.Errorf("nodes[0].Title = %q, want %q", nodes[0].Title, "Issue One")
	}
	if nodes[0].Author.Login != "user1" {
		t.Errorf("nodes[0].Author.Login = %q, want %q", nodes[0].Author.Login, "user1")
	}

	if nodes[1].Number != 2 {
		t.Errorf("nodes[1].Number = %d, want 2", nodes[1].Number)
	}
}
