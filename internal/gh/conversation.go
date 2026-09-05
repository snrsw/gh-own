package gh

import "time"

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
