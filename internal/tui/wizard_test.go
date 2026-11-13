package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mipmip/hip/internal/cache"
	"github.com/mipmip/hip/internal/config"
	"github.com/mipmip/hip/internal/creator"
	"github.com/mipmip/hip/internal/provider"
)

func testWizard(step wizardStep, p *config.Provider, owner, name string) *wizard {
	w := newWizard()
	w.step = step
	w.provider = p
	w.owner = owner
	w.name.SetValue(name)
	return w
}

func TestWizardLaunchAndContext(t *testing.T) {
	m, _ := newTestModel(t)

	m.nav = levelDetail
	m.handleKey(key("n"))
	if m.wizard != nil {
		t.Fatal("wizard opened in detail view")
	}

	m.nav = levelProviders
	m.filtering = true
	m.handleKey(key("n"))
	if m.wizard != nil {
		t.Fatal("wizard opened while filtering")
	}

	m.filtering = false
	m.about = true
	m.handleKey(key("n"))
	if m.wizard != nil {
		t.Fatal("wizard opened while about popup is open")
	}

	m.about = false
	m.cloning = true
	m.handleKey(key("n"))
	if m.wizard != nil {
		t.Fatal("wizard opened while clone modal is open")
	}

	m.cloning = false
	m.handleKey(key("n"))
	if m.wizard == nil || m.wizard.step != wizardProvider {
		t.Fatalf("provider-level launch = %+v, want provider step", m.wizard)
	}
}

func TestWizardPrefillsOwnerAndRepoContext(t *testing.T) {
	m, _ := newTestModel(t)
	p := m.providers[0]

	m.nav = levelOwners
	m.selProvider = p
	m.selOwner = "acme"
	m.handleKey(key("n"))
	if m.wizard == nil || m.wizard.step != wizardName {
		t.Fatalf("owner-level wizard = %+v, want name step", m.wizard)
	}
	if m.wizard.provider.Name != p.Name || m.wizard.owner != "acme" {
		t.Fatalf("prefill = %s/%s, want %s/acme", m.wizard.provider.Name, m.wizard.owner, p.Name)
	}

	m.wizard = nil
	m.nav = levelRepos
	m.selOwner = "mipmip"
	m.cursor = 0
	m.handleKey(key("n"))
	if m.wizard == nil || m.wizard.step != wizardName || m.wizard.owner != "mipmip" {
		t.Fatalf("repo-level wizard = %+v, want name step for mipmip", m.wizard)
	}
}

func TestWizardProviderOwnerNameTransitions(t *testing.T) {
	m, _ := newTestModel(t)
	m.nav = levelProviders
	m.handleKey(key("n"))

	m.handleKey(key("enter"))
	if m.wizard.step != wizardOwner {
		t.Fatalf("step after provider = %v, want owner", m.wizard.step)
	}
	m.handleKey(key("enter"))
	if m.wizard.step != wizardName {
		t.Fatalf("step after owner = %v, want name", m.wizard.step)
	}

	m.handleKey(key("enter"))
	if m.wizard.step != wizardName || m.wizard.err == "" {
		t.Fatalf("empty name accepted: step=%v err=%q", m.wizard.step, m.wizard.err)
	}

	m.wizard.name.SetValue("bad name")
	m.handleKey(key("enter"))
	if !strings.Contains(m.wizard.err, "must not contain") {
		t.Fatalf("invalid name error = %q", m.wizard.err)
	}

	m.wizard.name.SetValue("newrepo")
	m.handleKey(key("enter"))
	if m.wizard.step != wizardVisibility || m.wizard.err != "" {
		t.Fatalf("step after valid name = %v err=%q", m.wizard.step, m.wizard.err)
	}
	if m.wizard.visibility != 0 {
		t.Fatal("visibility should default to private")
	}
	m.handleKey(key(" "))
	if m.wizard.visibility != 1 {
		t.Fatal("space did not toggle visibility")
	}
	m.handleKey(key("enter"))
	if m.wizard.step != wizardReadme {
		t.Fatalf("step after visibility = %v, want readme", m.wizard.step)
	}
}

