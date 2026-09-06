package gh

import (
	"reflect"
	"strings"
	"testing"

	"github.com/snrsw/gh-own/internal/cistatus"
	"github.com/snrsw/gh-own/internal/config"
)

func TestParsePRSearchResult_CustomKeyPreserved(t *testing.T) {
	parsed := map[string][]PRSearchNode{
		"drafts": {{Number: 1, Title: "PR1"}},
		"myTab":  {{Number: 2, Title: "PR2"}},
	}

	result, err := parsePRSearchResult(parsed)
	if err != nil {
		t.Fatalf("parsePRSearchResult returned error: %v", err)
	}

	if len(result.Custom) == 0 {
		t.Fatal("Custom map is empty, want key \"myTab\"")
	}
	nodes, ok := result.Custom["myTab"]
	if !ok {
		t.Fatal("Custom[\"myTab\"] not found")
	}
	if len(nodes) != 1 || nodes[0].Number != 2 {
		t.Errorf("Custom[\"myTab\"] = %v, want [{Number:2}]", nodes)
	}
}

func TestParsePRSearchResult_NoCustomKeys(t *testing.T) {
	parsed := map[string][]PRSearchNode{
		"drafts":           {{Number: 1}},
		"needsAction":      {{Number: 2}},
		"readyToMerge":     {{Number: 3}},
		"waiting":          {{Number: 4}},
		"participatedUser": {{Number: 5}},
		"reviewRequested":  {{Number: 6}},
	}

	result, err := parsePRSearchResult(parsed)
	if err != nil {
		t.Fatalf("parsePRSearchResult returned error: %v", err)
	}

	if result.Custom == nil {
		t.Fatal("Custom should not be nil")
	}
	if len(result.Custom) != 0 {
		t.Errorf("Custom has %d keys, want 0", len(result.Custom))
	}
}

func TestParsePRSearchResult_MergesAuthoredAndAssigned(t *testing.T) {
	// Each state bucket is fed by an author-variant and an assignee-variant
	// query (the {owner} expansion: "drafts" + "draftsAssigned"); both must
	// merge into one bucket and dedup by URL.
	parsed := map[string][]PRSearchNode{
		"drafts": {{Number: 1, URL: "https://github.com/org/repo/pull/1"}},
		"draftsAssigned": {
			{Number: 1, URL: "https://github.com/org/repo/pull/1"}, // duplicate (authored & assigned)
			{Number: 2, URL: "https://github.com/org/repo/pull/2"},
		},
	}

	result, err := parsePRSearchResult(parsed)
	if err != nil {
		t.Fatalf("parsePRSearchResult returned error: %v", err)
	}

	if len(result.Drafts) != 2 {
		t.Errorf("Drafts has %d nodes, want 2 (merged and deduplicated)", len(result.Drafts))
	}
}

func TestMergeSearchPRsResults_MergesCustom(t *testing.T) {
	a := &PRSearchResult{
		Custom: map[string][]PRSearchNode{
			"alpha": {{Number: 1, URL: "https://github.com/org/repo/pull/1"}},
			"beta":  {{Number: 2, URL: "https://github.com/org/repo/pull/2"}},
		},
	}
	b := &PRSearchResult{
		Custom: map[string][]PRSearchNode{
			"beta":  {{Number: 3, URL: "https://github.com/org/repo/pull/3"}},
			"gamma": {{Number: 4, URL: "https://github.com/org/repo/pull/4"}},
		},
	}

	merged := MergeSearchPRsResults(a, b)

	if len(merged.Custom) != 3 {
		t.Fatalf("Custom has %d keys, want 3", len(merged.Custom))
	}
	if len(merged.Custom["alpha"]) != 1 {
		t.Errorf("Custom[alpha] has %d nodes, want 1", len(merged.Custom["alpha"]))
	}
	if len(merged.Custom["beta"]) != 2 {
		t.Errorf("Custom[beta] has %d nodes, want 2", len(merged.Custom["beta"]))
	}
	if len(merged.Custom["gamma"]) != 1 {
		t.Errorf("Custom[gamma] has %d nodes, want 1", len(merged.Custom["gamma"]))
	}
}

func TestMergeSearchPRsResults_DeduplicatesCustomByURL(t *testing.T) {
	a := &PRSearchResult{
		Custom: map[string][]PRSearchNode{
			"myTab": {{Number: 1, URL: "https://github.com/org/repo/pull/1"}},
		},
	}
	b := &PRSearchResult{
		Custom: map[string][]PRSearchNode{
			"myTab": {
				{Number: 1, URL: "https://github.com/org/repo/pull/1"},
				{Number: 2, URL: "https://github.com/org/repo/pull/2"},
			},
		},
	}

	merged := MergeSearchPRsResults(a, b)

	if len(merged.Custom["myTab"]) != 2 {
		t.Errorf("Custom[myTab] has %d nodes, want 2 (deduplicated)", len(merged.Custom["myTab"]))
	}
}

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

