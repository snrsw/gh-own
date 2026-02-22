package checkout

import (
	"fmt"
	"testing"
)

func TestCheckout(t *testing.T) {
	var gotName string
	var gotArgs []string
	var gotDir string

	runner := func(name string, args []string, dir string) error {
		gotName = name
		gotArgs = args
		gotDir = dir
		return nil
	}

	err := Checkout("/home/user/src/owner/repo", 42, runner)
	if err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}

	if gotName != "gh" {
		t.Errorf("command name = %q, want %q", gotName, "gh")
	}
	wantArgs := []string{"pr", "checkout", "42"}
	if fmt.Sprint(gotArgs) != fmt.Sprint(wantArgs) {
		t.Errorf("command args = %v, want %v", gotArgs, wantArgs)
	}
	if gotDir != "/home/user/src/owner/repo" {
		t.Errorf("command dir = %q, want %q", gotDir, "/home/user/src/owner/repo")
	}
}

func TestFindRepoDir(t *testing.T) {
	runner := func(name string, args []string, dir string) (string, error) {
		return "/home/user/src/github.com/owner/repo\n", nil
	}

	got, err := FindRepoDir("owner/repo", runner)
	if err != nil {
		t.Fatalf("FindRepoDir() error = %v", err)
	}
	if got != "/home/user/src/github.com/owner/repo" {
		t.Errorf("FindRepoDir() = %q, want %q", got, "/home/user/src/github.com/owner/repo")
	}
}

func TestFindRepoDir_NotFound(t *testing.T) {
	runner := func(name string, args []string, dir string) (string, error) {
		return "", fmt.Errorf("exit status 1")
	}

	_, err := FindRepoDir("owner/repo", runner)
	if err == nil {
		t.Fatal("FindRepoDir() should return error when ghq fails")
	}
}
