package pr

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/snrsw/gh-own/internal/gh"
)

func ownedNode(num int, comments ...gh.Comment) gh.PRSearchNode {
	n := gh.PRSearchNode{Number: num, URL: "https://github.com/o/r/pull/" + strconv.Itoa(num)}
	n.Author.Login = "me"
	n.Conversation = gh.Conversation{Author: "me", CreatedAt: "2024-03-10T08:00:00Z", Comments: comments}
	return n
}

func otherNode(num int, threads ...gh.ReviewThread) gh.PRSearchNode {
	n := gh.PRSearchNode{Number: num, URL: "https://github.com/o/r/pull/" + strconv.Itoa(num)}
	n.Author.Login = "alice"
	n.Conversation = gh.Conversation{Author: "alice", CreatedAt: "2024-03-10T08:00:00Z", Threads: threads}
	return n
}

func numbers(nodes []gh.PRSearchNode) []int {
	out := make([]int, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n.Number)
	}
	return out
}

func TestPromoteNeedsAction_MovesOwnedPRsWithNewComments(t *testing.T) {
	newComment := gh.Comment{Login: "bob", Body: "hm", At: "2024-03-10T09:00:00Z"}
	r := &gh.PRSearchResult{
		NeedsAction:  []gh.PRSearchNode{ownedNode(1)},
		Drafts:       []gh.PRSearchNode{ownedNode(2, newComment), ownedNode(3)},
		ReadyToMerge: []gh.PRSearchNode{ownedNode(4, newComment), ownedNode(5)},
		Waiting:      []gh.PRSearchNode{ownedNode(6, newComment), ownedNode(7)},
	}

	got := promoteNeedsAction(r, "me")

	if !slices.Equal(numbers(got.NeedsAction), []int{1, 2, 4, 6}) {
		t.Errorf("NeedsAction = %v, want [1 2 4 6]", numbers(got.NeedsAction))
	}
	if !slices.Equal(numbers(got.Drafts), []int{3}) {
		t.Errorf("Drafts = %v, want [3]", numbers(got.Drafts))
	}
	if !slices.Equal(numbers(got.ReadyToMerge), []int{5}) {
		t.Errorf("ReadyToMerge = %v, want [5]", numbers(got.ReadyToMerge))
	}
	if !slices.Equal(numbers(got.Waiting), []int{7}) {
		t.Errorf("Waiting = %v, want [7]", numbers(got.Waiting))
	}
	for _, n := range got.NeedsAction[1:] {
		if n.Attention.Reason != gh.AttentionCommented || n.Attention.Login != "bob" {
			t.Errorf("#%d Attention = %+v, want commented by bob", n.Number, n.Attention)
		}
	}
	if got.NeedsAction[0].Attention.Reason != "" {
		t.Errorf("#1 Attention = %+v, want none: nobody commented", got.NeedsAction[0].Attention)
	}
}

func TestPromoteNeedsAction_OtherPeoplesPRs(t *testing.T) {
	reply := gh.ReviewThread{Comments: []gh.Comment{
		{Login: "me", Body: "rename", At: "2024-03-10T09:00:00Z"},
		{Login: "alice", Body: "why?", At: "2024-03-10T10:00:00Z"},
	}}
	unrelated := gh.ReviewThread{Comments: []gh.Comment{
		{Login: "bob", Body: "rename", At: "2024-03-10T09:00:00Z"},
		{Login: "alice", Body: "why?", At: "2024-03-10T10:00:00Z"},
	}}
	r := &gh.PRSearchResult{
		Participated:    []gh.PRSearchNode{otherNode(1, reply), otherNode(2, unrelated)},
		ReviewRequested: []gh.PRSearchNode{otherNode(3, reply), otherNode(4, unrelated)},
	}

	got := promoteNeedsAction(r, "me")

	if !slices.Equal(numbers(got.NeedsAction), []int{1, 3}) {
		t.Errorf("NeedsAction = %v, want [1 3]", numbers(got.NeedsAction))
	}
	if !slices.Equal(numbers(got.Participated), []int{2}) {
		t.Errorf("Participated = %v, want [2]: a promoted PR leaves Participated", numbers(got.Participated))
	}
	if !slices.Equal(numbers(got.ReviewRequested), []int{3, 4}) {
		t.Errorf("ReviewRequested = %v, want [3 4]: the review is still due", numbers(got.ReviewRequested))
	}
	if got.NeedsAction[0].Attention.Reason != gh.AttentionReplied {
		t.Errorf("Attention = %+v, want replied", got.NeedsAction[0].Attention)
	}
}

