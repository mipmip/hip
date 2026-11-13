package creator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mipmip/hip/internal/provider"
)

type fakeCreator struct {
	repo  provider.Repo
	err   error
	opts  provider.CreateRepoOptions
	calls int
}

func (f *fakeCreator) CreateRepo(_ context.Context, opts provider.CreateRepoOptions) (provider.Repo, error) {
	f.calls++
	f.opts = opts
	return f.repo, f.err
}

type fakeStore struct {
	err      error
	provider string
	repo     provider.Repo
	calls    int
}

func (f *fakeStore) AddRepo(providerName string, repo provider.Repo) error {
	f.calls++
	f.provider = providerName
	f.repo = repo
	return f.err
}

type fakeWorktree struct {
	dir       string
	readme    string
	license   string
	committed bool
	pushed    string
	cleaned   bool

	writeErr  error
	commitErr error
	pushErr   error
}

func (f *fakeWorktree) Dir() string { return f.dir }

func (f *fakeWorktree) WriteREADME(content string) error {
	f.readme = content
	return f.writeErr
}

func (f *fakeWorktree) WriteLicense(content string) error {
	f.license = content
	return f.writeErr
}

func (f *fakeWorktree) CommitAll(context.Context, string) error {
	f.committed = true
	return f.commitErr
}

func (f *fakeWorktree) Push(_ context.Context, remoteURL string) error {
	f.pushed = remoteURL
	return f.pushErr
}

func (f *fakeWorktree) Cleanup() error {
	f.cleaned = true
	return nil
}

func newFakeEngine(t *testing.T, fc *fakeCreator, fs *fakeStore, fw *fakeWorktree) *Engine {
	t.Helper()
	eng := NewEngine(fc, fs, func(r provider.Repo) string { return r.SSHURL })
	eng.newWorktree = func(context.Context) (WorktreeHandle, error) { return fw, nil }
	return eng
}