func TestWizardLicenseAndHolder(t *testing.T) {
	oldGlobal := os.Getenv("GIT_CONFIG_GLOBAL")
	oldSystem := os.Getenv("GIT_CONFIG_SYSTEM")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	defer func() {
		os.Setenv("GIT_CONFIG_GLOBAL", oldGlobal)
		os.Setenv("GIT_CONFIG_SYSTEM", oldSystem)
	}()

	m, _ := newTestModel(t)
	m.wizard = testWizard(wizardCheckout, m.providers[0], "mipmip", "repo")
	m.wizard.license = 1
	m.handleKey(key("enter"))
	if m.wizard.step != wizardHolder {
		t.Fatalf("missing holder step = %v, want holder prompt", m.wizard.step)
	}
	m.handleKey(key("enter"))
	if m.wizard.err == "" {
		t.Fatal("empty holder accepted")
	}
	m.wizard.holder.SetValue("Ada Lovelace")
	m.handleKey(key("enter"))
	if m.wizard.step != wizardReview {
		t.Fatalf("step after holder = %v, want review", m.wizard.step)
	}

	mit := licenseText("MIT", "Ada Lovelace")
	if !strings.Contains(mit, "Ada Lovelace") || !strings.Contains(mit, "MIT License") {
		t.Fatalf("MIT text missing holder or title: %q", mit)
	}
	apache := licenseText("Apache-2.0", "Ada Lovelace")
	if !strings.HasPrefix(apache, "SPDX-License-Identifier: Apache-2.0") || !strings.Contains(apache, "Copyright (c)") {
		t.Fatalf("Apache text missing SPDX/copyright: %q", apache[:min(80, len(apache))])
	}
	gpl := licenseText("GPL-3.0", "Ada Lovelace")
	if !strings.HasPrefix(gpl, "SPDX-License-Identifier: GPL-3.0") {
		t.Fatalf("GPL text missing SPDX: %q", gpl[:min(80, len(gpl))])
	}
	if licenseText("none", "Ada") != "" {
		t.Fatal("none license should be empty")
	}
}

func TestWizardEditorCancelAndSuccess(t *testing.T) {
	m, _ := newTestModel(t)
	m.wizard = testWizard(wizardReadme, m.providers[0], "o", "r")
	m.finishREADME(wizardEditorDoneMsg{Err: errors.New("exit 1")})
	if m.wizard != nil || !strings.Contains(m.status, "create canceled") {
		t.Fatalf("editor cancel wizard=%v status=%q", m.wizard, m.status)
	}

	path := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(path, []byte("# edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m.wizard = testWizard(wizardReadme, m.providers[0], "o", "r")
	m.finishREADME(wizardEditorDoneMsg{Path: path})
	if m.wizard == nil || m.wizard.readmeText != "# edited\n" || m.wizard.step != wizardLicense {
		t.Fatalf("editor success wizard=%+v", m.wizard)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("README temp file still exists: %v", err)
	}
}

func TestWizardCreatedSelectsAndClones(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	m, fc := newTestModel(t)
	p := m.providers[0]
	repo := provider.Repo{
		Owner: "mipmip", Name: "created", SSHURL: "ssh://new", HTTPSURL: "https://new",
		DefaultBranch: "main", Visibility: provider.VisibilityPrivate,
	}
	if err := cache.AddRepo(p.Name, repo); err != nil {
		t.Fatal(err)
	}
	m.wizard = testWizard(wizardReview, p, "mipmip", "created")
	m.wizard.checkout = true

	_, cmd := m.Update(wizardCreatedMsg{Result: creator.Result{Repo: repo}, Checkout: true})
	if m.wizard != nil {
		t.Fatal("wizard still open after success")
	}
	if m.nav != levelRepos || m.selOwner != "mipmip" || m.cursor != 0 {
		t.Fatalf("selection nav=%v owner=%q cursor=%d", m.nav, m.selOwner, m.cursor)
	}
	if cmd == nil {
		t.Fatal("checkout command missing")
	}
	msg := cmd()
	m.Update(msg)
	if len(fc.calls) != 1 || !strings.Contains(fc.calls[0], "mipmip/created") {
		t.Fatalf("clone calls = %v", fc.calls)
	}
}

func TestWizardCreatedErrorStaysOpen(t *testing.T) {
	m, _ := newTestModel(t)
	m.wizard = testWizard(wizardReview, m.providers[0], "o", "r")
	m.wizard.busy = true
	m.handleWizardCreated(wizardCreatedMsg{Err: errors.New("push failed")})
	if m.wizard == nil || m.wizard.busy || !strings.Contains(m.wizard.err, "push failed") {
		t.Fatalf("wizard after error = %+v", m.wizard)
	}
}

func TestWizardViewAndFooter(t *testing.T) {
	m, _ := newTestModel(t)
	if got := m.wizardView(); got != "" {
		t.Fatalf("nil wizard view = %q", got)
	}
	m.wizard = testWizard(wizardReview, m.providers[0], "mipmip", "created")
	m.wizard.visibility = 1
	m.wizard.readme = true
	m.wizard.license = 2
	m.wizard.checkout = true
	view := m.wizardView()
	for _, want := range []string{"New repository", "github", "mipmip", "created", "public", "Apache-2.0", "checkout"} {
		if !strings.Contains(view, want) {
			t.Fatalf("review view missing %q:\n%s", want, view)
		}
	}
	if !strings.Contains(m.footerText(), "n: new repo") {
		t.Fatalf("footer missing new repo hint: %q", m.footerText())
	}
}
