// Package main is the entry point for the gh-own application.
package main

import (
	"log/slog"
	"os"

	"github.com/snrsw/gh-own/cmd"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	cmd.Execute()
}
