package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mipmip/hip/internal/config"
	"github.com/mipmip/hip/internal/fetch"
)

// OwnerFetcher is a provider client that can fetch a single owner's
// repositories while emitting progress events. The built-in GitHub, Codeberg
// and GitLab clients all satisfy it; callers that want progress (the TUI and the
// CLI) type-assert a Provider to OwnerFetcher.
type OwnerFetcher interface {
	// FetchOwner fetches one owner's repositories page by page, emitting progress
	// through emit (which may be nil). When owner equals the provider's configured
	// Username it fetches the authenticated user's own repositories; any other
	// owner is fetched as an organization/group. On cancellation it returns
	// ctx.Err() after emitting a Canceled event; on failure it returns the error
	// after a Failed event.
	FetchOwner(ctx context.Context, emit fetch.Emit, owner string) ([]Repo, error)
}

// repoKey is the owner/name dedupe key used when merging pages and owners.
func repoKey(r Repo) string { return r.Owner + "/" + r.Name }

// listReposOverFetch implements the plain, event-free ListRepos contract over an
// owner-at-a-time FetchOwner. An empty owners slice means the authenticated
// user's own repositories (the provider's Username). Results across owners are
// merged and deduped by owner/name; per-owner filtering already happened in
// FetchOwner, so the merged result is returned as-is.
func listReposOverFetch(
	ctx context.Context,
	owners []string,
	p *config.Provider,
	fetchOwner func(ctx context.Context, emit fetch.Emit, owner string) ([]Repo, error),
) ([]Repo, error) {
	req := owners
	if len(req) == 0 {
		req = []string{p.Username}
	}

	var all []Repo
	seen := make(map[string]struct{})
	for _, owner := range req {
		repos, err := fetchOwner(ctx, nil, owner)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			key := repoKey(r)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			all = append(all, r)
		}
	}
	return all, nil
}

// osLookupEnv is the production EnvLookup; it mirrors os.LookupEnv.
func osLookupEnv(key string) (string, bool) { return os.LookupEnv(key) }

// readAllAndClose reads the full response body and closes it, always closing the
// body even when reading fails.
func readAllAndClose(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// normalizeCreateOptions trims and validates create input. An empty visibility
// defaults to private.
func normalizeCreateOptions(opts CreateRepoOptions) (CreateRepoOptions, error) {
	opts.Owner = strings.TrimSpace(opts.Owner)
	opts.Name = strings.TrimSpace(opts.Name)
	opts.Description = strings.TrimSpace(opts.Description)
	if opts.Visibility == "" {
		opts.Visibility = VisibilityPrivate
	}
	if opts.Owner == "" {
		return opts, fmt.Errorf("create repository: owner is required")
	}
	if opts.Name == "" {
		return opts, fmt.Errorf("create repository: name is required")
	}
	if opts.Visibility != VisibilityPublic && opts.Visibility != VisibilityPrivate {
		return opts, fmt.Errorf("create repository: visibility must be public or private")
	}
	return opts, nil
}

// postJSON performs an authenticated JSON POST and returns the body and status.
func postJSON(
	ctx context.Context,
	client *http.Client,
	u string,
	payload any,
	headers map[string]string,
) ([]byte, int, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("encoding request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(encoded))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	body, err := readAllAndClose(resp)
	return body, resp.StatusCode, err
}

// createPermissionHint returns the provider-specific scope hint shown on a 403
// during repository creation.
func createPermissionHint(kind config.ProviderType, visibility string) string {
	switch kind {
	case config.ProviderGitHub:
		if visibility == VisibilityPrivate {
			return "GitHub token needs the `repo` scope to create private repositories"
		}
		return "GitHub token needs the `public_repo` scope to create public repositories"
	case config.ProviderGitLab:
		return "GitLab token needs the `api` scope to create projects"
	case config.ProviderCodeberg:
		return "Codeberg token needs the `write:repository` scope to create repositories"
	default:
		return ""
	}
}

// createHTTPError formats a non-success create response, adding the scope hint
// on a 403.
func createHTTPError(label, endpoint string, status int, body []byte, hint string) error {
	msg := fmt.Sprintf("%s: %s: %s: %s", label, endpoint, http.StatusText(status), strings.TrimSpace(string(body)))
	if status == http.StatusForbidden && hint != "" {
		msg += " (" + hint + ")"
	}
	return fmt.Errorf("%s", msg)
}

// defaultBranchOrMain fills in the conventional main branch when a provider
// returns an empty default branch for a new empty repository.
func defaultBranchOrMain(branch string) string {
	if branch == "" {
		return "main"
	}
	return branch
}

// NewDefaultRegistry returns a Registry with the built-in GitHub, Codeberg and
// GitLab clients registered. Constructors use http.DefaultClient and resolve
// auth via the real CLIs and environment.
func NewDefaultRegistry() *Registry {
	reg := NewRegistry()
	reg.Register(config.ProviderGitHub, func(p *config.Provider) (Provider, error) {
		return NewGitHub(*p, nil), nil
	})
	reg.Register(config.ProviderCodeberg, func(p *config.Provider) (Provider, error) {
		return NewCodeberg(*p, nil), nil
	})
	reg.Register(config.ProviderGitLab, func(p *config.Provider) (Provider, error) {
		return NewGitLab(*p, nil), nil
	})
	return reg
}
