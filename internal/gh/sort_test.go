package gh

import (
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   time.Time
		isZero bool
	}{
		{name: "rfc3339", input: "2024-03-15T10:00:00Z", want: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)},
		{name: "rfc3339 nano", input: "2024-03-15T10:00:00.123456789Z", want: time.Date(2024, 3, 15, 10, 0, 0, 123456789, time.UTC)},
		{name: "offset", input: "2024-03-15T12:00:00+09:00", want: time.Date(2024, 3, 15, 3, 0, 0, 0, time.UTC)},
		{name: "empty", input: "", isZero: true},
		{name: "unparsable", input: "not-a-time", isZero: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTimestamp(tt.input)
			if tt.isZero {
				if !got.IsZero() {
					t.Errorf("ParseTimestamp(%q) = %v, want zero time", tt.input, got)
				}
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseTimestamp(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSortAt(t *testing.T) {
	tests := []struct {
		name      string
		activity  LatestActivity
		updatedAt string
		want      string
	}{
		{
			name:      "uses activity time when there is activity",
			activity:  LatestActivity{Kind: "commented", Login: "alice", At: "2024-03-10T00:00:00Z"},
			updatedAt: "2024-03-15T00:00:00Z",
			want:      "2024-03-10T00:00:00Z",
		},
		{
			name:      "falls back to updatedAt without activity",
			activity:  LatestActivity{},
			updatedAt: "2024-03-15T00:00:00Z",
			want:      "2024-03-15T00:00:00Z",
		},
		{
			name:      "falls back to updatedAt when activity time is unparsable",
			activity:  LatestActivity{Kind: "commented", Login: "alice", At: "not-a-time"},
			updatedAt: "2024-03-15T00:00:00Z",
			want:      "2024-03-15T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := ParseTimestamp(tt.want)
			if got := SortAt(tt.activity, tt.updatedAt); !got.Equal(want) {
				t.Errorf("SortAt() = %v, want %v", got, want)
			}
		})
	}
}

func TestSortAt_UnknownTimestampsAreZero(t *testing.T) {
	if got := SortAt(LatestActivity{}, ""); !got.IsZero() {
		t.Errorf("SortAt() = %v, want zero time", got)
	}
}

type sortable struct {
	name string
	at   string
}

func sortableAt(s sortable) time.Time { return ParseTimestamp(s.at) }

func names(items []sortable) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.name
	}
	return out
}

func equalNames(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestSortByUpdatedDesc(t *testing.T) {
	tests := []struct {
		name  string
		input []sortable
		want  []string
	}{
		{
			name: "orders newest first",
			input: []sortable{
				{name: "old", at: "2024-03-01T00:00:00Z"},
				{name: "new", at: "2024-03-20T00:00:00Z"},
				{name: "mid", at: "2024-03-10T00:00:00Z"},
			},
			want: []string{"new", "mid", "old"},
		},
		{
			name: "already sorted stays put",
			input: []sortable{
				{name: "new", at: "2024-03-20T00:00:00Z"},
				{name: "old", at: "2024-03-01T00:00:00Z"},
			},
			want: []string{"new", "old"},
		},
		{
			name: "compares instants across offsets",
			input: []sortable{
				{name: "utc", at: "2024-03-15T10:00:00Z"},
				{name: "jst", at: "2024-03-15T12:00:00+09:00"}, // 03:00Z, older
			},
			want: []string{"utc", "jst"},
		},
		{
			name: "unknown timestamps sort last",
			input: []sortable{
				{name: "empty", at: ""},
				{name: "dated", at: "2024-03-01T00:00:00Z"},
				{name: "garbage", at: "not-a-time"},
			},
			want: []string{"dated", "empty", "garbage"},
		},
		{
			name:  "single item",
			input: []sortable{{name: "only", at: "2024-03-01T00:00:00Z"}},
			want:  []string{"only"},
		},
		{
			name:  "empty slice",
			input: []sortable{},
			want:  []string{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortByUpdatedDesc(tt.input, sortableAt)
			if got := names(tt.input); !equalNames(got, tt.want) {
				t.Errorf("SortByUpdatedDesc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortByUpdatedDesc_StableOnTies(t *testing.T) {
	items := []sortable{
		{name: "first", at: "2024-03-01T00:00:00Z"},
		{name: "second", at: "2024-03-01T00:00:00Z"},
		{name: "third", at: "2024-03-01T00:00:00Z"},
	}

	SortByUpdatedDesc(items, sortableAt)

	want := []string{"first", "second", "third"}
	if got := names(items); !equalNames(got, want) {
		t.Errorf("SortByUpdatedDesc() = %v, want %v", got, want)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string][]int{"drafts": nil, "assigned": nil, "waiting": nil}

	want := []string{"assigned", "drafts", "waiting"}
	got := sortedKeys(m)
	if len(got) != len(want) {
		t.Fatalf("sortedKeys() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("sortedKeys()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
