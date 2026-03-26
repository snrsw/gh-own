package pr

import (
	"strings"
	"testing"

	"github.com/snrsw/gh-own/internal/cistatus"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/reviewstatus"
)

func TestNewGroupedPullRequests_PropagatesCustom(t *testing.T) {
	ghResult := &gh.PRSearchResult{
		Created: []gh.PRSearchNode{{Number: 1}},
		Custom: map[string][]gh.PRSearchNode{
			"myTab": {{Number: 10, Title: "Custom PR"}},
		},
	}

	grouped := NewGroupedPullRequests(ghResult, "")

	if len(grouped.Custom) != 1 {
		t.Fatalf("Custom has %d keys, want 1", len(grouped.Custom))
	}
	sr, ok := grouped.Custom["myTab"]
	if !ok {
		t.Fatal("Custom[\"myTab\"] not found")
	}
	if sr.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", sr.TotalCount)
	}
}

func TestBuildTabs_DefaultTabsOnly(t *testing.T) {
	grouped := &GroupedPullRequests{
		Created:         gh.SearchResult[pullRequest]{TotalCount: 1, Items: []pullRequest{{Number: 1}}},
		Participated:    gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		Assigned:        gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		ReviewRequested: gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
	}

	tabs := grouped.BuildTabs()

	if len(tabs) != 4 {
		t.Fatalf("BuildTabs() returned %d tabs, want 4", len(tabs))
	}
}

func TestBuildTabs_WithCustomTabs(t *testing.T) {
	grouped := &GroupedPullRequests{
		Created:         gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		Participated:    gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		Assigned:        gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		ReviewRequested: gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		Custom: map[string]gh.SearchResult[pullRequest]{
			"zeta":  {TotalCount: 1, Items: []pullRequest{{Number: 1}}},
			"alpha": {TotalCount: 2, Items: []pullRequest{{Number: 2}, {Number: 3}}},
		},
	}

	tabs := grouped.BuildTabs()

	if len(tabs) != 6 {
		t.Fatalf("BuildTabs() returned %d tabs, want 6", len(tabs))
	}
	// Custom tabs should be sorted alphabetically at indices 4-5
	if tabs[4].Name() != "Alpha (2)" {
		t.Errorf("tabs[4].Name() = %q, want %q", tabs[4].Name(), "Alpha (2)")
	}
	if tabs[5].Name() != "Zeta (1)" {
		t.Errorf("tabs[5].Name() = %q, want %q", tabs[5].Name(), "Zeta (1)")
	}
}

func TestBuildTabs_CustomTabNameIncludesCount(t *testing.T) {
	grouped := &GroupedPullRequests{
		Created:         gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		Participated:    gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		Assigned:        gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		ReviewRequested: gh.SearchResult[pullRequest]{TotalCount: 0, Items: []pullRequest{}},
		Custom: map[string]gh.SearchResult[pullRequest]{
			"myTab": {TotalCount: 3, Items: []pullRequest{{Number: 1}, {Number: 2}, {Number: 3}}},
		},
	}

	tabs := grouped.BuildTabs()

	if len(tabs) != 5 {
		t.Fatalf("BuildTabs() returned %d tabs, want 5", len(tabs))
	}
	if tabs[4].Name() != "MyTab (3)" {
		t.Errorf("tabs[4].Name() = %q, want %q", tabs[4].Name(), "MyTab (3)")
	}
}