func TestCreatePushesContentAndAddsRepo(t *testing.T) {
	repo := provider.Repo{Owner: "o", Name: "r", SSHURL: "git@example.com:o/r.git"}
	fc := &fakeCreator{repo: repo}
	fs := &fakeStore{}
	fw := &fakeWorktree{dir: "/tmp/hip-create-test"}
	eng := newFakeEngine(t, fc, fs, fw)

	res, err := eng.Create(context.Background(), Request{
		ProviderName: "gh",
		Options: provider.CreateRepoOptions{
			Owner: "o", Name: "r", Description: "desc", Visibility: provider.VisibilityPrivate,
		},
		README:  "# r\n\ndesc\n",
		License: "MIT text",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.Repo.Name != "r" {
		t.Fatalf("result repo = %+v", res.Repo)
	}
	if fc.opts.Name != "r" || fc.opts.Visibility != provider.VisibilityPrivate {
		t.Fatalf("create options = %+v", fc.opts)
	}
	if fw.readme == "" || fw.license != "MIT text" || !fw.committed {
		t.Fatalf("worktree not populated: %+v", fw)
	}
	if fw.pushed != repo.SSHURL {
		t.Fatalf("pushed %q, want %q", fw.pushed, repo.SSHURL)
	}
	if !fw.cleaned {
		t.Fatal("worktree should be cleaned after success")
	}
	if fs.calls != 1 || fs.provider != "gh" || fs.repo.Name != "r" {
		t.Fatalf("store = %+v", fs)
	}
}

func TestCreateWithoutContentSkipsWorktree(t *testing.T) {
	fc := &fakeCreator{repo: provider.Repo{Owner: "o", Name: "r"}}
	fs := &fakeStore{}
	eng := NewEngine(fc, fs, nil)
	eng.newWorktree = func(context.Context) (WorktreeHandle, error) {
		t.Fatal("worktree should not be created without content")
		return nil, nil
	}

	res, err := eng.Create(context.Background(), Request{
		ProviderName: "gh",
		Options:      provider.CreateRepoOptions{Owner: "o", Name: "r"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.Repo.Name != "r" || fs.calls != 1 {
		t.Fatalf("result = %+v, store calls = %d", res, fs.calls)
	}
}

func TestCreateRemoteFailureCleansWorktree(t *testing.T) {
	fc := &fakeCreator{err: errors.New("forbidden")}
	fs := &fakeStore{}
	fw := &fakeWorktree{dir: "/tmp/hip-create-test"}
	eng := newFakeEngine(t, fc, fs, fw)

	_, err := eng.Create(context.Background(), Request{
		ProviderName: "gh",
		Options:      provider.CreateRepoOptions{Owner: "o", Name: "r"},
		README:       "# r\n",
	})
	if err == nil || !strings.Contains(err.Error(), "creating repository") {
		t.Fatalf("error = %v", err)
	}
	if !fw.cleaned {
		t.Fatal("worktree should be cleaned when remote creation fails")
	}
	if fs.calls != 0 {
		t.Fatal("cache must not change when remote creation fails")
	}
}

func TestCreatePushFailureKeepsWorktree(t *testing.T) {
	fc := &fakeCreator{repo: provider.Repo{Owner: "o", Name: "r", SSHURL: "git@example.com:o/r.git"}}
	fs := &fakeStore{}
	fw := &fakeWorktree{dir: "/tmp/hip-create-keep", pushErr: errors.New("denied")}
	eng := newFakeEngine(t, fc, fs, fw)

	_, err := eng.Create(context.Background(), Request{
		ProviderName: "gh",
		Options:      provider.CreateRepoOptions{Owner: "o", Name: "r"},
		README:       "# r\n",
	})
	if err == nil {
		t.Fatal("expected push error")
	}
	var pushErr *PushError
	if !errors.As(err, &pushErr) {
		t.Fatalf("error type = %T, want *PushError", err)
	}
	if pushErr.Worktree != fw.dir {
		t.Fatalf("worktree = %q, want %q", pushErr.Worktree, fw.dir)
	}
	if fw.cleaned {
		t.Fatal("worktree must be kept after push failure")
	}
	if fs.calls != 0 {
		t.Fatal("cache must not change when push fails")
	}
}

func TestCreateCacheFailureReportsRepo(t *testing.T) {
	fc := &fakeCreator{repo: provider.Repo{Owner: "o", Name: "r", SSHURL: "git@example.com:o/r.git"}}
	fs := &fakeStore{err: errors.New("disk full")}
	fw := &fakeWorktree{dir: "/tmp/hip-create-test"}
	eng := newFakeEngine(t, fc, fs, fw)

	res, err := eng.Create(context.Background(), Request{
		ProviderName: "gh",
		Options:      provider.CreateRepoOptions{Owner: "o", Name: "r"},
		README:       "# r\n",
	})
	if err == nil || !strings.Contains(err.Error(), "adding repository to cache") {
		t.Fatalf("error = %v", err)
	}
	if res.Repo.Name != "r" {
		t.Fatalf("result repo = %+v", res.Repo)
	}
	if !fw.cleaned {
		t.Fatal("worktree should be cleaned after a successful push")
	}
}

func TestCreateWorktreeFailureStopsBeforeRemote(t *testing.T) {
	fc := &fakeCreator{}
	fs := &fakeStore{}
	eng := NewEngine(fc, fs, nil)
	eng.newWorktree = func(context.Context) (WorktreeHandle, error) {
		return nil, errors.New("no temp space")
	}

	_, err := eng.Create(context.Background(), Request{
		ProviderName: "gh",
		Options:      provider.CreateRepoOptions{Owner: "o", Name: "r"},
		README:       "# r\n",
	})
	if err == nil || !strings.Contains(err.Error(), "preparing worktree") {
		t.Fatalf("error = %v", err)
	}
	if fc.calls != 0 || fs.calls != 0 {
		t.Fatal("remote and cache must not be touched when worktree preparation fails")
	}
}

func TestCreateCommitFailureCleansWorktree(t *testing.T) {
	fc := &fakeCreator{}
	fs := &fakeStore{}
	fw := &fakeWorktree{dir: "/tmp/hip-create-test", commitErr: errors.New("no identity")}
	eng := newFakeEngine(t, fc, fs, fw)

	_, err := eng.Create(context.Background(), Request{
		ProviderName: "gh",
		Options:      provider.CreateRepoOptions{Owner: "o", Name: "r"},
		README:       "# r\n",
	})
	if err == nil {
		t.Fatal("expected commit error")
	}
	if !fw.cleaned {
		t.Fatal("worktree should be cleaned after commit failure")
	}
	if fc.calls != 0 || fs.calls != 0 {
		t.Fatal("remote and cache must not be touched after commit failure")
	}
}
