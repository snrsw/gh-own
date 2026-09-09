package pr

import "github.com/snrsw/gh-own/internal/gh"

// promoteNeedsAction moves every pull request whose conversation is waiting on
// the user (see gh.Conversation.Attention) into the Needs action bucket, next to
// the ones reviewers requested changes on.
//
// The state buckets — Drafts, Ready to merge, Waiting — hold pull requests the
// user owns, so a comment there counts; Participated and Review Requested hold
// other people's, where only a mention or a reply to the user counts. A promoted
// pull request leaves Drafts, Ready to merge, Waiting and Participated so that
// those tabs list only what is not waiting on the user; Review Requested keeps
// it, since the review is still due whatever else happened.
//
// A pull request the search returned in several buckets is judged once, in the
// bucket order above, so it appears in Needs action at most once.
//
// An empty login disables the promotion, which is what the demo data relies on.
func promoteNeedsAction(r *gh.PRSearchResult, login string) *gh.PRSearchResult {
	if login == "" {
		return r
	}

	needs := annotate(r.NeedsAction, login, true)
	seen := make(map[string]bool, len(needs))
	for _, n := range needs {
		seen[n.URL] = true
	}
	// take appends the nodes waiting on the user, skipping those already
	// taken so that a pull request found in several buckets is added once.
	take := func(nodes []gh.PRSearchNode, owned bool) {
		for _, n := range nodes {
			if seen[n.URL] {
				continue
			}
			n.Attention = n.Conversation.Attention(login, owned)
			if n.Attention.Reason == "" {
				continue
			}
			seen[n.URL] = true
			needs = append(needs, n)
		}
	}
	take(r.Drafts, true)
	take(r.ReadyToMerge, true)
	take(r.Waiting, true)
	take(r.Participated, false)
	take(r.ReviewRequested, false)

	out := *r
	out.NeedsAction = needs
	out.Drafts = without(r.Drafts, seen)
	out.ReadyToMerge = without(r.ReadyToMerge, seen)
	out.Waiting = without(r.Waiting, seen)
	out.Participated = without(r.Participated, seen)
	return &out
}

// annotate returns every node, with Attention filled in where it applies.
func annotate(nodes []gh.PRSearchNode, login string, owned bool) []gh.PRSearchNode {
	out := make([]gh.PRSearchNode, 0, len(nodes))
	for _, n := range nodes {
		n.Attention = n.Conversation.Attention(login, owned)
		out = append(out, n)
	}
	return out
}

func without(nodes []gh.PRSearchNode, urls map[string]bool) []gh.PRSearchNode {
	out := make([]gh.PRSearchNode, 0, len(nodes))
	for _, n := range nodes {
		if urls[n.URL] {
			continue
		}
		out = append(out, n)
	}
	return out
}
