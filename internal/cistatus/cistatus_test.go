package cistatus

import (
	"strings"
	"testing"
)

func TestCIStatus_String(t *testing.T) {
	tests := []struct {
		status CIStatus
		want   string
	}{
		{CIStatusSuccess, "success"},
		{CIStatusFailure, "failure"},
		{CIStatusPending, "pending"},
		{CIStatusNone, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("CIStatus.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseState(t *testing.T) {
	tests := []struct {
		state string
		want  CIStatus
	}{
		{"SUCCESS", CIStatusSuccess},
		{"FAILURE", CIStatusFailure},
		{"ERROR", CIStatusFailure},
		{"PENDING", CIStatusPending},
		{"EXPECTED", CIStatusPending},
		{"", CIStatusNone},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if got := ParseState(tt.state); got != tt.want {
				t.Errorf("ParseState(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestRenderCIStatus(t *testing.T) {
	tests := []struct {
		status   CIStatus
		contains string
	}{
		{CIStatusSuccess, "✓"},
		{CIStatusFailure, "✗"},
		{CIStatusPending, "●"},
		{CIStatusNone, "-"},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			got := RenderCIStatus(tt.status)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("RenderCIStatus(%v) = %q, should contain %q", tt.status, got, tt.contains)
			}
		})
	}
}