func TestSearchPRs_EmptyEntries(t *testing.T) {
	results, err := SearchPRs(nil, nil, true)

	if err != nil {
		t.Errorf("SearchPRs with empty username returned error: %v", err)
	}

	if len(results.Drafts) != 0 {
		t.Errorf("SearchPRs with empty username returned %d results, want 0", len(results.Drafts))
	}
	if len(results.NeedsAction) != 0 {
		t.Errorf("SearchPRs with empty username returned %d results, want 0", len(results.NeedsAction))
	}
	if len(results.Participated) != 0 {
		t.Errorf("SearchPRs with empty username returned %d results, want 0", len(results.Participated))
	}
	if len(results.ReviewRequested) != 0 {
		t.Errorf("SearchPRs with empty username returned %d results, want 0", len(results.ReviewRequested))
	}
}

func TestSearchPRs_EmptyEntries_HasEmptyCustom(t *testing.T) {
	results, err := SearchPRs(nil, nil, true)

	if err != nil {
		t.Fatalf("SearchPRs returned error: %v", err)
	}

	if results.Custom == nil {
		t.Fatal("Custom should not be nil")
	}
	if len(results.Custom) != 0 {
		t.Errorf("Custom has %d keys, want 0", len(results.Custom))
	}
}

func TestSearchPRs_EmptyUsernameWithTeams(t *testing.T) {
	results, err := SearchPRsTeams(nil, "", []string{"my-org/team-a"}, config.Filters{}, true)

	if err != nil {
		t.Errorf("SearchPRs with empty username returned error: %v", err)
	}

	if len(results.Drafts) != 0 {
		t.Errorf("SearchPRs with empty username returned %d drafts, want 0", len(results.Drafts))
	}
	if len(results.ReviewRequested) != 0 {
		t.Errorf("SearchPRs with empty username returned %d reviewRequested, want 0", len(results.ReviewRequested))
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

func TestParsePRSearchNodes_WithPush(t *testing.T) {
	node := prSearchRawNode{Number: 1, Title: "Test"}
	node.Commits.Nodes = []rawCommitNode{{}}
	node.Commits.Nodes[0].Commit.CommittedDate = "2024-03-10T12:00:00Z"
	commitUser := struct {
		Login string `json:"login"`
	}{Login: "charlie"}
	node.Commits.Nodes[0].Commit.Author.User = &commitUser

	nodes := parsePRSearchNodes([]prSearchRawNode{node})

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LatestActivity.Kind != "pushed" {
		t.Errorf("Kind = %q, want %q", nodes[0].LatestActivity.Kind, "pushed")
	}
	if nodes[0].LatestActivity.Login != "charlie" {
		t.Errorf("Login = %q, want %q", nodes[0].LatestActivity.Login, "charlie")
	}
}

func TestParsePRSearchNodes_NoActivity(t *testing.T) {
	node := prSearchRawNode{Number: 1, Title: "Test"}

	nodes := parsePRSearchNodes([]prSearchRawNode{node})

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

func TestParsePRSearchNodes_WithApprovedReview(t *testing.T) {
	node := prSearchRawNode{Number: 1, Title: "Test"}
	node.Reviews.Nodes = []rawReview{{SubmittedAt: "2024-03-10T12:00:00Z", State: "APPROVED"}}
	node.Reviews.Nodes[0].Author.Login = "bob"

	nodes := parsePRSearchNodes([]prSearchRawNode{node})

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].LatestActivity.Kind != "approved" {
		t.Errorf("Kind = %q, want %q", nodes[0].LatestActivity.Kind, "approved")
	}
	if nodes[0].LatestActivity.Login != "bob" {
		t.Errorf("Login = %q, want %q", nodes[0].LatestActivity.Login, "bob")
	}
}

func TestParsePRSearchNodes_WithComment(t *testing.T) {
	node := prSearchRawNode{Number: 1, Title: "Test"}
	node.Comments.Nodes = []rawComment{{CreatedAt: "2024-03-10T12:00:00Z"}}
	node.Comments.Nodes[0].Author.Login = "alice"

	nodes := parsePRSearchNodes([]prSearchRawNode{node})

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

func TestParsePRSearchNodes_CIStatus(t *testing.T) {
	successState := "SUCCESS"
	nodeWithCI := prSearchRawNode{
		Number: 1,
		Title:  "With CI",
	}
	nodeWithCI.Commits.Nodes = []rawCommitNode{{}}
	nodeWithCI.Commits.Nodes[0].Commit.StatusCheckRollup = &struct {
		State string `json:"state"`
	}{State: successState}

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

func TestParsePRSearchNodes_ReviewDecision(t *testing.T) {
	tests := []struct {
		name     string
		decision string
		want     string
	}{
		{"approved", "APPROVED", "APPROVED"},
		{"changes requested", "CHANGES_REQUESTED", "CHANGES_REQUESTED"},
		{"review required", "REVIEW_REQUIRED", "REVIEW_REQUIRED"},
		{"none", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := prSearchRawNode{Number: 1, Title: "Test"}
			node.ReviewDecision = tt.decision

			nodes := parsePRSearchNodes([]prSearchRawNode{node})

			if len(nodes) != 1 {
				t.Fatalf("expected 1 node, got %d", len(nodes))
			}
			if nodes[0].ReviewDecision != tt.want {
				t.Errorf("ReviewDecision = %q, want %q", nodes[0].ReviewDecision, tt.want)
			}
		})
	}
}

func TestParsePRSearchResult_DeterministicBucketOrder(t *testing.T) {
	// Buckets fed by several queries are assembled from a map filled by
	// parallel searches, so the concatenation order must not depend on map
	// iteration order. A single run would pass by chance; loop to be sure.
	parsed := map[string][]PRSearchNode{
		"drafts":         {{Number: 1, URL: "https://github.com/org/repo/pull/1"}},
		"draftsAssigned": {{Number: 2, URL: "https://github.com/org/repo/pull/2"}},
	}

	want := []int{1, 2}
	for i := 0; i < 20; i++ {
		result, err := parsePRSearchResult(parsed)
		if err != nil {
			t.Fatalf("parsePRSearchResult returned error: %v", err)
		}
		got := make([]int, 0, len(result.Drafts))
		for _, node := range result.Drafts {
			got = append(got, node.Number)
		}
		if len(got) != len(want) {
			t.Fatalf("Drafts = %v, want %v", got, want)
		}
		for j := range want {
			if got[j] != want[j] {
				t.Fatalf("Drafts = %v, want %v (iteration %d)", got, want, i)
			}
		}
	}
}

func TestPRTeamEntries(t *testing.T) {
	got := prTeamEntries([]string{"my-org/team-a", "my-org/team-b"}, config.Filters{Org: "my-org"})

	want := map[string]string{
		"participatedTeam0": "is:pr is:open team:my-org/team-a org:my-org sort:updated-desc",
		"participatedTeam1": "is:pr is:open team:my-org/team-b org:my-org sort:updated-desc",
	}
	if len(got) != len(want) {
		t.Fatalf("prTeamEntries() has %d entries, want %d", len(got), len(want))
	}
	for key, wantQuery := range want {
		if got[key] != wantQuery {
			t.Errorf("prTeamEntries()[%q] = %q, want %q", key, got[key], wantQuery)
		}
	}
}

func TestPRTeamEntries_WithoutOrg(t *testing.T) {
	got := prTeamEntries([]string{"my-org/team-a"}, config.Filters{})

	want := "is:pr is:open team:my-org/team-a sort:updated-desc"
	if got["participatedTeam0"] != want {
		t.Errorf("prTeamEntries()[%q] = %q, want %q", "participatedTeam0", got["participatedTeam0"], want)
	}
}

func TestPRTeamEntries_AppliesExcludeAuthors(t *testing.T) {
	got := prTeamEntries([]string{"my-org/team-a"}, config.Filters{
		ExcludeAuthors: []string{"renovate[bot]"},
	})

	want := "is:pr is:open team:my-org/team-a -author:renovate[bot] sort:updated-desc"
	if got["participatedTeam0"] != want {
		t.Errorf("prTeamEntries()[%q] = %q, want %q", "participatedTeam0", got["participatedTeam0"], want)
	}
}

const conversationJSON = `{"nodes": [{
	"number": 7,
	"title": "Add thing",
	"url": "https://github.com/owner/repo/pull/7",
	"body": "cc @me",
	"createdAt": "2024-03-10T08:00:00Z",
	"author": {"login": "alice"},
	"repository": {"nameWithOwner": "owner/repo"},
	"commits": {"nodes": [
		{"commit": {"statusCheckRollup": {"state": "FAILURE"}, "committedDate": "2024-03-10T09:00:00Z", "author": {"user": {"login": "alice"}}}},
		{"commit": {"statusCheckRollup": {"state": "SUCCESS"}, "committedDate": "2024-03-10T10:00:00Z", "author": {"user": null}}}
	]},
	"comments": {"nodes": [
		{"author": {"__typename": "Bot", "login": "codecov"}, "body": "90%", "createdAt": "2024-03-10T11:00:00Z"},
		{"author": {"__typename": "User", "login": "bob"}, "body": "nice", "createdAt": "2024-03-10T12:00:00Z"}
	]},
	"reviews": {"nodes": [
		{"author": {"__typename": "User", "login": "me"}, "body": "", "submittedAt": null, "state": "PENDING"},
		{"author": {"__typename": "User", "login": "carol"}, "body": "hm", "submittedAt": "2024-03-10T13:00:00Z", "state": "COMMENTED"}
	]},
	"reviewThreads": {"nodes": [
		{"isResolved": true, "comments": {"nodes": [
			{"author": {"__typename": "User", "login": "me"}, "body": "typo", "createdAt": "2024-03-10T09:30:00Z"},
			{"author": null, "body": "fixed", "createdAt": "2024-03-10T09:45:00Z"}
		]}}
	]}
}]}`

func parseConversationJSON(t *testing.T) PRSearchNode {
	t.Helper()
	nodes, err := parsePRSearchJSON([]byte(conversationJSON))
	if err != nil {
		t.Fatalf("parsePRSearchJSON returned error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	return nodes[0]
}

func TestParsePRSearchJSON_Conversation(t *testing.T) {
	n := parseConversationJSON(t)

	want := Conversation{
		Author:    "alice",
		CreatedAt: "2024-03-10T08:00:00Z",
		Body:      "cc @me",
		Comments: []Comment{
			{Login: "codecov", IsBot: true, Body: "90%", At: "2024-03-10T11:00:00Z"},
			{Login: "bob", Body: "nice", At: "2024-03-10T12:00:00Z"},
		},
		// The PENDING review is dropped: it is not submitted yet.
		Reviews: []Review{
			{Comment: Comment{Login: "carol", Body: "hm", At: "2024-03-10T13:00:00Z"}, State: "COMMENTED"},
		},
		Threads: []ReviewThread{
			{IsResolved: true, Comments: []Comment{
				{Login: "me", Body: "typo", At: "2024-03-10T09:30:00Z"},
				{Body: "fixed", At: "2024-03-10T09:45:00Z"}, // deleted account: null author
			}},
		},
		Commits: []Commit{
			{Login: "alice", At: "2024-03-10T09:00:00Z"},
			{At: "2024-03-10T10:00:00Z"},
		},
	}
	if !reflect.DeepEqual(n.Conversation, want) {
		t.Errorf("Conversation = %+v\nwant %+v", n.Conversation, want)
	}
}

func TestParsePRSearchJSON_ListFieldsComeFromTheLatestEntries(t *testing.T) {
	n := parseConversationJSON(t)

	if n.StatusState != "SUCCESS" {
		t.Errorf("StatusState = %q, want SUCCESS (from the latest commit)", n.StatusState)
	}
	want := LatestActivity{Kind: "commented", Login: "carol", At: "2024-03-10T13:00:00Z"}
	if n.LatestActivity != want {
		t.Errorf("LatestActivity = %+v, want %+v", n.LatestActivity, want)
	}
}

func TestParsePRSearchJSON_ConversationAttention(t *testing.T) {
	c := parseConversationJSON(t).Conversation

	// The description mentions "me", but the thread reply at 09:30 answered it.
	if att, ok := c.Attention("me", false); ok {
		t.Errorf("Attention(not owned) = %+v, want none: the mention was answered", att)
	}
	// Owned, carol's COMMENTED review is the newest thing since that reply; the
	// bot comment in between does not count.
	att, ok := c.Attention("me", true)
	want := Attention{Reason: AttentionCommented, Login: "carol", At: "2024-03-10T13:00:00Z"}
	if !ok || att != want {
		t.Errorf("Attention(owned) = %+v, ok = %v; want %+v", att, ok, want)
	}
}

func TestParsePRSearchNodes_LatestActivitySkipsPushWithoutUser(t *testing.T) {
	node := prSearchRawNode{Number: 1, Title: "Test"}
	node.Commits.Nodes = []rawCommitNode{{}}
	node.Commits.Nodes[0].Commit.CommittedDate = "2024-03-10T12:00:00Z"

	nodes := parsePRSearchNodes([]prSearchRawNode{node})

	if nodes[0].LatestActivity.Login != "" {
		t.Errorf("LatestActivity = %+v, want none for a commit with no GitHub user", nodes[0].LatestActivity)
	}
}

func TestPRSearchQuery_ConversationFields(t *testing.T) {
	with := prSearchQuery(true)
	without := prSearchQuery(false)

	for _, field := range []string{"reviewThreads", "body", "__typename", "comments(last: 10)"} {
		if !strings.Contains(with, field) {
			t.Errorf("prSearchQuery(true) lacks %q", field)
		}
		if strings.Contains(without, field) {
			t.Errorf("prSearchQuery(false) should not fetch %q", field)
		}
	}
	for _, q := range []string{with, without} {
		for _, field := range []string{"statusCheckRollup", "comments(last:", "reviews(last:", "reviewDecision"} {
			if !strings.Contains(q, field) {
				t.Errorf("query lacks %q:\n%s", field, q)
			}
		}
	}
}