func TestPullRequest_RepositoryFullName(t *testing.T) {
	tests := []struct {
		name          string
		repositoryURL string
		expected      string
	}{
		{
			name:          "valid GitHub API URL",
			repositoryURL: "https://api.github.com/repos/owner/repo",
			expected:      "owner/repo",
		},
		{
			name:          "valid URL with longer path",
			repositoryURL: "https://api.github.com/repos/my-org/my-repo",
			expected:      "my-org/my-repo",
		},
		{
			name:          "URL with extra path segments",
			repositoryURL: "https://api.github.com/repos/owner/repo/pulls/123",
			expected:      "pulls/123",
		},
		{
			name:          "empty URL",
			repositoryURL: "",
			expected:      "",
		},
		{
			name:          "short URL with fewer than 5 parts",
			repositoryURL: "https://api.github.com/repos",
			expected:      "",
		},
		{
			name:          "exactly 5 parts",
			repositoryURL: "https://api.github.com/repos/owner/repo",
			expected:      "owner/repo",
		},
		{
			name:          "GitHub Enterprise URL",
			repositoryURL: "https://github.example.com/api/v3/repos/org/project",
			expected:      "org/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &pullRequest{
				RepositoryURL: tt.repositoryURL,
			}
			result := pr.repositoryFullName()
			if result != tt.expected {
				t.Errorf("PullRequest.repositoryFullName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPullRequest_ToItem_NoActivity(t *testing.T) {
	pr := pullRequest{
		Number:        42,
		User:          gh.User{Login: "alice"},
		RepositoryURL: "https://api.github.com/repos/owner/repo",
		Title:         "Fix bug",
		CreatedAt:     "2024-03-10T08:00:00Z",
		UpdatedAt:     "2024-03-10T12:00:00Z",
	}

	desc := pr.toItem("").Description()

	if !strings.Contains(desc, "updated") {
		t.Errorf("Description() = %q, should contain %q", desc, "updated")
	}
}

func TestPullRequest_ToItem_WithActivity(t *testing.T) {
	pr := pullRequest{
		Number:        42,
		User:          gh.User{Login: "alice"},
		RepositoryURL: "https://api.github.com/repos/owner/repo",
		Title:         "Fix bug",
		CreatedAt:     "2024-03-10T08:00:00Z",
		LatestActivity: gh.LatestActivity{
			Kind:  "approved",
			Login: "bob",
			At:    "2024-03-10T12:00:00Z",
		},
	}

	desc := pr.toItem("").Description()

	if !strings.Contains(desc, ", approved by @bob") {
		t.Errorf("Description() = %q, should contain %q", desc, ", approved by @bob")
	}
}

func TestPullRequest_ToItem(t *testing.T) {
	tests := []struct {
		name           string
		pr             pullRequest
		expectedTitle  string
		filterContains []string
		descContains   []string
	}{
		{
			name: "regular PR",
			pr: pullRequest{
				Number:        123,
				User:          gh.User{Login: "contributor"},
				RepositoryURL: "https://api.github.com/repos/owner/repo",
				Title:         "Add new feature",
				Draft:         false,
				CreatedAt:     "2024-03-10T08:00:00Z",
			},
			expectedTitle:  "owner/repo",
			filterContains: []string{"#123", "Add new feature"},
			descContains:   []string{"2024-03-10", "@contributor"},
		},
		{
			name: "draft PR",
			pr: pullRequest{
				Number:        456,
				User:          gh.User{Login: "author"},
				RepositoryURL: "https://api.github.com/repos/org/project",
				Title:         "Work in progress",
				Draft:         true,
				CreatedAt:     "2024-01-15T12:00:00Z",
			},
			expectedTitle:  "org/project",
			filterContains: []string{"#456", "Work in progress"},
			descContains:   []string{"2024-01-15", "@author"},
		},
		{
			name: "PR with CI status success",
			pr: pullRequest{
				Number:        789,
				User:          gh.User{Login: "dev"},
				RepositoryURL: "https://api.github.com/repos/owner/repo",
				Title:         "Feature with CI",
				Draft:         false,
				CreatedAt:     "2024-03-15T10:00:00Z",
				CIStatus:      cistatus.CIStatusSuccess,
			},
			expectedTitle:  "owner/repo",
			filterContains: []string{"#789", "Feature with CI"},
		},
		{
			name: "PR with CI status failure",
			pr: pullRequest{
				Number:        101,
				User:          gh.User{Login: "dev"},
				RepositoryURL: "https://api.github.com/repos/owner/repo",
				Title:         "Failing CI",
				Draft:         false,
				CreatedAt:     "2024-03-15T10:00:00Z",
				CIStatus:      cistatus.CIStatusFailure,
			},
			expectedTitle:  "owner/repo",
			filterContains: []string{"#101", "Failing CI"},
		},
		{
			name: "PR with approved review and CI success",
			pr: pullRequest{
				Number:        200,
				User:          gh.User{Login: "dev"},
				RepositoryURL: "https://api.github.com/repos/owner/repo",
				Title:         "Ready to merge",
				Draft:         false,
				CreatedAt:     "2024-03-15T10:00:00Z",
				CIStatus:      cistatus.CIStatusSuccess,
				ReviewStatus:  reviewstatus.ReviewStatusApproved,
			},
			expectedTitle:  "owner/repo",
			filterContains: []string{"#200", "Ready to merge"},
		},
		{
			name: "PR with changes requested",
			pr: pullRequest{
				Number:        201,
				User:          gh.User{Login: "dev"},
				RepositoryURL: "https://api.github.com/repos/owner/repo",
				Title:         "Needs work",
				Draft:         false,
				CreatedAt:     "2024-03-15T10:00:00Z",
				ReviewStatus:  reviewstatus.ReviewStatusChangesRequested,
			},
			expectedTitle:  "owner/repo",
			filterContains: []string{"#201", "Needs work"},
		},
		{
			name: "PR with review required",
			pr: pullRequest{
				Number:        202,
				User:          gh.User{Login: "dev"},
				RepositoryURL: "https://api.github.com/repos/owner/repo",
				Title:         "Awaiting review",
				Draft:         false,
				CreatedAt:     "2024-03-15T10:00:00Z",
				ReviewStatus:  reviewstatus.ReviewStatusReviewRequired,
			},
			expectedTitle:  "owner/repo",
			filterContains: []string{"#202", "Awaiting review"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.pr.toItem("")

			if got := item.Title(); got != tt.expectedTitle {
				t.Errorf("Title() = %q, want %q", got, tt.expectedTitle)
			}

			filterVal := item.FilterValue()
			for _, part := range tt.filterContains {
				if !strings.Contains(filterVal, part) {
					t.Errorf("FilterValue() = %q, should contain %q", filterVal, part)
				}
			}

			desc := item.Description()
			for _, part := range tt.descContains {
				if !strings.Contains(desc, part) {
					t.Errorf("Description() = %q, should contain %q", desc, part)
				}
			}
		})
	}
}

func TestRenderPRNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		draft    bool
		contains string
	}{
		{
			name:     "regular PR",
			number:   42,
			draft:    false,
			contains: "#42",
		},
		{
			name:     "draft PR",
			number:   99,
			draft:    true,
			contains: "#99",
		},
		{
			name:     "large number",
			number:   12345,
			draft:    false,
			contains: "#12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderPRNumber(tt.number, tt.draft)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("RenderPRNumber(%d, %v) = %q, should contain %q", tt.number, tt.draft, result, tt.contains)
			}
		})
	}
}

