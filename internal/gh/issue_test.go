package gh

import (
	"testing"
)

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

func TestSearchIssues_EmptyUsernameWithTeams(t *testing.T) {
	results, err := SearchIssuesTeams(nil, "", []string{"my-org/team-a"})

	if err != nil {
		t.Errorf("SearchIssues with empty username returned error: %v", err)
	}

	if len(results.Created) != 0 {
		t.Errorf("SearchIssues with empty username returned %d created, want 0", len(results.Created))
	}
	if len(results.Participated) != 0 {
		t.Errorf("SearchIssues with empty username returned %d participated, want 0", len(results.Participated))
	}
}

func TestSearchIssues_EmptyUsername(t *testing.T) {
	results, err := SearchIssues(nil, "")

	if err != nil {
		t.Errorf("SearchIssues with empty username returned error: %v", err)
	}

	if len(results.Assigned) != 0 {
		t.Errorf("SearchIssues with empty username returned %d results, want 0", len(results.Assigned))
	}
	if len(results.Created) != 0 {
		t.Errorf("SearchIssues with empty username returned %d results, want 0", len(results.Created))
	}
	if len(results.Participated) != 0 {
		t.Errorf("SearchIssues with empty username returned %d results, want 0", len(results.Participated))
	}
}

func TestParseIssueSearchNodes_NoActivity(t *testing.T) {
	node := issueSearchRawNode{Number: 1, Title: "Test"}

	nodes := parseIssueSearchNodes([]issueSearchRawNode{node})

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LatestActivity.Login != "" {
		t.Errorf("Login = %q, want empty", nodes[0].LatestActivity.Login)
	}
	if nodes[0].LatestActivity.Kind != "" {
		t.Errorf("Kind = %q, want empty", nodes[0].LatestActivity.Kind)
	}
}

func TestParseIssueSearchNodes_WithComment(t *testing.T) {
	node := issueSearchRawNode{Number: 1, Title: "Test"}
	node.Comments.Nodes = []struct {
		Author    struct{ Login string `json:"login"` } `json:"author"`
		CreatedAt string                                `json:"createdAt"`
	}{{CreatedAt: "2024-03-10T12:00:00Z"}}
	node.Comments.Nodes[0].Author.Login = "alice"

	nodes := parseIssueSearchNodes([]issueSearchRawNode{node})

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LatestActivity.Kind != "commented" {
		t.Errorf("Kind = %q, want %q", nodes[0].LatestActivity.Kind, "commented")
	}
	if nodes[0].LatestActivity.Login != "alice" {
		t.Errorf("Login = %q, want %q", nodes[0].LatestActivity.Login, "alice")
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
