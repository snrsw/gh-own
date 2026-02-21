package gh

import "testing"

func TestNewLatestActivity_NoActivity(t *testing.T) {
	got := NewLatestActivity("", "", "", "", "", "", "")
	if got.Login != "" {
		t.Errorf("Login = %q, want empty", got.Login)
	}
	if got.Kind != "" {
		t.Errorf("Kind = %q, want empty", got.Kind)
	}
}

func TestNewLatestActivity_CommentOnly(t *testing.T) {
	got := NewLatestActivity("alice", "2024-03-10T10:00:00Z", "", "", "", "", "")
	if got.Kind != "commented" {
		t.Errorf("Kind = %q, want %q", got.Kind, "commented")
	}
	if got.Login != "alice" {
		t.Errorf("Login = %q, want %q", got.Login, "alice")
	}
}

func TestNewLatestActivity_ReviewApproved(t *testing.T) {
	got := NewLatestActivity("", "", "bob", "2024-03-10T10:00:00Z", "APPROVED", "", "")
	if got.Kind != "approved" {
		t.Errorf("Kind = %q, want %q", got.Kind, "approved")
	}
	if got.Login != "bob" {
		t.Errorf("Login = %q, want %q", got.Login, "bob")
	}
}

func TestNewLatestActivity_ReviewChangesRequested(t *testing.T) {
	got := NewLatestActivity("", "", "carol", "2024-03-10T10:00:00Z", "CHANGES_REQUESTED", "", "")
	if got.Kind != "changes requested" {
		t.Errorf("Kind = %q, want %q", got.Kind, "changes requested")
	}
	if got.Login != "carol" {
		t.Errorf("Login = %q, want %q", got.Login, "carol")
	}
}

func TestNewLatestActivity_ReviewDismissed(t *testing.T) {
	got := NewLatestActivity("", "", "dave", "2024-03-10T10:00:00Z", "DISMISSED", "", "")
	if got.Kind != "dismissed" {
		t.Errorf("Kind = %q, want %q", got.Kind, "dismissed")
	}
	if got.Login != "dave" {
		t.Errorf("Login = %q, want %q", got.Login, "dave")
	}
}

func TestNewLatestActivity_PushOnly(t *testing.T) {
	got := NewLatestActivity("", "", "", "", "", "eve", "2024-03-10T10:00:00Z")
	if got.Kind != "pushed" {
		t.Errorf("Kind = %q, want %q", got.Kind, "pushed")
	}
	if got.Login != "eve" {
		t.Errorf("Login = %q, want %q", got.Login, "eve")
	}
}

func TestNewLatestActivity_CommentMoreRecentThanReview(t *testing.T) {
	got := NewLatestActivity(
		"alice", "2024-03-10T12:00:00Z",
		"bob", "2024-03-10T10:00:00Z", "APPROVED",
		"", "",
	)
	if got.Kind != "commented" {
		t.Errorf("Kind = %q, want %q", got.Kind, "commented")
	}
	if got.Login != "alice" {
		t.Errorf("Login = %q, want %q", got.Login, "alice")
	}
}

func TestNewLatestActivity_ReviewMoreRecentThanComment(t *testing.T) {
	got := NewLatestActivity(
		"alice", "2024-03-10T10:00:00Z",
		"bob", "2024-03-10T12:00:00Z", "APPROVED",
		"", "",
	)
	if got.Kind != "approved" {
		t.Errorf("Kind = %q, want %q", got.Kind, "approved")
	}
	if got.Login != "bob" {
		t.Errorf("Login = %q, want %q", got.Login, "bob")
	}
}
