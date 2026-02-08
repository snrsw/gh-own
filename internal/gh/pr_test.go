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

func TestSearchPRs_EmptyUsernameWithTeams(t *testing.T) {
	results, err := SearchPRs(nil, "", []string{"my-org/team-a"})

	if err != nil {
		t.Errorf("SearchPRs with empty username returned error: %v", err)
	}

	if len(results.Created) != 0 {
		t.Errorf("SearchPRs with empty username returned %d created, want 0", len(results.Created))
	}
	if len(results.ReviewRequested) != 0 {
		t.Errorf("SearchPRs with empty username returned %d reviewRequested, want 0", len(results.ReviewRequested))
	}
}

func TestSearchPRs_EmptyUsername(t *testing.T) {
	results, err := SearchPRs(nil, "", nil)

	if err != nil {
		t.Errorf("SearchPRs with empty username returned error: %v", err)
	}

	if len(results.Assigned) != 0 {
		t.Errorf("SearchPRs with empty username returned %d results, want 0", len(results.Assigned))
	}
	if len(results.Created) != 0 {
		t.Errorf("SearchPRs with empty username returned %d results, want 0", len(results.Created))
	}
	if len(results.Participated) != 0 {
		t.Errorf("SearchPRs with empty username returned %d results, want 0", len(results.Participated))
	}
	if len(results.ReviewRequested) != 0 {
		t.Errorf("SearchPRs with empty username returned %d results, want 0", len(results.ReviewRequested))
	}
}

func TestPRSearchQuery_ContainsAliases(t *testing.T) {
	query := prSearchQuery

	if query == "" {
		t.Fatal("prSearchQuery is empty")
	}

	requiredParts := []string{
		"created: search",
		"assigned: search",
		"participated: search",
		"reviewRequested: search",
		"... on PullRequest",
		"statusCheckRollup",
	}

	for _, part := range requiredParts {
		if !strings.Contains(query, part) {
			t.Errorf("prSearchQuery should contain %q", part)
		}
	}
}

