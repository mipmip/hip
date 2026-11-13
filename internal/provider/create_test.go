package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mipmip/hip/internal/config"
)

func TestNormalizeCreateOptions(t *testing.T) {
	got, err := normalizeCreateOptions(CreateRepoOptions{Owner: " acme ", Name: " repo "})
	if err != nil {
		t.Fatal(err)
	}
	if got.Owner != "acme" || got.Name != "repo" || got.Visibility != VisibilityPrivate {
		t.Fatalf("normalized options = %+v", got)
	}
	if _, err := normalizeCreateOptions(CreateRepoOptions{Name: "repo"}); err == nil {
		t.Fatal("missing owner should fail")
	}
	if _, err := normalizeCreateOptions(CreateRepoOptions{Owner: "acme"}); err == nil {
		t.Fatal("missing name should fail")
	}
	if _, err := normalizeCreateOptions(CreateRepoOptions{Owner: "acme", Name: "repo", Visibility: "internal"}); err == nil {
		t.Fatal("internal visibility should fail")
	}
}

func TestGitHubCreateRepoPersonal(t *testing.T) {
	t.Setenv("HIP_GITHUB_TOKEN", "tok")
	var got ghCreateRepo
	mux := http.NewServeMux()
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("auth = %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"name":"newrepo","description":"desc","owner":{"login":"me"},
			"ssh_url":"git@github.com:me/newrepo.git","clone_url":"https://github.com/me/newrepo.git",
			"html_url":"https://github.com/me/newrepo","default_branch":""}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Username: "me", Type: config.ProviderGitHub, APIURL: srv.URL, Auth: config.Auth{Env: "HIP_GITHUB_TOKEN"}}
	repo, err := NewGitHub(cfg, srv.Client()).CreateRepo(context.Background(), CreateRepoOptions{
		Owner: "me", Name: "newrepo", Description: "desc", Visibility: VisibilityPublic,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "newrepo" || got.Description != "desc" || got.Private || got.AutoInit {
		t.Fatalf("request body = %+v", got)
	}
	if repo.Owner != "me" || repo.Name != "newrepo" || repo.Visibility != VisibilityPublic || repo.DefaultBranch != "main" {
		t.Fatalf("created repo = %+v", repo)
	}
}

func TestGitHubCreateRepoOrganizationAndScopeError(t *testing.T) {
	t.Setenv("HIP_GITHUB_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/acme/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"forbidden"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Username: "me", Type: config.ProviderGitHub, APIURL: srv.URL, Auth: config.Auth{Env: "HIP_GITHUB_TOKEN"}}
	_, err := NewGitHub(cfg, srv.Client()).CreateRepo(context.Background(), CreateRepoOptions{
		Owner: "acme", Name: "newrepo", Visibility: VisibilityPrivate,
	})
	if err == nil {
		t.Fatal("expected forbidden error")
	}
	if !strings.Contains(err.Error(), "`repo` scope") {
		t.Fatalf("error lacks scope hint: %v", err)
	}
}

func TestGitLabCreateRepoPersonal(t *testing.T) {
	t.Setenv("HIP_GITLAB_TOKEN", "tok")
	var got glCreateProject
	mux := http.NewServeMux()
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"name":"newrepo","path":"newrepo","description":"desc",
			"ssh_url_to_repo":"git@gitlab.com:me/newrepo.git","http_url_to_repo":"https://gitlab.com/me/newrepo.git",
			"web_url":"https://gitlab.com/me/newrepo","default_branch":"","visibility":"private",
			"namespace":{"full_path":"me","path":"me"}}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Username: "me", Type: config.ProviderGitLab, APIURL: srv.URL, Auth: config.Auth{Env: "HIP_GITLAB_TOKEN"}}
	repo, err := NewGitLab(cfg, srv.Client()).CreateRepo(context.Background(), CreateRepoOptions{
		Owner: "me", Name: "newrepo", Description: "desc", Visibility: VisibilityPrivate,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "newrepo" || got.Path != "newrepo" || got.Visibility != VisibilityPrivate || got.NamespaceID != 0 || got.InitializeWithReadme {
		t.Fatalf("request body = %+v", got)
	}
	if repo.Owner != "me" || repo.Name != "newrepo" || repo.Visibility != VisibilityPrivate || repo.DefaultBranch != "main" {
		t.Fatalf("created repo = %+v", repo)
	}
}

func TestGitLabCreateRepoGroupResolvesNamespace(t *testing.T) {
	t.Setenv("HIP_GITLAB_TOKEN", "tok")
	var got glCreateProject
	mux := http.NewServeMux()
	mux.HandleFunc("/namespaces", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") != "acme" {
			t.Errorf("search = %q", r.URL.Query().Get("search"))
		}
		w.Write([]byte(`[{"id":7,"full_path":"acme","kind":"group"},{"id":8,"full_path":"acme-other","kind":"group"}]`))
	})
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"name":"newrepo","path":"newrepo","visibility":"public","namespace":{"full_path":"acme"}}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Username: "me", Type: config.ProviderGitLab, APIURL: srv.URL, Auth: config.Auth{Env: "HIP_GITLAB_TOKEN"}}
	repo, err := NewGitLab(cfg, srv.Client()).CreateRepo(context.Background(), CreateRepoOptions{
		Owner: "acme", Name: "newrepo", Visibility: VisibilityPublic,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.NamespaceID != 7 {
		t.Fatalf("namespace id = %d, want 7", got.NamespaceID)
	}
	if repo.Owner != "acme" || repo.Visibility != VisibilityPublic {
		t.Fatalf("created repo = %+v", repo)
	}
}

func TestGitLabCreateRepoMissingNamespace(t *testing.T) {
	t.Setenv("HIP_GITLAB_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/namespaces", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":8,"full_path":"other"}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Username: "me", Type: config.ProviderGitLab, APIURL: srv.URL, Auth: config.Auth{Env: "HIP_GITLAB_TOKEN"}}
	_, err := NewGitLab(cfg, srv.Client()).CreateRepo(context.Background(), CreateRepoOptions{
		Owner: "acme", Name: "newrepo", Visibility: VisibilityPrivate,
	})
	if err == nil || !strings.Contains(err.Error(), "namespace") {
		t.Fatalf("error = %v, want namespace error", err)
	}
}

func TestCodebergCreateRepoPersonal(t *testing.T) {
	t.Setenv("HIP_CODEBERG_TOKEN", "tok")
	var got giteaCreateRepo
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "token tok" {
			t.Errorf("auth = %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"name":"newrepo","description":"desc","owner":{"login":"me"},
			"ssh_url":"git@codeberg.org:me/newrepo.git","clone_url":"https://codeberg.org/me/newrepo.git",
			"html_url":"https://codeberg.org/me/newrepo","default_branch":"","private":false}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Username: "me", Type: config.ProviderCodeberg, APIURL: srv.URL, Auth: config.Auth{Env: "HIP_CODEBERG_TOKEN"}}
	repo, err := NewCodeberg(cfg, srv.Client()).CreateRepo(context.Background(), CreateRepoOptions{
		Owner: "me", Name: "newrepo", Description: "desc", Visibility: VisibilityPublic,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "newrepo" || got.Private || got.AutoInit {
		t.Fatalf("request body = %+v", got)
	}
	if repo.Owner != "me" || repo.Visibility != VisibilityPublic || repo.DefaultBranch != "main" {
		t.Fatalf("created repo = %+v", repo)
	}
}

func TestCodebergCreateRepoOrganization(t *testing.T) {
	t.Setenv("HIP_CODEBERG_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/orgs/acme/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"name":"newrepo","owner":{"login":"acme"},"private":true}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Username: "me", Type: config.ProviderCodeberg, APIURL: srv.URL, Auth: config.Auth{Env: "HIP_CODEBERG_TOKEN"}}
	repo, err := NewCodeberg(cfg, srv.Client()).CreateRepo(context.Background(), CreateRepoOptions{
		Owner: "acme", Name: "newrepo", Visibility: VisibilityPrivate,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.Owner != "acme" || repo.Visibility != VisibilityPrivate {
		t.Fatalf("created repo = %+v", repo)
	}
}

func TestCreateRepoScopeHints(t *testing.T) {
	t.Setenv("HIP_CODEBERG_TOKEN", "tok")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"forbidden"}`))
	}))
	defer srv.Close()

	cfg := config.Provider{Username: "me", Type: config.ProviderCodeberg, APIURL: srv.URL, Auth: config.Auth{Env: "HIP_CODEBERG_TOKEN"}}
	_, err := NewCodeberg(cfg, srv.Client()).CreateRepo(context.Background(), CreateRepoOptions{
		Owner: "me", Name: "newrepo", Visibility: VisibilityPrivate,
	})
	if err == nil || !strings.Contains(err.Error(), "write:repository") {
		t.Fatalf("error = %v, want Codeberg scope hint", err)
	}
}
