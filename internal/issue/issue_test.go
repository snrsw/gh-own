package issue

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/ui"
)

func TestNewGroupedIssues_PropagatesCustom(t *testing.T) {
	ghResult := &gh.IssueSearchResult{
		Created: []gh.IssueSearchNode{{Number: 1}},
		Custom: map[string][]gh.IssueSearchNode{
			"myTab": {{Number: 10, Title: "Custom Issue"}},
		},
	}

	grouped := NewGroupedIssues(ghResult, "")

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

func TestBuildTabs_Issue_DefaultTabsOnly(t *testing.T) {
	grouped := &GroupedIssues{
		Created:      gh.SearchResult[issue]{TotalCount: 1, Items: []issue{{Number: 1}}},
		Participated: gh.SearchResult[issue]{TotalCount: 0, Items: []issue{}},
		Assigned:     gh.SearchResult[issue]{TotalCount: 0, Items: []issue{}},
	}

	tabs := grouped.BuildTabs()

	if len(tabs) != 3 {
		t.Fatalf("BuildTabs() returned %d tabs, want 3", len(tabs))
	}
}

func TestBuildTabs_Issue_WithCustomTabs(t *testing.T) {
	grouped := &GroupedIssues{
		Created:      gh.SearchResult[issue]{TotalCount: 0, Items: []issue{}},
		Participated: gh.SearchResult[issue]{TotalCount: 0, Items: []issue{}},
		Assigned:     gh.SearchResult[issue]{TotalCount: 0, Items: []issue{}},
		Custom: map[string]gh.SearchResult[issue]{
			"zeta":  {TotalCount: 1, Items: []issue{{Number: 1}}},
			"alpha": {TotalCount: 2, Items: []issue{{Number: 2}, {Number: 3}}},
		},
	}

	tabs := grouped.BuildTabs()

	if len(tabs) != 5 {
		t.Fatalf("BuildTabs() returned %d tabs, want 5", len(tabs))
	}
	if tabs[3].Name() != "Alpha (2)" {
		t.Errorf("tabs[3].Name() = %q, want %q", tabs[3].Name(), "Alpha (2)")
	}
	if tabs[4].Name() != "Zeta (1)" {
		t.Errorf("tabs[4].Name() = %q, want %q", tabs[4].Name(), "Zeta (1)")
	}
}

func TestIssue_RepositoryFullName(t *testing.T) {
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
			repositoryURL: "https://api.github.com/repos/owner/repo/issues/123",
			expected:      "issues/123",
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
			issue := &issue{
				RepositoryURL: tt.repositoryURL,
			}
			result := issue.repositoryFullName()
			if result != tt.expected {
				t.Errorf("Issue.repositoryFullName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIssue_ToItem_NoActivity(t *testing.T) {
	i := issue{
		Number:        7,
		User:          gh.User{Login: "carol"},
		RepositoryURL: "https://api.github.com/repos/owner/repo",
		Title:         "Fix login",
		CreatedAt:     "2024-03-10T08:00:00Z",
		UpdatedAt:     "2024-03-10T12:00:00Z",
	}

	desc := i.toItem("").Description()

	if !strings.Contains(desc, "updated") {
		t.Errorf("Description() = %q, should contain %q", desc, "updated")
	}
}

func TestIssue_ToItem_WithActivity(t *testing.T) {
	i := issue{
		Number:        7,
		User:          gh.User{Login: "carol"},
		RepositoryURL: "https://api.github.com/repos/owner/repo",
		Title:         "Fix login",
		CreatedAt:     "2024-03-10T08:00:00Z",
		LatestActivity: gh.LatestActivity{
			Kind:  "commented",
			Login: "alice",
			At:    "2024-03-10T12:00:00Z",
		},
	}

	desc := i.toItem("").Description()

	// Activity now comes first in description, styled via RenderActivityKind
	if !strings.Contains(desc, "commented") {
		t.Errorf("Description() = %q, should contain %q", desc, "commented")
	}
	if !strings.Contains(desc, "by @alice") {
		t.Errorf("Description() = %q, should contain %q", desc, "by @alice")
	}
	if !strings.Contains(desc, "opened on") {
		t.Errorf("Description() = %q, should contain %q", desc, "opened on")
	}
}

func TestIssue_ToItem(t *testing.T) {
	_issue := issue{
		Number:        42,
		User:          gh.User{Login: "testuser"},
		RepositoryURL: "https://api.github.com/repos/owner/repo",
		Title:         "Fix the bug",
		State:         "open",
		HTMLURL:       "https://github.com/owner/repo/issues/42",
		UpdatedAt:     "2024-03-15T10:30:00Z",
		CreatedAt:     "2024-03-10T08:00:00Z",
	}

	item := _issue.toItem("")

	// Title is just the repo name; number and title text are on the title line
	if got := item.Title(); got != "owner/repo" {
		t.Errorf("Title() = %q, want %q", got, "owner/repo")
	}

	// FilterValue includes number and title for search
	filterVal := item.FilterValue()
	for _, part := range []string{"#42", "Fix the bug"} {
		if !strings.Contains(filterVal, part) {
			t.Errorf("FilterValue() = %q, should contain %q", filterVal, part)
		}
	}

	// Description contains date and author; number has moved to the title line
	desc := item.Description()
	for _, part := range []string{"2024-03-10", "@testuser"} {
		if !strings.Contains(desc, part) {
			t.Errorf("Description() = %q, should contain %q", desc, part)
		}
	}
}

func TestGroupedIssues_IssueItems(t *testing.T) {
	tests := []struct {
		name     string
		input    gh.SearchResult[issue]
		expected int
	}{
		{
			name: "multiple issues",
			input: gh.SearchResult[issue]{
				TotalCount: 2,
				Items: []issue{
					{Number: 1, RepositoryURL: "https://api.github.com/repos/owner/repo1", Title: "Issue 1"},
					{Number: 2, RepositoryURL: "https://api.github.com/repos/owner/repo2", Title: "Issue 2"},
				},
			},

			expected: 2,
		},
		{
			name: "empty list",
			input: gh.SearchResult[issue]{
				TotalCount: 0,
				Items:      []issue{},
			},
			expected: 0,
		},
		{
			name: "single issue",
			input: gh.SearchResult[issue]{
				TotalCount: 1,
				Items: []issue{
					{Number: 42, RepositoryURL: "https://api.github.com/repos/owner/repo", Title: "Solo Issue"},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouped := &GroupedIssues{Created: tt.input}
			items := grouped.issueItems(grouped.Created)

			if len(items) != tt.expected {
				t.Errorf("issueItems() returned %d items, want %d", len(items), tt.expected)
			}
		})
	}
}

func TestFromGraphQL_PropagatesLatestActivity(t *testing.T) {
	node := gh.IssueSearchNode{
		Number: 1,
		Title:  "Test",
	}
	node.LatestActivity = gh.LatestActivity{Kind: "commented", Login: "alice", At: "2024-03-10T10:00:00Z"}
	node.Repository.NameWithOwner = "owner/repo"

	i := fromGraphQL(node)

	if i.LatestActivity.Kind != "commented" {
		t.Errorf("Kind = %q, want %q", i.LatestActivity.Kind, "commented")
	}
	if i.LatestActivity.Login != "alice" {
		t.Errorf("Login = %q, want %q", i.LatestActivity.Login, "alice")
	}
}

func TestIssueFromNode(t *testing.T) {
	node := gh.IssueSearchNode{
		Number:    42,
		Title:     "Test Issue",
		URL:       "https://github.com/owner/repo/issues/42",
		State:     "OPEN",
		UpdatedAt: "2024-03-15T10:00:00Z",
		CreatedAt: "2024-03-10T08:00:00Z",
	}
	node.Author.Login = "testuser"
	node.Repository.NameWithOwner = "owner/repo"

	got := fromGraphQL(node)

	if got.Number != 42 {
		t.Errorf("Number = %d, want 42", got.Number)
	}
	if got.Title != "Test Issue" {
		t.Errorf("Title = %q, want %q", got.Title, "Test Issue")
	}
	if got.HTMLURL != "https://github.com/owner/repo/issues/42" {
		t.Errorf("HTMLURL = %q, want %q", got.HTMLURL, "https://github.com/owner/repo/issues/42")
	}
	if got.State != "OPEN" {
		t.Errorf("State = %q, want %q", got.State, "OPEN")
	}
	if got.User.Login != "testuser" {
		t.Errorf("User.Login = %q, want %q", got.User.Login, "testuser")
	}
	if got.RepositoryURL != "https://api.github.com/repos/owner/repo" {
		t.Errorf("RepositoryURL = %q, want %q", got.RepositoryURL, "https://api.github.com/repos/owner/repo")
	}
}

func issueNumbers(items []issue) []int {
	numbers := make([]int, len(items))
	for i, is := range items {
		numbers[i] = is.Number
	}
	return numbers
}

func TestNewGroupedIssues_SortsByUpdatedDesc(t *testing.T) {
	ghResult := &gh.IssueSearchResult{
		Created: []gh.IssueSearchNode{
			{Number: 1, UpdatedAt: "2024-03-01T00:00:00Z"},
			{Number: 2, UpdatedAt: "2024-03-20T00:00:00Z"},
			{Number: 3, UpdatedAt: "2024-03-10T00:00:00Z"},
		},
	}

	grouped := NewGroupedIssues(ghResult, "")

	want := []int{2, 3, 1}
	if got := issueNumbers(grouped.Created.Items); !slices.Equal(got, want) {
		t.Errorf("Created.Items = %v, want %v", got, want)
	}
	if grouped.Created.TotalCount != 3 {
		t.Errorf("Created.TotalCount = %d, want 3", grouped.Created.TotalCount)
	}
}

func TestNewGroupedIssues_SortsCustomTabs(t *testing.T) {
	ghResult := &gh.IssueSearchResult{
		Custom: map[string][]gh.IssueSearchNode{
			"myTab": {
				{Number: 1, UpdatedAt: "2024-03-01T00:00:00Z"},
				{Number: 2, UpdatedAt: "2024-03-20T00:00:00Z"},
			},
		},
	}

	grouped := NewGroupedIssues(ghResult, "")

	want := []int{2, 1}
	if got := issueNumbers(grouped.Custom["myTab"].Items); !slices.Equal(got, want) {
		t.Errorf("Custom[\"myTab\"].Items = %v, want %v", got, want)
	}
}

func TestNewGroupedIssues_SortsByDisplayedTime(t *testing.T) {
	// Issue 1 has the newest updatedAt but the oldest activity time, and the
	// activity time is what its description line shows.
	ghResult := &gh.IssueSearchResult{
		Created: []gh.IssueSearchNode{
			{
				Number:         1,
				UpdatedAt:      "2024-03-20T00:00:00Z",
				LatestActivity: gh.LatestActivity{Kind: "commented", Login: "alice", At: "2024-03-01T00:00:00Z"},
			},
			{Number: 2, UpdatedAt: "2024-03-10T00:00:00Z"},
		},
	}

	grouped := NewGroupedIssues(ghResult, "")

	want := []int{2, 1}
	if got := issueNumbers(grouped.Created.Items); !slices.Equal(got, want) {
		t.Errorf("Created.Items = %v, want %v", got, want)
	}
}

func TestNewGroupedIssues_MissingTimestampSortsLast(t *testing.T) {
	ghResult := &gh.IssueSearchResult{
		Created: []gh.IssueSearchNode{
			{Number: 1},
			{Number: 2, UpdatedAt: "2024-03-01T00:00:00Z"},
		},
	}

	grouped := NewGroupedIssues(ghResult, "")

	want := []int{2, 1}
	if got := issueNumbers(grouped.Created.Items); !slices.Equal(got, want) {
		t.Errorf("Created.Items = %v, want %v", got, want)
	}
}

func TestGroupedIssues_IssueItems_CarrySortAt(t *testing.T) {
	// The sort timestamp has to survive the hop from the domain struct into the
	// ui.Item; without it the list has nothing to reorder.
	ghResult := &gh.IssueSearchResult{
		Created: []gh.IssueSearchNode{
			{Number: 1, UpdatedAt: "2024-03-01T00:00:00Z"},
			{Number: 2, UpdatedAt: "2024-03-20T00:00:00Z"},
		},
	}
	grouped := NewGroupedIssues(ghResult, "")

	items := grouped.issueItems(grouped.Created)

	want := []time.Time{
		time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	if len(items) != len(want) {
		t.Fatalf("issueItems() returned %d items, want %d", len(items), len(want))
	}
	for i, item := range items {
		it, ok := item.(ui.Item)
		if !ok {
			t.Fatalf("item %d is %T, want ui.Item", i, item)
		}
		if !it.SortAt().Equal(want[i]) {
			t.Errorf("item %d SortAt() = %v, want %v", i, it.SortAt(), want[i])
		}
	}
}
