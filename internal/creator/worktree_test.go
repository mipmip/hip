package creator

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitEnv(t *testing.T) {
	t.Helper()
	for _, kv := range [][2]string{
		{"GIT_CONFIG_GLOBAL", "/dev/null"},
		{"GIT_CONFIG_SYSTEM", "/dev/null"},
		{"GIT_AUTHOR_NAME", "Test"},
		{"GIT_AUTHOR_EMAIL", "test@example.com"},
		{"GIT_COMMITTER_NAME", "Test"},
		{"GIT_COMMITTER_EMAIL", "test@example.com"},
	} {
		t.Setenv(kv[0], kv[1])
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func TestWorktreeCommitAndPush(t *testing.T) {
	gitEnv(t)
	root := t.TempDir()
	bare := filepath.Join(root, "remote.git")
	runGit(t, root, "init", "--bare", "-b", "main", bare)

	ctx := context.Background()
	wt, err := NewWorktree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.Cleanup()

	if err := wt.WriteREADME("# demo\n"); err != nil {
		t.Fatal(err)
	}
	if err := wt.WriteLicense("MIT\n"); err != nil {
		t.Fatal(err)
	}
	if err := wt.CommitAll(ctx, "Initial commit"); err != nil {
		t.Fatal(err)
	}
	if err := wt.Push(ctx, bare); err != nil {
		t.Fatal(err)
	}

	if got := strings.TrimSpace(runGit(t, bare, "log", "--format=%s", "main")); got != "Initial commit" {
		t.Fatalf("commit subject = %q", got)
	}
	files := runGit(t, bare, "ls-tree", "--name-only", "main")
	if !strings.Contains(files, "README.md") || !strings.Contains(files, "LICENSE") {
		t.Fatalf("files = %q", files)
	}
	if got := runGit(t, bare, "show", "main:README.md"); got != "# demo\n" {
		t.Fatalf("README = %q", got)
	}
}

func TestWorktreeCleanup(t *testing.T) {
	gitEnv(t)
	wt, err := NewWorktree(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	dir := wt.Dir()
	if err := wt.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists: %v", err)
	}
	if err := wt.Cleanup(); err != nil {
		t.Fatalf("second cleanup: %v", err)
	}
}

type failRunner struct {
	failOn string
}

func (r failRunner) Run(ctx context.Context, dir string, args ...string) error {
	if strings.Join(args, " ") == r.failOn {
		return errors.New("boom")
	}
	return execRunner{}.Run(ctx, dir, args...)
}

func TestNewWorktreeInitError(t *testing.T) {
	_, err := NewWorktreeWithRunner(context.Background(), failRunner{failOn: "init -b main"})
	if err == nil || !strings.Contains(err.Error(), "initializing temp worktree") {
		t.Fatalf("err = %v", err)
	}
}

func TestWorktreeCommandErrors(t *testing.T) {
	gitEnv(t)
	wt, err := NewWorktree(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.Cleanup()

	wt.run = failRunner{failOn: "commit -m initial"}
	if err := wt.CommitAll(context.Background(), "initial"); err == nil || !strings.Contains(err.Error(), "committing initial files") {
		t.Fatalf("commit err = %v", err)
	}

	wt.run = failRunner{failOn: "push -u origin main"}
	if err := wt.Push(context.Background(), "file:///nonexistent"); err == nil || !strings.Contains(err.Error(), "pushing main") {
		t.Fatalf("push err = %v", err)
	}
}
