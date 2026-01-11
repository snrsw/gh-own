package ui

import (
	"strings"
	"testing"
	"time"
)

func TestHumanizeDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "just now"},
		{"30 seconds", 30 * time.Second, "just now"},
		{"59 seconds", 59 * time.Second, "just now"},
		{"1 minute", 1 * time.Minute, "1m"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"59 minutes", 59 * time.Minute, "59m"},
		{"1 hour", 1 * time.Hour, "1h"},
		{"5 hours", 5 * time.Hour, "5h"},
		{"23 hours", 23 * time.Hour, "23h"},
		{"24 hours", 24 * time.Hour, "1d"},
		{"2 days", 48 * time.Hour, "2d"},
		{"6 days", 6 * 24 * time.Hour, "6d"},
		{"7 days", 7 * 24 * time.Hour, "1w"},
		{"14 days", 14 * 24 * time.Hour, "2w"},
		{"30 days", 30 * 24 * time.Hour, "4w"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := humanizeDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("humanizeDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestUpdatedAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		input      string
		expected   string
		prefix     string // use prefix check instead of exact match
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "-",
		},
		{
			name:     "invalid format",
			input:    "not-a-date",
			expected: "not-a-date",
		},
		{
			name:     "just now RFC3339",
			input:    now.Format(time.RFC3339),
			expected: "just now ago",
		},
		{
			name:     "just now RFC3339Nano",
			input:    now.Format(time.RFC3339Nano),
			expected: "just now ago",
		},
		{
			name:     "5 minutes ago",
			input:    now.Add(-5 * time.Minute).Format(time.RFC3339),
			expected: "5m ago",
		},
		{
			name:     "2 hours ago",
			input:    now.Add(-2 * time.Hour).Format(time.RFC3339),
			expected: "2h ago",
		},
		{
			name:     "3 days ago",
			input:    now.Add(-3 * 24 * time.Hour).Format(time.RFC3339),
			expected: "3d ago",
		},
		{
			name:     "2 weeks ago",
			input:    now.Add(-14 * 24 * time.Hour).Format(time.RFC3339),
			expected: "2w ago",
		},
		{
			name:   "future time",
			input:  now.Add(10 * time.Minute).Format(time.RFC3339),
			prefix: "in ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UpdatedAgo(tt.input)
			if tt.prefix != "" {
				if !strings.HasPrefix(result, tt.prefix) {
					t.Errorf("UpdatedAgo(%q) = %q, want prefix %q", tt.input, result, tt.prefix)
				}
			} else if result != tt.expected {
				t.Errorf("UpdatedAgo(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreatedOn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "-",
		},
		{
			name:     "invalid format",
			input:    "not-a-date",
			expected: "not-a-date",
		},
		{
			name:     "valid RFC3339",
			input:    "2024-03-15T10:30:00Z",
			expected: "2024-03-15",
		},
		{
			name:     "valid RFC3339 with timezone",
			input:    "2024-12-25T15:00:00+09:00",
			expected: "2024-12-25",
		},
		{
			name:     "valid RFC3339Nano",
			input:    "2024-01-01T00:00:00.123456789Z",
			expected: "2024-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreatedOn(tt.input)
			if result != tt.expected {
				t.Errorf("CreatedOn(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
