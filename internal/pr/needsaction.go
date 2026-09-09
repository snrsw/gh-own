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
// An empty login disables the promotion, which is what the demo data relies on.
func promoteNeedsAction(r *gh.PRSearchResult, login string) *gh.PRSearchResult {
	if login == "" {
		return r
	}

	needs := annotate(r.NeedsAction, login, true)
	needs = append(needs, waitingOnUser(r.Drafts, login, true)...)
	needs = append(needs, waitingOnUser(r.ReadyToMerge, login, true)...)
	needs = append(needs, waitingOnUser(r.Waiting, login, true)...)
	needs = append(needs, waitingOnUser(r.Participated, login, false)...)
	needs = append(needs, waitingOnUser(r.ReviewRequested, login, false)...)
	needs = dedupByURL(needs)

	promoted := make(map[string]bool, len(needs))
	for _, n := range needs {
		promoted[n.URL] = true
	}

	out := *r
	out.NeedsAction = needs
	out.Drafts = without(r.Drafts, promoted)
	out.ReadyToMerge = without(r.ReadyToMerge, promoted)
	out.Waiting = without(r.Waiting, promoted)
	out.Participated = without(r.Participated, promoted)
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

// waitingOnUser returns only the nodes waiting on the user, with Attention
// filled in.
func waitingOnUser(nodes []gh.PRSearchNode, login string, owned bool) []gh.PRSearchNode {
	var out []gh.PRSearchNode
	for _, n := range nodes {
		n.Attention = n.Conversation.Attention(login, owned)
		if n.Attention.Reason == "" {
			continue
		}
		out = append(out, n)
	}
	return out
}

func dedupByURL(nodes []gh.PRSearchNode) []gh.PRSearchNode {
	seen := make(map[string]bool, len(nodes))
	out := make([]gh.PRSearchNode, 0, len(nodes))
	for _, n := range nodes {
		if seen[n.URL] {
			continue
		}
		seen[n.URL] = true
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