func TestBuildPRSearchVariables(t *testing.T) {
	vars := buildPRSearchVariables("testuser", nil)

	expected := map[string]string{
		"created":         "is:pr is:open author:testuser",
		"assigned":        "is:pr is:open assignee:testuser",
		"participated":    "is:pr is:open (mentions:testuser OR commenter:testuser)",
		"reviewRequested": "is:pr is:open review-requested:testuser",
	}

	for key, want := range expected {
		got, ok := vars[key]
		if !ok {
			t.Errorf("missing variable %q", key)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}
func TestBuildPRSearchVariables_WithTeams_ReviewRequested(t *testing.T) {
	vars := buildPRSearchVariables("testuser", []string{"my-org/team-a"})

	got, ok := vars["reviewRequested"]
	if !ok {
		t.Fatal("missing reviewRequested variable")
	}
	want := "is:pr is:open (review-requested:testuser OR team-review-requested:my-org/team-a)"
	if got != want {
		t.Errorf("reviewRequested = %q, want %q", got, want)
	}
}

func TestBuildPRSearchVariables_WithTeams_Participated(t *testing.T) {
	vars := buildPRSearchVariables("testuser", []string{"my-org/team-a"})

	got, ok := vars["participated"]
	if !ok {
		t.Fatal("missing participated variable")
	}
	want := "is:pr is:open (mentions:testuser OR commenter:testuser OR team:my-org/team-a)"
	if got != want {
		t.Errorf("participated = %q, want %q", got, want)
	}
}

func TestBuildPRSearchVariables_MultipleTeams(t *testing.T) {
	vars := buildPRSearchVariables("testuser", []string{"org-a/team-1", "org-b/team-2"})

	wantReview := "is:pr is:open (review-requested:testuser OR team-review-requested:org-a/team-1 OR team-review-requested:org-b/team-2)"
	if got := vars["reviewRequested"]; got != wantReview {
		t.Errorf("reviewRequested = %q, want %q", got, wantReview)
	}

	wantParticipated := "is:pr is:open (mentions:testuser OR commenter:testuser OR team:org-a/team-1 OR team:org-b/team-2)"
	if got := vars["participated"]; got != wantParticipated {
		t.Errorf("participated = %q, want %q", got, wantParticipated)
	}
}

func TestBuildPRSearchVariables_EmptyTeams(t *testing.T) {
	vars := buildPRSearchVariables("testuser", []string{})

	expected := map[string]string{
		"created":         "is:pr is:open author:testuser",
		"assigned":        "is:pr is:open assignee:testuser",
		"participated":    "is:pr is:open (mentions:testuser OR commenter:testuser)",
		"reviewRequested": "is:pr is:open review-requested:testuser",
	}

	for key, want := range expected {
		got, ok := vars[key]
		if !ok {
			t.Errorf("missing variable %q", key)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestParsePRSearchNodes(t *testing.T) {
	node1 := prSearchRawNode{
		Number:    10,
		Title:     "Add feature",
		URL:       "https://github.com/owner/repo/pull/10",
		IsDraft:   false,
		UpdatedAt: "2024-03-15T10:00:00Z",
		CreatedAt: "2024-03-10T08:00:00Z",
	}
	node1.Author.Login = "user1"
	node1.Repository.NameWithOwner = "owner/repo"

	node2 := prSearchRawNode{
		Number:    20,
		Title:     "Fix bug",
		URL:       "https://github.com/owner/repo/pull/20",
		IsDraft:   true,
		UpdatedAt: "2024-03-16T10:00:00Z",
		CreatedAt: "2024-03-11T08:00:00Z",
	}
	node2.Author.Login = "user2"
	node2.Repository.NameWithOwner = "owner/repo2"

	rawNodes := []prSearchRawNode{
		node1,
		{Number: 0}, // Should be skipped
		node2,
	}

	nodes := parsePRSearchNodes(rawNodes)

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	if nodes[0].Number != 10 {
		t.Errorf("nodes[0].Number = %d, want 10", nodes[0].Number)
	}
	if nodes[0].Title != "Add feature" {
		t.Errorf("nodes[0].Title = %q, want %q", nodes[0].Title, "Add feature")
	}
	if nodes[0].IsDraft != false {
		t.Errorf("nodes[0].IsDraft = %v, want false", nodes[0].IsDraft)
	}
	if nodes[0].Author.Login != "user1" {
		t.Errorf("nodes[0].Author.Login = %q, want %q", nodes[0].Author.Login, "user1")
	}
	if nodes[0].Repository.NameWithOwner != "owner/repo" {
		t.Errorf("nodes[0].Repository.NameWithOwner = %q, want %q", nodes[0].Repository.NameWithOwner, "owner/repo")
	}

	if nodes[1].Number != 20 {
		t.Errorf("nodes[1].Number = %d, want 20", nodes[1].Number)
	}
	if nodes[1].IsDraft != true {
		t.Errorf("nodes[1].IsDraft = %v, want true", nodes[1].IsDraft)
	}
}

func TestParsePRSearchNodes_CIStatus(t *testing.T) {
	successState := "SUCCESS"
	nodeWithCI := prSearchRawNode{
		Number: 1,
		Title:  "With CI",
	}
	nodeWithCI.Commits.Nodes = []struct {
		Commit struct {
			StatusCheckRollup *struct {
				State string `json:"state"`
			} `json:"statusCheckRollup"`
		} `json:"commit"`
	}{
		{Commit: struct {
			StatusCheckRollup *struct {
				State string `json:"state"`
			} `json:"statusCheckRollup"`
		}{StatusCheckRollup: &struct {
			State string `json:"state"`
		}{State: successState}}},
	}

	nodeWithoutCI := prSearchRawNode{
		Number: 2,
		Title:  "Without CI",
	}

	nodes := parsePRSearchNodes([]prSearchRawNode{nodeWithCI, nodeWithoutCI})

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	if nodes[0].StatusState != "SUCCESS" {
		t.Errorf("nodes[0].StatusState = %q, want %q", nodes[0].StatusState, "SUCCESS")
	}
	if nodes[0].CIStatus() != cistatus.CIStatusSuccess {
		t.Errorf("nodes[0].CIStatus() = %v, want %v", nodes[0].CIStatus(), cistatus.CIStatusSuccess)
	}

	if nodes[1].StatusState != "" {
		t.Errorf("nodes[1].StatusState = %q, want empty", nodes[1].StatusState)
	}
	if nodes[1].CIStatus() != cistatus.CIStatusNone {
		t.Errorf("nodes[1].CIStatus() = %v, want %v", nodes[1].CIStatus(), cistatus.CIStatusNone)
	}
}
