// Package ui provides the user interface components for the application.
package ui

import (
	"fmt"
	"time"
)

func UpdatedAgo(updatedAt string) string {
	if updatedAt == "" {
		return "-"
	}

	// updated_at of GitHub API is usually RFC3339
	t, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		// Sometimes it comes with nanoseconds, so as a precaution
		t2, err2 := time.Parse(time.RFC3339Nano, updatedAt)
		if err2 != nil {
			return updatedAt // If it cannot be parsed, return the original string
		}
		t = t2
	}

	d := time.Since(t)
	if d < 0 {
		// If it's in the future, round it and display
		d = -d
		return "in " + humanizeDuration(d)
	}
	return humanizeDuration(d) + " ago"
}

func CreatedOn(createdAt string) string {
	if createdAt == "" {
		return "-"
	}

	// created_at of GitHub API is usually RFC3339
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		// Sometimes it comes with nanoseconds, so as a precaution
		t2, err2 := time.Parse(time.RFC3339Nano, createdAt)
		if err2 != nil {
			return createdAt // If it cannot be parsed, return the original string
		}
		t = t2
	}

	return t.Format("2006-01-02")
}

func humanizeDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
}
