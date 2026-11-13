package creator

import (
	"context"
	"fmt"

	"github.com/mipmip/hip/internal/provider"
)

// RemoteCreator creates the remote repository.
type RemoteCreator interface {
	CreateRepo(ctx context.Context, opts provider.CreateRepoOptions) (provider.Repo, error)
}

// Store adds a created repository to the cache.
type Store interface {
	AddRepo(providerName string, repo provider.Repo) error
}

// WorktreeHandle is the temporary local repository used for the initial commit.
type WorktreeHandle interface {
	Dir() string
	WriteREADME(content string) error
	WriteLicense(content string) error
	CommitAll(ctx context.Context, message string) error
	Push(ctx context.Context, remoteURL string) error
	Cleanup() error
}

// Request describes one repository creation.
type Request struct {
	ProviderName string
	Options      provider.CreateRepoOptions
	README       string
	License      string
}

// Result is the created repository.
type Result struct {
	Repo provider.Repo
}

// PushError reports a failed initial push and the worktree that was kept for
// recovery.
type PushError struct {
	Worktree string
	Err      error
}

func (e *PushError) Error() string {
	return fmt.Sprintf("pushing initial commit: %v (worktree kept at %s)", e.Err, e.Worktree)
}

func (e *PushError) Unwrap() error { return e.Err }

// Engine creates remotes, pushes an optional initial commit, and updates the
// cache.
type Engine struct {
	creator     RemoteCreator
	store       Store
	newWorktree func(context.Context) (WorktreeHandle, error)
	cloneURL    func(provider.Repo) string
}

// NewEngine builds an Engine. cloneURL selects the SSH or HTTPS remote URL.
func NewEngine(creator RemoteCreator, store Store, cloneURL func(provider.Repo) string) *Engine {
	if cloneURL == nil {
		cloneURL = func(r provider.Repo) string { return r.SSHURL }
	}
	return &Engine{
		creator: creator,
		store:   store,
		newWorktree: func(ctx context.Context) (WorktreeHandle, error) {
			return NewWorktree(ctx)
		},
		cloneURL: cloneURL,
	}
}

// Create runs the creation pipeline. The worktree is prepared before the
// remote exists so editor and commit failures cannot leave an orphan remote.
func (e *Engine) Create(ctx context.Context, req Request) (Result, error) {
	hasContent := req.README != "" || req.License != ""
	var wt WorktreeHandle
	if hasContent {
		var err error
		wt, err = e.newWorktree(ctx)
		if err != nil {
			return Result{}, fmt.Errorf("preparing worktree: %w", err)
		}
		if req.README != "" {
			if err := wt.WriteREADME(req.README); err != nil {
				_ = wt.Cleanup()
				return Result{}, fmt.Errorf("writing README: %w", err)
			}
		}
		if req.License != "" {
			if err := wt.WriteLicense(req.License); err != nil {
				_ = wt.Cleanup()
				return Result{}, fmt.Errorf("writing LICENSE: %w", err)
			}
		}
		if err := wt.CommitAll(ctx, "Initial commit"); err != nil {
			_ = wt.Cleanup()
			return Result{}, err
		}
	}

	repo, err := e.creator.CreateRepo(ctx, req.Options)
	if err != nil {
		if wt != nil {
			_ = wt.Cleanup()
		}
		return Result{}, fmt.Errorf("creating repository: %w", err)
	}

	if hasContent {
		if err := wt.Push(ctx, e.cloneURL(repo)); err != nil {
			return Result{}, &PushError{Worktree: wt.Dir(), Err: err}
		}
		_ = wt.Cleanup()
	}

	if err := e.store.AddRepo(req.ProviderName, repo); err != nil {
		return Result{Repo: repo}, fmt.Errorf("adding repository to cache: %w", err)
	}
	return Result{Repo: repo}, nil
}
