package reviewstatus

import (
	"strings"
	"testing"
)

func TestReviewStatus_String(t *testing.T) {
	tests := []struct {
		status ReviewStatus
		want   string
	}{
		{ReviewStatusApproved, "approved"},
		{ReviewStatusChangesRequested, "changes_requested"},
		{ReviewStatusReviewRequired, "review_required"},
		{ReviewStatusNone, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("ReviewStatus.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseReviewDecision(t *testing.T) {
	tests := []struct {
		decision string
		want     ReviewStatus
	}{
		{"APPROVED", ReviewStatusApproved},
		{"CHANGES_REQUESTED", ReviewStatusChangesRequested},
		{"REVIEW_REQUIRED", ReviewStatusReviewRequired},
		{"", ReviewStatusNone},
	}

	for _, tt := range tests {
		t.Run(tt.decision, func(t *testing.T) {
			if got := ParseReviewDecision(tt.decision); got != tt.want {
				t.Errorf("ParseReviewDecision(%q) = %v, want %v", tt.decision, got, tt.want)
			}
		})
	}
}

func TestRenderReviewStatus(t *testing.T) {
	tests := []struct {
		status   ReviewStatus
		contains string
	}{
		{ReviewStatusApproved, "✔"},
		{ReviewStatusChangesRequested, "⊘"},
		{ReviewStatusReviewRequired, "◇"},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			got := RenderReviewStatus(tt.status)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("RenderReviewStatus(%v) = %q, should contain %q", tt.status, got, tt.contains)
			}
		})
	}
}

func TestRenderReviewStatus_None(t *testing.T) {
	got := RenderReviewStatus(ReviewStatusNone)
	if got != "" {
		t.Errorf("RenderReviewStatus(None) = %q, want empty string", got)
	}
}
