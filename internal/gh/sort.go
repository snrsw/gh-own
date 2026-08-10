package gh

import (
	"sort"
	"time"
)

// ParseTimestamp parses a GitHub timestamp. An empty or unparsable value yields
// the zero time, which callers treat as "oldest" — same convention as mostRecent.
func ParseTimestamp(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	return time.Time{}
}

// SortAt returns the timestamp a list entry is ordered by. It mirrors what the
// description line renders: the latest activity time when there is one, and the
// item's own updatedAt otherwise, so the order matches the times on screen.
func SortAt(activity LatestActivity, updatedAt string) time.Time {
	if activity.Login != "" {
		if t := ParseTimestamp(activity.At); !t.IsZero() {
			return t
		}
	}
	return ParseTimestamp(updatedAt)
}

// SortByUpdatedDesc orders items in place, newest first, by the timestamp
// returned by at. Items with an unknown timestamp sort last. The sort is stable,
// so items sharing a timestamp keep their input order.
func SortByUpdatedDesc[T any](items []T, at func(T) time.Time) {
	sort.SliceStable(items, func(i, j int) bool {
		return at(items[i]).After(at(items[j]))
	})
}

// sortedKeys returns the map's keys in ascending order so that iterating over
// search results — which arrive in a map filled by parallel queries — produces
// a deterministic order.
func sortedKeys[T any](m map[string][]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
