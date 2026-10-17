package gh

import (
	"testing"
	"time"
)

const (
	t1 = "2024-03-10T10:00:00Z"
	t2 = "2024-03-10T11:00:00Z"
	t3 = "2024-03-10T12:00:00Z"
	t4 = "2024-03-10T13:00:00Z"
)

func comment(login, body, at string) Comment {
	return Comment{Login: login, Body: body, At: at}
}

func botComment(login, body, at string) Comment {
	return Comment{Login: login, IsBot: true, Body: body, At: at}
}

func ownPR() Conversation {
	return Conversation{Author: "me", CreatedAt: t1}
}

func TestAttention_OwnedPR(t *testing.T) {
	tests := []struct {
		name string
		conv Conversation
		want Attention
	}{
		{
			name: "no conversation",
			conv: ownPR(),
		},
		{
			name: "comment by someone else after opening",
			conv: func() Conversation {
				c := ownPR()
				c.Comments = []Comment{comment("alice", "looks odd", t2)}
				return c
			}(),
			want: Attention{Reason: AttentionCommented, Login: "alice", At: t2},
		},
		{
			name: "comment already answered by a later comment of mine",
			conv: func() Conversation {
				c := ownPR()
				c.Comments = []Comment{comment("alice", "looks odd", t2), comment("me", "fixed", t3)}
				return c
			}(),
		},
		{
			name: "comment already answered by a later push of mine",
			conv: func() Conversation {
				c := ownPR()
				c.Comments = []Comment{comment("alice", "looks odd", t2)}
				c.Commits = []Commit{{Login: "me", At: t3}}
				return c
			}(),
		},
		{
			name: "comment already answered by a later review-thread reply of mine",
			conv: func() Conversation {
				c := ownPR()
				c.Comments = []Comment{comment("alice", "looks odd", t2)}
				c.Threads = []ReviewThread{{Comments: []Comment{comment("alice", "nit", t1), comment("me", "done", t3)}}}
				return c
			}(),
		},
		{
			name: "my own comment does not count",
			conv: func() Conversation {
				c := ownPR()
				c.Comments = []Comment{comment("me", "ping", t2)}
				return c
			}(),
		},
		{
			name: "bot comment does not count",
			conv: func() Conversation {
				c := ownPR()
				c.Comments = []Comment{botComment("codecov", "coverage 90%", t2)}
				return c
			}(),
		},
		{
			name: "bot mention does count",
			conv: func() Conversation {
				c := ownPR()
				c.Comments = []Comment{botComment("release-bot", "@me please tag", t2)}
				return c
			}(),
			want: Attention{Reason: AttentionMentioned, Login: "release-bot", At: t2},
		},
		{
			name: "COMMENTED review counts as a comment",
			conv: func() Conversation {
				c := ownPR()
				c.Reviews = []Review{{Comment: comment("bob", "", t2), State: "COMMENTED"}}
				return c
			}(),
			want: Attention{Reason: AttentionCommented, Login: "bob", At: t2},
		},
		{
			name: "CHANGES_REQUESTED review counts as a comment",
			conv: func() Conversation {
				c := ownPR()
				c.Reviews = []Review{{Comment: comment("bob", "", t2), State: "CHANGES_REQUESTED"}}
				return c
			}(),
			want: Attention{Reason: AttentionCommented, Login: "bob", At: t2},
		},
		{
			name: "approval does not count even with a body",
			conv: func() Conversation {
				c := ownPR()
				c.Reviews = []Review{{Comment: comment("bob", "LGTM", t2), State: "APPROVED"}}
				return c
			}(),
		},
		{
			name: "approval that mentions me counts as a mention",
			conv: func() Conversation {
				c := ownPR()
				c.Reviews = []Review{{Comment: comment("bob", "@me merge when green", t2), State: "APPROVED"}}
				return c
			}(),
			want: Attention{Reason: AttentionMentioned, Login: "bob", At: t2},
		},
		{
			name: "review-thread comment on my PR counts as a comment",
			conv: func() Conversation {
				c := ownPR()
				c.Threads = []ReviewThread{{Comments: []Comment{comment("alice", "typo", t2)}}}
				return c
			}(),
			want: Attention{Reason: AttentionCommented, Login: "alice", At: t2},
		},
		{
			name: "resolved thread reply on my PR still counts as a comment",
			conv: func() Conversation {
				c := ownPR()
				c.Threads = []ReviewThread{{IsResolved: true, Comments: []Comment{comment("me", "why?", t2), comment("alice", "because", t3)}}}
				return c
			}(),
			want: Attention{Reason: AttentionCommented, Login: "alice", At: t3},
		},
		{
			name: "assigned PR opened by someone else with a comment I never saw",
			conv: Conversation{Author: "alice", CreatedAt: t1, Comments: []Comment{comment("bob", "hey", t2)}},
			want: Attention{Reason: AttentionCommented, Login: "bob", At: t2},
		},
		{
			name: "unparsable timestamp is ignored",
			conv: func() Conversation {
				c := ownPR()
				c.Comments = []Comment{comment("alice", "hi", "not-a-time")}
				return c
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conv.Attention("me", true); got != tt.want {
				t.Errorf("Attention = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAttention_OtherPR(t *testing.T) {
	other := Conversation{Author: "alice", CreatedAt: t1}
	tests := []struct {
		name string
		conv Conversation
		want Attention
	}{
		{
			name: "plain comment on someone else's PR does not count",
			conv: func() Conversation {
				c := other
				c.Comments = []Comment{comment("bob", "hi", t2)}
				return c
			}(),
		},
		{
			name: "mention in a comment",
			conv: func() Conversation {
				c := other
				c.Comments = []Comment{comment("bob", "cc @me", t2)}
				return c
			}(),
			want: Attention{Reason: AttentionMentioned, Login: "bob", At: t2},
		},
		{
			name: "mention in the description of a PR I never touched",
			conv: Conversation{Author: "alice", CreatedAt: t1, Body: "@me please look"},
			want: Attention{Reason: AttentionMentioned, Login: "alice", At: t1},
		},
		{
			name: "mention in the description I already answered",
			conv: Conversation{Author: "alice", CreatedAt: t1, Body: "@me please look", Comments: []Comment{comment("me", "looking", t2)}},
		},
		{
			name: "mention I already answered by commenting",
			conv: func() Conversation {
				c := other
				c.Comments = []Comment{comment("bob", "cc @me", t2), comment("me", "on it", t3)}
				return c
			}(),
		},
		{
			name: "reply in an unresolved thread I commented in",
			conv: func() Conversation {
				c := other
				c.Threads = []ReviewThread{{Comments: []Comment{comment("me", "rename this", t2), comment("alice", "why?", t3)}}}
				return c
			}(),
			want: Attention{Reason: AttentionReplied, Login: "alice", At: t3},
		},
		{
			name: "reply from a bot still counts as a reply",
			conv: func() Conversation {
				c := other
				c.Threads = []ReviewThread{{Comments: []Comment{comment("me", "rename this", t2), botComment("assistant", "done", t3)}}}
				return c
			}(),
			want: Attention{Reason: AttentionReplied, Login: "assistant", At: t3},
		},
		{
			name: "reply in a resolved thread does not count",
			conv: func() Conversation {
				c := other
				c.Threads = []ReviewThread{{IsResolved: true, Comments: []Comment{comment("me", "rename this", t2), comment("alice", "done", t3)}}}
				return c
			}(),
		},
		{
			name: "thread I never commented in does not count",
			conv: func() Conversation {
				c := other
				c.Threads = []ReviewThread{{Comments: []Comment{comment("bob", "rename this", t2), comment("alice", "done", t3)}}}
				return c
			}(),
		},
		{
			name: "reply I already answered elsewhere on the PR",
			conv: func() Conversation {
				c := other
				c.Threads = []ReviewThread{{Comments: []Comment{comment("me", "rename this", t2), comment("alice", "why?", t3)}}}
				c.Reviews = []Review{{Comment: comment("me", "", t4), State: "APPROVED"}}
				return c
			}(),
		},
		{
			name: "mention outranks a reply, whichever is newer",
			conv: func() Conversation {
				c := other
				c.Comments = []Comment{comment("bob", "@me?", t2)}
				c.Threads = []ReviewThread{{Comments: []Comment{comment("me", "rename this", t1), comment("alice", "why?", t3)}}}
				return c
			}(),
			want: Attention{Reason: AttentionMentioned, Login: "bob", At: t2},
		},
		{
			name: "among equal reasons the newest event is reported",
			conv: func() Conversation {
				c := other
				c.Comments = []Comment{comment("bob", "@me?", t2), comment("carol", "@me!", t3)}
				return c
			}(),
			want: Attention{Reason: AttentionMentioned, Login: "carol", At: t3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conv.Attention("me", false); got != tt.want {
				t.Errorf("Attention = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAttention_EmptyLogin(t *testing.T) {
	c := Conversation{Author: "alice", Comments: []Comment{comment("bob", "@ hi", t2)}}
	if att := c.Attention("", true); att != (Attention{}) {
		t.Errorf("Attention with an empty login = %+v, want none", att)
	}
}

func TestAttention_EventWithoutLoginIsIgnored(t *testing.T) {
	// A deleted account comes back as a null author; there is nobody to show.
	c := ownPR()
	c.Comments = []Comment{comment("", "@me look", t2)}
	c.Reviews = []Review{{Comment: comment("", "", t2), State: "CHANGES_REQUESTED"}}
	c.Threads = []ReviewThread{{Comments: []Comment{comment("me", "rename", t2), comment("", "done", t3)}}}
	if att := c.Attention("me", true); att != (Attention{}) {
		t.Errorf("Attention = %+v, want none: every event has an empty login", att)
	}
}

// resurrectedMention is the truncation scenario: bob mentioned "me" in a
// review, "me" answered with a comment, and n more comments by others followed.
// With n reaching the page size the answer has scrolled out of the fetched
// comments while the review is still fetched, so only the horizon can tell
// that the mention was answered.
func resurrectedMention(n int) Conversation {
	c := Conversation{Author: "alice", CreatedAt: t1}
	c.Reviews = []Review{{Comment: comment("bob", "@me?", t2), State: "COMMENTED"}}
	for i := range n {
		c.Comments = append(c.Comments, comment("carol", "chatter", laterThan(t4, i)))
	}
	return c
}

// laterThan returns a timestamp i minutes after at.
func laterThan(at string, i int) string {
	return parseTimestamp(at).Add(time.Duration(i) * time.Minute).Format(time.RFC3339)
}

func TestAttention_Horizon(t *testing.T) {
	tests := []struct {
		name  string
		conv  Conversation
		owned bool
		want  Attention
	}{
		{
			name: "answered mention beyond a full comments window is not resurrected",
			conv: resurrectedMention(conversationPageSize),
		},
		{
			name: "mention beyond a comments window that is not full still counts",
			conv: resurrectedMention(conversationPageSize - 1),
			want: Attention{Reason: AttentionMentioned, Login: "bob", At: t2},
		},
		{
			name: "mention beyond a full commits window is not resurrected",
			conv: func() Conversation {
				c := Conversation{Author: "alice", CreatedAt: t1}
				c.Comments = []Comment{comment("bob", "@me?", t2)}
				for i := range commitsPageSize {
					c.Commits = append(c.Commits, Commit{Login: "alice", At: laterThan(t3, i)})
				}
				return c
			}(),
		},
		{
			name: "the oldest entry of a full window itself still counts",
			conv: func() Conversation {
				c := ownPR()
				for i := range conversationPageSize {
					c.Comments = append(c.Comments, comment("carol", "chatter", laterThan(t2, i)))
				}
				return c
			}(),
			owned: true,
			want:  Attention{Reason: AttentionCommented, Login: "carol", At: laterThan(t2, conversationPageSize-1)},
		},
		{
			name: "a full thread window sets no horizon",
			conv: func() Conversation {
				c := Conversation{Author: "alice", CreatedAt: t1}
				c.Comments = []Comment{comment("bob", "@me?", t2)}
				for i := range conversationPageSize {
					c.Threads = append(c.Threads, ReviewThread{Comments: []Comment{comment("alice", "nit", laterThan(t3, i))}})
				}
				return c
			}(),
			want: Attention{Reason: AttentionMentioned, Login: "bob", At: t2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conv.Attention("me", tt.owned); got != tt.want {
				t.Errorf("Attention = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMentions(t *testing.T) {
	tests := []struct {
		body string
		want bool
	}{
		{"@alice", true},
		{"cc @alice please", true},
		{"@Alice", true},
		{"(@alice)", true},
		{"@alice, thanks", true},
		{"@alice-bot", false},
		{"@alice_x", false},
		{"@alice2", false},
		{"bob@alice", false},
		{"@alice/core", false},
		{"alice", false},
		{"@bob and @alice", true},
		{"@alicex @alice", true},
		{"", false},
	}
	for _, tt := range tests {
		if got := mentions(tt.body, "alice"); got != tt.want {
			t.Errorf("mentions(%q, alice) = %v, want %v", tt.body, got, tt.want)
		}
	}
	if mentions("@", "") {
		t.Error("mentions with an empty login should be false")
	}
}

func TestStripQuotedAndCode(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{"plain mention", "hey @alice", true},
		{"mention in inline code", "run `@alice`", false},
		{"mention in double-backtick code", "run ``a ` @alice``", false},
		{"mention next to inline code", "`x` @alice", true},
		{"mention in a quoted line", "> @alice said\nno", false},
		{"mention in an indented quoted line", "  > @alice said", false},
		{"mention below a quoted line", "> bob said\n@alice?", true},
		{"mention in a fence", "```\n@alice\n```\ntext", false},
		{"mention outside the fence on another line", "```\ncode\n```\n@alice", true},
		{"mention before an unterminated fence", "@alice\n```\ncode", true},
		{"unterminated backtick is literal", "`@alice", true},
		{"unterminated backtick keeps the rest of the line", "a ` b @alice", true},
		{"unterminated backtick after a span", "`x` @alice `", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mentions(tt.body, "alice"); got != tt.want {
				t.Errorf("mentions(%q, alice) = %v, want %v (stripped: %q)", tt.body, got, tt.want, stripQuotedAndCode(tt.body))
			}
		})
	}
}
