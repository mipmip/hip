package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mipmip/huphop/internal/collections"
)

// sessionTpl is the session-per-collection shape from the README: the session is
// named after the collection, the window after the repository.
const sessionTpl = `tmux-helper session="{{.Collection}}" window="{{.Repo}}" dir={{.Target}}`

// fallbackTpl is the documented conditional for configurations that want a name
// outside a collection too.
const fallbackTpl = `tmux -s "{{if .Collection}}{{.Collection}}{{else}}{{.Short}}->{{.OwnerLower}}{{end}}"`

func TestSwitchCommandRendersActiveCollection(t *testing.T) {
	m := modeModel(t, sessionTpl, true, func([]string) error { return nil }, &recordCloner{})
	m.selColl = "terraform-modules"
	it := m.reposByProvider["gh"][0]

	got, err := m.renderSwitchCommand(m.activeSwitchCommand(), it)
	if err != nil {
		t.Fatalf("renderSwitchCommand: %v", err)
	}
	want := `tmux-helper session="terraform-modules" window="proj" dir=/tmp/gh.me/proj`
	if got != want {
		t.Fatalf("rendered = %q, want %q", got, want)
	}
}

func TestSwitchCommandCollectionEmptyOutsideACollection(t *testing.T) {
	m := modeModel(t, sessionTpl, true, func([]string) error { return nil }, &recordCloner{})
	it := m.reposByProvider["gh"][0]

	got, err := m.renderSwitchCommand(m.activeSwitchCommand(), it)
	if err != nil {
		t.Fatalf("renderSwitchCommand: %v", err)
	}
	if !strings.Contains(got, `session=""`) {
		t.Fatalf("rendered = %q, want an empty session name outside a collection", got)
	}
}

func TestSwitchCommandConditionalPicksTheFallbackBranch(t *testing.T) {
	m := modeModel(t, fallbackTpl, true, func([]string) error { return nil }, &recordCloner{})
	it := m.reposByProvider["gh"][0]

	// Outside a collection the else-branch names the session after provider+owner.
	got, err := m.renderSwitchCommand(m.activeSwitchCommand(), it)
	if err != nil {
		t.Fatalf("renderSwitchCommand: %v", err)
	}
	if got != `tmux -s "gh->me"` {
		t.Fatalf("fallback branch = %q, want %q", got, `tmux -s "gh->me"`)
	}

	// Inside a collection the if-branch wins.
	m.selColl = "tf"
	got, err = m.renderSwitchCommand(m.activeSwitchCommand(), it)
	if err != nil {
		t.Fatalf("renderSwitchCommand: %v", err)
	}
	if got != `tmux -s "tf"` {
		t.Fatalf("collection branch = %q, want %q", got, `tmux -s "tf"`)
	}
}

func TestSwitchCommandCollectionNameWithSpaceStaysOneArgument(t *testing.T) {
	m := modeModel(t, `tmux new-session -s "{{.Collection}}"`, true, func([]string) error { return nil }, &recordCloner{})
	m.selColl = `client acme; rm -rf ~`
	it := m.reposByProvider["gh"][0]

	rendered, err := m.renderSwitchCommand(m.activeSwitchCommand(), it)
	if err != nil {
		t.Fatalf("renderSwitchCommand: %v", err)
	}
	argv, err := splitArgs(rendered)
	if err != nil {
		t.Fatalf("splitArgs: %v", err)
	}
	want := []string{"tmux", "new-session", "-s", `client acme; rm -rf ~`}
	if !reflect.DeepEqual(argv, want) {
		t.Fatalf("argv = %#v, want %#v", argv, want)
	}
}

func TestMultiplexSwitchInsideCollectionUsesItsName(t *testing.T) {
	var argv []string
	m := modeModel(t, sessionTpl, true, func(a []string) error { argv = a; return nil }, &recordCloner{})

	// Browse the collection that holds the model's one repository.
	path := attachStore(t, m, "tf")
	if _, _, err := m.coll.Add("tf", []collections.Member{
		{Provider: "gh", Owner: "me", Repo: "proj"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := collections.Save(path, m.coll); err != nil {
		t.Fatal(err)
	}
	m.selProvider = nil
	m.selOwner = ""
	m.selColl = "tf"
	m.nav = levelRepos
	if repos := m.visibleRepos(); len(repos) != 1 {
		t.Fatalf("precondition: collection should list 1 repo, got %d", len(repos))
	}

	send(m, key("enter")) // multiplex primary action: already cloned -> switch

	want := []string{"tmux-helper", "session=tf", "window=proj", "dir=/tmp/gh.me/proj"}
	if !reflect.DeepEqual(argv, want) {
		t.Fatalf("argv = %#v, want %#v", argv, want)
	}
	if !m.quitting {
		t.Fatal("a successful switch should quit the TUI")
	}
}
