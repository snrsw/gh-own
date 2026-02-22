// Package checkout provides functions to checkout a PR branch in a local repository.
package checkout

import (
	"fmt"
	"strings"
)

// RunFunc executes a command with the given name, args, and working directory.
type RunFunc func(name string, args []string, dir string) error

// OutputFunc executes a command and returns its stdout.
type OutputFunc func(name string, args []string, dir string) (string, error)

// FindRepoDir finds the local repository path using ghq.
func FindRepoDir(repoName string, run OutputFunc) (string, error) {
	out, err := run("ghq", []string{"list", "--full-path", "-e", repoName}, "")
	if err != nil {
		return "", fmt.Errorf("ghq list: %w", err)
	}
	dir := strings.TrimSpace(out)
	if dir == "" {
		return "", fmt.Errorf("ghq: repository %q not found", repoName)
	}
	return dir, nil
}

// Checkout runs `gh pr checkout <number>` in the given directory.
func Checkout(repoDir string, number int, run RunFunc) error {
	return run("gh", []string{"pr", "checkout", fmt.Sprint(number)}, repoDir)
}
