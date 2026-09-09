package gh

import "time"

// Page sizes of the conversation lists fetched by the search query. Both the
// query (prConversationFields) and the truncation horizon (Conversation.horizon)
// are built from them, so they cannot drift apart.
const (
	// conversationPageSize is how many comments, reviews, review threads, and
	// comments per thread are fetched.
	conversationPageSize = 10
	// commitsPageSize is how many commits are fetched.
	commitsPageSize = 5
)

// Conversation is the recent discussion on a pull request: the description,
// the latest comments and reviews, the latest inline review threads, and the
// latest commits. It is what Attention reads to decide whether the discussion
// is waiting on the user.
//
// Every list is in chronological order, oldest first, and holds only the tail
// of the full history — the search query fetches a fixed number of each.
type Conversation struct {
	// Author is the login of the pull request author.
	Author string
	// CreatedAt is when the pull request was opened.
	CreatedAt string
	// Body is the pull request description.
	Body string
	// Comments are the issue-style comments on the pull request.
	Comments []Comment
	// Reviews are the submitted reviews.
	Reviews []Review
	// Threads are the inline review threads.
	Threads []ReviewThread
	// Commits are the commits on the pull request branch.
	Commits []Commit
}

// Comment is an issue comment or an inline review-thread comment.
type Comment struct {
	Login string
	// IsBot is true when the author is a GitHub App rather than a person.
	IsBot bool
	Body  string
	At    string
}

// Review is a submitted pull request review. Its State is GitHub's enum:
// APPROVED, CHANGES_REQUESTED, COMMENTED, DISMISSED or PENDING.
type Review struct {
	Comment
	State string
}

// ReviewThread is an inline review thread with its comments, oldest first.
type ReviewThread struct {
	IsResolved bool
	Comments   []Comment
}

// Commit is a commit on the pull request branch.
type Commit struct {
	// Login is the GitHub user the commit author maps to; empty when unknown.
	Login string
	At    string
}

// lastActivityBy returns the time of the user's most recent activity on the
// pull request: opening it, commenting, reviewing, replying in a thread, or
// pushing a commit. The zero time means the user has never touched it.
func (c Conversation) lastActivityBy(login string) time.Time {
	var last time.Time
	bump := func(actor, at string) {
		if actor != login {
			return
		}
		if t := parseTimestamp(at); t.After(last) {
			last = t
		}
	}

	bump(c.Author, c.CreatedAt)
	for _, cm := range c.Comments {
		bump(cm.Login, cm.At)
	}
	for _, r := range c.Reviews {
		bump(r.Login, r.At)
	}
	for _, th := range c.Threads {
		for _, cm := range th.Comments {
			bump(cm.Login, cm.At)
		}
	}
	for _, cm := range c.Commits {
		bump(cm.Login, cm.At)
	}
	return last
}

// horizon returns the time before which events cannot be judged: the lists
// hold only the tail of the history, and once a list is full, an activity of
// the user older than its oldest fetched entry may have scrolled out of view.
// Anything that happened before that entry might already have been answered by
// such a hidden activity, so Attention ignores it. The horizon is the newest of
// the oldest entries of the full lists among comments, reviews, and commits;
// the zero time means no list is full and nothing is hidden.
//
// Review threads are left out on purpose: threads are ordered by creation, not
// by activity, so a full thread window says nothing about which replies are
// hidden, and a thread's own comment window rarely fills up.
//
// Reviews are counted after the PENDING ones were dropped, so a page holding a
// pending review is not seen as full; the horizon then errs on the permissive
// side, which is the harmless direction.
func (c Conversation) horizon() time.Time {
	var h time.Time
	bump := func(at string) {
		if t := parseTimestamp(at); t.After(h) {
			h = t
		}
	}
	if len(c.Comments) >= conversationPageSize {
		bump(c.Comments[0].At)
	}
	if len(c.Reviews) >= conversationPageSize {
		bump(c.Reviews[0].At)
	}
	if len(c.Commits) >= commitsPageSize {
		bump(c.Commits[0].At)
	}
	return h
}

// latestActivity picks the most recent comment, review, and push out of the
// conversation and lets NewLatestActivity choose between them.
func (c Conversation) latestActivity() LatestActivity {
	var commentLogin, commentAt string
	if n := len(c.Comments); n > 0 {
		commentLogin, commentAt = c.Comments[n-1].Login, c.Comments[n-1].At
	}
	var reviewLogin, reviewAt, reviewState string
	if n := len(c.Reviews); n > 0 {
		reviewLogin, reviewAt, reviewState = c.Reviews[n-1].Login, c.Reviews[n-1].At, c.Reviews[n-1].State
	}
	var pushLogin, pushAt string
	if n := len(c.Commits); n > 0 && c.Commits[n-1].Login != "" {
		pushLogin, pushAt = c.Commits[n-1].Login, c.Commits[n-1].At
	}
	return NewLatestActivity(commentLogin, commentAt, reviewLogin, reviewAt, reviewState, pushLogin, pushAt)
}
