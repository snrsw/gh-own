package gh

import (
	"strings"
	"time"
)

// Attention reasons, from the strongest to the weakest. When several apply to
// one pull request the strongest is reported.
const (
	// AttentionMentioned: someone @-mentioned the user.
	AttentionMentioned = "mentioned you"
	// AttentionReplied: someone replied in an unresolved review thread the
	// user had commented in.
	AttentionReplied = "replied to you"
	// AttentionCommented: someone commented on a pull request the user owns.
	AttentionCommented = "commented"
)

// Attention records why a pull request is waiting on the user: what happened,
// who did it, and when. Its zero value means nothing is waiting.
type Attention struct {
	Reason string
	Login  string
	At     string
}

// Attention reports whether the conversation is waiting on the user and why.
//
// Only what happened after the user's last activity on the pull request counts
// (see lastActivityBy), so anything the user has already answered — by
// commenting, reviewing, or pushing — drops out. Three things wait on the user:
//
//   - an @-mention of the user in the description, a comment, a review, or a
//     review-thread comment, on any pull request;
//   - a reply in an unresolved review thread the user had commented in, on any
//     pull request;
//   - a comment, or a COMMENTED / CHANGES_REQUESTED review, by a person on a
//     pull request the user owns (owned is true). Bot comments are ignored
//     here, since CI and coverage bots comment on nearly every push;
//     approvals are ignored too, they are what Ready to merge is for.
func (c Conversation) Attention(login string, owned bool) (Attention, bool) {
	if login == "" {
		return Attention{}, false
	}
	found := attentionFinder{login: login, owned: owned, since: c.lastActivityBy(login)}

	if c.Author != login && mentions(c.Body, login) {
		found.add(AttentionMentioned, c.Author, c.CreatedAt)
	}
	for _, cm := range c.Comments {
		found.comment(cm, true)
	}
	for _, r := range c.Reviews {
		found.review(r)
	}
	for _, th := range c.Threads {
		found.thread(th)
	}
	return found.best, found.best.Reason != ""
}

// attentionFinder accumulates the strongest attention reason across the events
// of one conversation.
type attentionFinder struct {
	login string
	owned bool
	// since is the user's last activity; only events after it are considered.
	since time.Time
	best  Attention
}

// add records a reason if it beats the current best: a stronger reason wins,
// and among equal reasons the most recent event wins. Events the user wrote,
// or that predate the user's last activity, are ignored.
func (f *attentionFinder) add(reason, login, at string) {
	if login == f.login {
		return
	}
	t := parseTimestamp(at)
	if t.IsZero() || !t.After(f.since) {
		return
	}
	if attentionRank(reason) < attentionRank(f.best.Reason) {
		return
	}
	if reason == f.best.Reason && !t.After(parseTimestamp(f.best.At)) {
		return
	}
	f.best = Attention{Reason: reason, Login: login, At: at}
}

// comment handles an issue comment or a review-thread comment. countsAsComment
// says whether it may count as a plain comment on an owned pull request.
func (f *attentionFinder) comment(cm Comment, countsAsComment bool) {
	if mentions(cm.Body, f.login) {
		f.add(AttentionMentioned, cm.Login, cm.At)
	}
	if countsAsComment && f.owned && !cm.IsBot {
		f.add(AttentionCommented, cm.Login, cm.At)
	}
}

func (f *attentionFinder) review(r Review) {
	counts := r.State == "COMMENTED" || r.State == "CHANGES_REQUESTED"
	f.comment(r.Comment, counts)
}

func (f *attentionFinder) thread(th ReviewThread) {
	mine := false
	for _, cm := range th.Comments {
		if cm.Login == f.login {
			mine = true
			continue
		}
		f.comment(cm, true)
		if mine && !th.IsResolved {
			f.add(AttentionReplied, cm.Login, cm.At)
		}
	}
}

func attentionRank(reason string) int {
	switch reason {
	case AttentionMentioned:
		return 3
	case AttentionReplied:
		return 2
	case AttentionCommented:
		return 1
	default:
		return 0
	}
}

// mentions reports whether body @-mentions the login. The match is
// case-insensitive, as GitHub logins are, and must stand on its own: "@alice"
// matches, "@alice-bot", "bob@alice" and the team "@alice/core" do not.
func mentions(body, login string) bool {
	if login == "" {
		return false
	}
	body = strings.ToLower(body)
	needle := "@" + strings.ToLower(login)
	for from := 0; ; {
		idx := strings.Index(body[from:], needle)
		if idx < 0 {
			return false
		}
		start := from + idx
		end := start + len(needle)
		if isMentionBoundary(body, start-1) && isMentionBoundary(body, end) && (end >= len(body) || body[end] != '/') {
			return true
		}
		from = start + 1
	}
}

// isMentionBoundary reports whether position i in body is outside the text or
// holds a character that cannot be part of a login.
func isMentionBoundary(body string, i int) bool {
	if i < 0 || i >= len(body) {
		return true
	}
	return !isLoginChar(body[i])
}

func isLoginChar(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	case b == '-', b == '_':
		return true
	default:
		return false
	}
}
