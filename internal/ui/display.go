// Package ui provides the user interface components for the application.
package ui

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var userStyle = lipgloss.NewStyle().Foreground(colorUser)

// RenderUser returns "@login" highlighted if login != currentLogin, plain otherwise.
func RenderUser(login, currentLogin string) string {
	if login == currentLogin {
		return "@" + login
	}
	return userStyle.Render("@" + login)
}

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
	h := humanizeDuration(d)
	if h == "just now" {
		return h
	}
	return h + " ago"
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

// HumanizeTabName converts a kebab-case key like "needs-triage" to "Needs Triage".
func HumanizeTabName(s string) string {
	if s == "" {
		return ""
	}
	words := strings.Split(s, "-")
	for i, w := range words {
		if w == "" {
			continue
		}
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
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
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	}
	if d < 365*24*time.Hour {
		return fmt.Sprintf("%dmo", int(d.Hours()/(24*30)))
	}
	return fmt.Sprintf("%dy", int(d.Hours()/(24*365)))
}