func TestGroupedPullRequests_PRItems(t *testing.T) {
	tests := []struct {
		name     string
		input    gh.SearchResult[pullRequest]
		expected int
	}{
		{
			name: "multiple PRs",
			input: gh.SearchResult[pullRequest]{
				TotalCount: 2,
				Items: []pullRequest{
					{Number: 1, RepositoryURL: "https://api.github.com/repos/owner/repo1", Title: "PR 1"},
					{Number: 2, RepositoryURL: "https://api.github.com/repos/owner/repo2", Title: "PR 2"},
				},
			},
			expected: 2,
		},
		{
			name: "empty list",
			input: gh.SearchResult[pullRequest]{
				TotalCount: 0,
				Items:      []pullRequest{},
			},
			expected: 0,
		},
		{
			name: "single PR",
			input: gh.SearchResult[pullRequest]{
				TotalCount: 1,
				Items: []pullRequest{
					{Number: 42, RepositoryURL: "https://api.github.com/repos/owner/repo", Title: "Solo PR"},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouped := &GroupedPullRequests{Created: tt.input}
			items := grouped.prItems(grouped.Created)

			if len(items) != tt.expected {
				t.Errorf("prItems() returned %d items, want %d", len(items), tt.expected)
			}
		})
	}
}

func TestPullRequest_HasCIStatusField(t *testing.T) {
	pr := pullRequest{
		Number:   123,
		CIStatus: cistatus.CIStatusSuccess,
	}

	if pr.CIStatus != cistatus.CIStatusSuccess {
		t.Errorf("CIStatus = %v, want %v", pr.CIStatus, cistatus.CIStatusSuccess)
	}
}

func TestFromGraphQL_PropagatesLatestActivity(t *testing.T) {
	node := gh.PRSearchNode{
		Number: 1,
		Title:  "Test",
	}
	node.LatestActivity = gh.LatestActivity{Kind: "commented", Login: "alice", At: "2024-03-10T10:00:00Z"}
	node.Repository.NameWithOwner = "owner/repo"

	pr := fromGraphQL(node)

	if pr.LatestActivity.Kind != "commented" {
		t.Errorf("Kind = %q, want %q", pr.LatestActivity.Kind, "commented")
	}
	if pr.LatestActivity.Login != "alice" {
		t.Errorf("Login = %q, want %q", pr.LatestActivity.Login, "alice")
	}
}

func TestFromGraphQL(t *testing.T) {
	node := gh.PRSearchNode{
		Number:      123,
		Title:       "Test PR",
		URL:         "https://github.com/owner/repo/pull/123",
		IsDraft:     false,
		UpdatedAt:   "2024-03-10T10:00:00Z",
		CreatedAt:   "2024-03-10T08:00:00Z",
		StatusState: "SUCCESS",
	}
	node.Author.Login = "testuser"
	node.Repository.NameWithOwner = "owner/repo"

	pr := fromGraphQL(node)

	if pr.Number != 123 {
		t.Errorf("Number = %d, want 123", pr.Number)
	}
	if pr.Title != "Test PR" {
		t.Errorf("Title = %q, want %q", pr.Title, "Test PR")
	}
	if pr.User.Login != "testuser" {
		t.Errorf("User.Login = %q, want %q", pr.User.Login, "testuser")
	}
	if pr.CIStatus != cistatus.CIStatusSuccess {
		t.Errorf("CIStatus = %v, want %v", pr.CIStatus, cistatus.CIStatusSuccess)
	}
	if pr.HTMLURL != "https://github.com/owner/repo/pull/123" {
		t.Errorf("HTMLURL = %q, want %q", pr.HTMLURL, "https://github.com/owner/repo/pull/123")
	}
}

func TestFromGraphQL_PropagatesReviewStatus(t *testing.T) {
	tests := []struct {
		name     string
		decision string
		want     reviewstatus.ReviewStatus
	}{
		{"approved", "APPROVED", reviewstatus.ReviewStatusApproved},
		{"changes requested", "CHANGES_REQUESTED", reviewstatus.ReviewStatusChangesRequested},
		{"review required", "REVIEW_REQUIRED", reviewstatus.ReviewStatusReviewRequired},
		{"none", "", reviewstatus.ReviewStatusNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := gh.PRSearchNode{
				Number:         1,
				Title:          "Test",
				ReviewDecision: tt.decision,
			}
			node.Repository.NameWithOwner = "owner/repo"

			pr := fromGraphQL(node)

			if pr.ReviewStatus != tt.want {
				t.Errorf("ReviewStatus = %v, want %v", pr.ReviewStatus, tt.want)
			}
		})
	}
}

func TestFromGraphQLNodes(t *testing.T) {
	nodes := []gh.PRSearchNode{
		{Number: 1, Title: "PR 1", StatusState: "SUCCESS"},
		{Number: 2, Title: "PR 2", StatusState: "FAILURE"},
	}

	prs := fromGraphQLNodes(nodes)

	if len(prs) != 2 {
		t.Fatalf("FromGraphQLNodes returned %d items, want 2", len(prs))
	}
	if prs[0].Number != 1 {
		t.Errorf("prs[0].Number = %d, want 1", prs[0].Number)
	}
	if prs[1].CIStatus != cistatus.CIStatusFailure {
		t.Errorf("prs[1].CIStatus = %v, want %v", prs[1].CIStatus, cistatus.CIStatusFailure)
	}
}
