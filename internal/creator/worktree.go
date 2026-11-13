package creator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner runs a git command in dir.
type Runner interface {
	Run(ctx context.Context, dir string, args ...string) error
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Worktree is a temporary plain-git repository used to create the initial
// commit before the remote repository exists.
type Worktree struct {
	dir string
	run Runner
}

// NewWorktree creates a temporary directory and initializes a git repository
// with main as the initial branch.
func NewWorktree(ctx context.Context) (*Worktree, error) {
	return NewWorktreeWithRunner(ctx, execRunner{})
}

// NewWorktreeWithRunner is NewWorktree with an injectable git runner.
func NewWorktreeWithRunner(ctx context.Context, run Runner) (*Worktree, error) {
	dir, err := os.MkdirTemp("", "hip-create-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp worktree: %w", err)
	}
	wt := &Worktree{dir: dir, run: run}
	if err := run.Run(ctx, dir, "init", "-b", "main"); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("initializing temp worktree: %w", err)
	}
	return wt, nil
}

// Dir returns the worktree directory.
func (w *Worktree) Dir() string { return w.dir }

// WriteREADME writes README.md.
func (w *Worktree) WriteREADME(content string) error {
	return w.writeFile("README.md", content)
}

// WriteLicense writes LICENSE.
func (w *Worktree) WriteLicense(content string) error {
	return w.writeFile("LICENSE", content)
}

func (w *Worktree) writeFile(name, content string) error {
	if err := os.WriteFile(filepath.Join(w.dir, name), []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", name, err)
	}
	return nil
}

// CommitAll stages every file and creates the initial commit.
func (w *Worktree) CommitAll(ctx context.Context, message string) error {
	if err := w.run.Run(ctx, w.dir, "add", "-A"); err != nil {
		return fmt.Errorf("staging initial files: %w", err)
	}
	if err := w.run.Run(ctx, w.dir, "commit", "-m", message); err != nil {
		return fmt.Errorf("committing initial files: %w", err)
	}
	return nil
}

// Push adds origin and pushes main.
func (w *Worktree) Push(ctx context.Context, remoteURL string) error {
	if err := w.run.Run(ctx, w.dir, "remote", "add", "origin", remoteURL); err != nil {
		return fmt.Errorf("adding origin: %w", err)
	}
	if err := w.run.Run(ctx, w.dir, "push", "-u", "origin", "main"); err != nil {
		return fmt.Errorf("pushing main: %w", err)
	}
	return nil
}

// Cleanup removes the worktree directory.
func (w *Worktree) Cleanup() error {
	if w == nil {
		return nil
	}
	if err := os.RemoveAll(w.dir); err != nil {
		return fmt.Errorf("removing temp worktree %s: %w", w.dir, err)
	}
	return nil
}
