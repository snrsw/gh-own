// Package timing provides helpers for measuring and logging execution time of stages.
package timing

import (
	"log/slog"
	"time"
)

func Track(name string) func() {
	start := time.Now()
	slog.Debug(name, "status", "start")
	return func() {
		slog.Debug(name, "status", "done", "elapsed", time.Since(start))
	}
}
