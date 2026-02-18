package timing_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/snrsw/gh-own/internal/timing"
)

// captureHandler is a custom slog.Handler that stores records in a slice.
type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func (h *captureHandler) Records() []slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.records
}

func TestTrack_LogsStartMessage(t *testing.T) {
	h := &captureHandler{}
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(slog.New(slog.NewTextHandler(nil, nil))) })

	timing.Track("login")

	records := h.Records()
	if len(records) == 0 {
		t.Fatal("expected at least one log record, got none")
	}

	r := records[0]
	if r.Message != "login" {
		t.Errorf("expected message %q, got %q", "login", r.Message)
	}

	var status string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "status" {
			status = a.Value.String()
			return false
		}
		return true
	})
	if status != "start" {
		t.Errorf("expected status %q, got %q", "start", status)
	}
}

func TestTrack_DoneLogsElapsed(t *testing.T) {
	h := &captureHandler{}
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(slog.New(slog.NewTextHandler(nil, nil))) })

	done := timing.Track("login")
	done()

	records := h.Records()
	if len(records) < 2 {
		t.Fatalf("expected at least 2 log records, got %d", len(records))
	}

	r := records[1]
	if r.Message != "login" {
		t.Errorf("expected message %q, got %q", "login", r.Message)
	}

	var status string
	var hasElapsed bool
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "status":
			status = a.Value.String()
		case "elapsed":
			hasElapsed = true
		}
		return true
	})

	if status != "done" {
		t.Errorf("expected status %q, got %q", "done", status)
	}
	if !hasElapsed {
		t.Error("expected elapsed attribute, got none")
	}
}

func TestTrack_ElapsedIsPositive(t *testing.T) {
	h := &captureHandler{}
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(slog.New(slog.NewTextHandler(nil, nil))) })

	done := timing.Track("login")
	time.Sleep(10 * time.Millisecond)
	done()

	records := h.Records()
	if len(records) < 2 {
		t.Fatalf("expected at least 2 log records, got %d", len(records))
	}

	var elapsed time.Duration
	records[1].Attrs(func(a slog.Attr) bool {
		if a.Key == "elapsed" {
			d, ok := a.Value.Any().(time.Duration)
			if ok {
				elapsed = d
			}
			return false
		}
		return true
	})

	if elapsed <= 0 {
		t.Errorf("expected positive elapsed, got %v", elapsed)
	}
}