func TestPromoteNeedsAction_SamePRInSeveralBuckets(t *testing.T) {
	// A team query can return one of the user's own PRs into Participated as
	// well. It must appear once under Needs action and vanish from Participated,
	// even though a plain comment would not have promoted it from there.
	newComment := gh.Comment{Login: "bob", Body: "hm", At: "2024-03-10T09:00:00Z"}
	own := ownedNode(1, newComment)
	r := &gh.PRSearchResult{
		Waiting:      []gh.PRSearchNode{own},
		Participated: []gh.PRSearchNode{own},
	}

	got := promoteNeedsAction(r, "me")

	if !slices.Equal(numbers(got.NeedsAction), []int{1}) {
		t.Errorf("NeedsAction = %v, want [1]", numbers(got.NeedsAction))
	}
	if len(got.Waiting) != 0 || len(got.Participated) != 0 {
		t.Errorf("Waiting = %v, Participated = %v, want both empty", numbers(got.Waiting), numbers(got.Participated))
	}
}

func TestPromoteNeedsAction_EmptyLoginIsNoop(t *testing.T) {
	newComment := gh.Comment{Login: "bob", Body: "hm", At: "2024-03-10T09:00:00Z"}
	r := &gh.PRSearchResult{Waiting: []gh.PRSearchNode{ownedNode(1, newComment)}}

	got := promoteNeedsAction(r, "")

	if len(got.NeedsAction) != 0 || len(got.Waiting) != 1 {
		t.Errorf("promotion ran without a login: NeedsAction = %v, Waiting = %v", numbers(got.NeedsAction), numbers(got.Waiting))
	}
}

func TestNewGroupedPullRequests_PromotesAndDescribes(t *testing.T) {
	newComment := gh.Comment{Login: "bob", Body: "@me look", At: "2024-03-10T09:00:00Z"}
	r := &gh.PRSearchResult{
		Waiting: []gh.PRSearchNode{ownedNode(1, newComment), ownedNode(2)},
		Custom:  map[string][]gh.PRSearchNode{},
	}

	grouped := NewGroupedPullRequests(r, "me", true)

	if grouped.NeedsAction.TotalCount != 1 || grouped.Waiting.TotalCount != 1 {
		t.Fatalf("NeedsAction = %d, Waiting = %d, want 1 and 1", grouped.NeedsAction.TotalCount, grouped.Waiting.TotalCount)
	}
	item := grouped.NeedsAction.Items[0]
	if item.Attention.Reason != gh.AttentionMentioned {
		t.Errorf("Attention = %+v, want mentioned", item.Attention)
	}
	if !item.SortAt.Equal(gh.SortAt(gh.LatestActivity{Login: "bob", At: newComment.At}, "")) {
		t.Errorf("SortAt = %v, want the mention time %s", item.SortAt, newComment.At)
	}
	desc := item.toItem("me").Description()
	if !strings.Contains(desc, "mentioned you by @bob") {
		t.Errorf("Description() = %q, should contain %q", desc, "mentioned you by @bob")
	}
}

func TestNewGroupedPullRequests_ConversationOff(t *testing.T) {
	newComment := gh.Comment{Login: "bob", Body: "@me look", At: "2024-03-10T09:00:00Z"}
	r := &gh.PRSearchResult{
		Waiting: []gh.PRSearchNode{ownedNode(1, newComment)},
		Custom:  map[string][]gh.PRSearchNode{},
	}
	grouped := NewGroupedPullRequests(r, "me", false)

	if grouped.NeedsAction.TotalCount != 0 || grouped.Waiting.TotalCount != 1 {
		t.Errorf("NeedsAction = %d, Waiting = %d; want 0 and 1 with the conversation off", grouped.NeedsAction.TotalCount, grouped.Waiting.TotalCount)
	}
}
