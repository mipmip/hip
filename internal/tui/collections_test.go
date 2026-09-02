package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mipmip/huphop/internal/collections"
	"github.com/mipmip/huphop/internal/config"
)

// attachStore gives a model an empty collection store in a temp dir and returns
// the file path, so tests assert against what was actually persisted.
func attachStore(t *testing.T, m *Model, names ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "collections.yaml")
	m.collPath = path
	m.coll = &collections.Store{Version: collections.Version}
	for _, n := range names {
		if _, err := m.coll.Create(n); err != nil {
			t.Fatalf("Create %s: %v", n, err)
		}
	}
	return path
}

// loadStore reads the persisted store, failing the test if it cannot be read.
func loadStore(t *testing.T, path string) *collections.Store {
	t.Helper()
	st, err := collections.Load(path)
	if err != nil {
		t.Fatalf("Load %s: %v", path, err)
	}
	return st
}

func memberNames(t *testing.T, path, name string) []string {
	t.Helper()
	c, ok := loadStore(t, path).Get(name)
	if !ok {
		t.Fatalf("collection %q missing from %s", name, path)
	}
	var out []string
	for _, m := range c.Members {
		out = append(out, m.Key())
	}
	return out
}

func TestAddSelectionToExistingCollection(t *testing.T) {
	m := drillToMipmipRepos(t)
	path := attachStore(t, m, "terraform-modules", "dotfiles")

	send(m, key("space")) // select huphop
	send(m, key("down"))
	send(m, key("space")) // select dotfiles
	if len(m.selectedItems()) != 2 {
		t.Fatalf("precondition: expected 2 selections, got %d", len(m.selectedItems()))
	}

	send(m, key("A"))
	if m.picker == nil {
		t.Fatal("A should open the add-to-collection picker")
	}
	if len(m.picker.items) != 2 {
		t.Fatalf("picker items = %d, want 2", len(m.picker.items))
	}
	if strings.Join(m.picker.names, ",") != "terraform-modules,dotfiles" {
		t.Fatalf("picker names = %v, want the store order", m.picker.names)
	}

	send(m, key("enter")) // commit to the highlighted collection

	if m.picker != nil {
		t.Fatal("committing should close the picker")
	}
	if got := memberNames(t, path, "terraform-modules"); len(got) != 2 {
		t.Fatalf("persisted members = %v, want 2", got)
	}
	if len(m.selectedItems()) != 0 {
		t.Fatalf("commit should clear the selection, got %d", len(m.selectedItems()))
	}
	if !strings.Contains(m.status, "added 2 to terraform-modules") ||
		!strings.Contains(m.status, "0 already members") {
		t.Fatalf("status should report the counts, got %q", m.status)
	}
}

func TestAddReportsAlreadyMembers(t *testing.T) {
	m := drillToMipmipRepos(t)
	path := attachStore(t, m, "tf")

	send(m, key("space")) // huphop
	send(m, key("A"))
	send(m, key("enter"))

	// Add the same repository again plus a second one.
	send(m, key("space")) // huphop again
	send(m, key("down"))
	send(m, key("space")) // dotfiles
	send(m, key("A"))
	send(m, key("enter"))

	if !strings.Contains(m.status, "added 1 to tf") || !strings.Contains(m.status, "1 already members") {
		t.Fatalf("status = %q, want 1 added and 1 already a member", m.status)
	}
	if got := memberNames(t, path, "tf"); len(got) != 2 {
		t.Fatalf("persisted members = %v, want 2 (no duplicate)", got)
	}
}

func TestAddFallsBackToHighlightedRepo(t *testing.T) {
	m := drillToMipmipRepos(t)
	path := attachStore(t, m, "tf")

	send(m, key("A")) // nothing selected
	if m.picker == nil || len(m.picker.items) != 1 {
		t.Fatalf("picker should offer the highlighted repo, got %v", m.picker)
	}
	send(m, key("enter"))
	if got := memberNames(t, path, "tf"); len(got) != 1 || !strings.Contains(got[0], "huphop") {
		t.Fatalf("persisted members = %v, want just huphop", got)
	}
}

func TestAddPickerCancelChangesNothing(t *testing.T) {
	m := drillToMipmipRepos(t)
	path := attachStore(t, m, "tf")

	send(m, key("space"))
	send(m, key("A"))
	send(m, key("esc"))

	if m.picker != nil {
		t.Fatal("esc should close the picker")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("cancelling must not write the store, stat err = %v", err)
	}
	if len(m.selectedItems()) != 1 {
		t.Fatalf("cancelling must keep the selection, got %d", len(m.selectedItems()))
	}
}

func TestCreateCollectionFromPicker(t *testing.T) {
	m := drillToMipmipRepos(t)
	path := attachStore(t, m, "existing")

	send(m, key("space"))
	send(m, key("A"))
	send(m, key("down")) // move onto the "+ new collection" row
	if !m.picker.atNewRow() {
		t.Fatalf("cursor = %d, expected the new-collection row", m.picker.cursor)
	}
	send(m, key("enter")) // start naming
	if !m.picker.naming {
		t.Fatal("expected the name input to open")
	}
	send(m, key("tf-new"))
	send(m, key("enter"))

	if m.picker != nil {
		t.Fatal("creating should close the picker")
	}
	st := loadStore(t, path)
	if _, ok := st.Get("tf-new"); !ok {
		t.Fatalf("tf-new missing, store has %v", st.Names())
	}
	if got := memberNames(t, path, "tf-new"); len(got) != 1 {
		t.Fatalf("members = %v, want 1", got)
	}
}

func TestCreateCollectionNameCollisionKeepsModalOpen(t *testing.T) {
	m := drillToMipmipRepos(t)
	path := attachStore(t, m, "taken")

	send(m, key("space"))
	send(m, key("A"))
	send(m, key("down"))
	send(m, key("enter"))
	send(m, key("taken"))
	send(m, key("enter"))

	if m.picker == nil {
		t.Fatal("a collision should keep the modal open")
	}
	if m.picker.err == "" || !strings.Contains(m.picker.err, "already exists") {
		t.Fatalf("expected a collision error, got %q", m.picker.err)
	}
	if names := m.coll.Names(); len(names) != 1 {
		t.Fatalf("store should be unchanged, got %v", names)
	}
	if c, _ := m.coll.Get("taken"); len(c.Members) != 0 {
		t.Fatalf("a refused create must not add members, got %v", c.Members)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("a refused create must not write the store, stat err = %v", err)
	}

	// Esc steps back to the collection list without closing the modal.
	send(m, key("esc"))
	if m.picker == nil || m.picker.naming {
		t.Fatalf("esc from naming should return to the list, picker = %v", m.picker)
	}
}

func TestAddToCollectionInertInMultiplexAndWithoutStore(t *testing.T) {
	mm := modeModel(t, `tmux -c {{.Target}}`, true, func([]string) error { return nil }, &recordCloner{})
	attachStore(t, mm, "tf")
	send(mm, key("A"))
	if mm.picker != nil {
		t.Fatal("multiplex mode should not open the picker")
	}

	m := drillToMipmipRepos(t)
	m.coll = nil
	m.collErr = os.ErrPermission
	send(m, key("A"))
	if m.picker != nil {
		t.Fatal("an unreadable store should not open the picker")
	}
	if !strings.Contains(m.status, "collections unavailable") {
		t.Fatalf("status should explain the unavailable store, got %q", m.status)
	}
}

// collNavModel returns a model at the providers level with one collection whose
// members span both configured providers.
func collNavModel(t *testing.T) (*Model, string) {
	t.Helper()
	m, _ := newTestModel(t)
	path := attachStore(t, m, "tf")
	if _, _, err := m.coll.Add("tf", []collections.Member{
		{Provider: "github", Owner: "mipmip", Repo: "huphop"},
		{Provider: "codeberg", Owner: "pim", Repo: "notes"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := collections.Save(path, m.coll); err != nil {
		t.Fatal(err)
	}
	return m, path
}

func TestNavigateIntoCollectionAndBackOut(t *testing.T) {
	m, _ := collNavModel(t)

	// The synthetic Collections row is last at the providers level.
	rows := m.providerRows()
	m.cursor = len(rows) - 1
	if rows[m.cursor] != collectionsLabel {
		t.Fatalf("expected the collections row last, got %v", rows)
	}
	send(m, key("enter"))
	if m.nav != levelCollections {
		t.Fatalf("nav = %d, want levelCollections", m.nav)
	}
	if got := m.visibleCollections(); len(got) != 1 || got[0].Name != "tf" || got[0].Count != 2 {
		t.Fatalf("collections level = %v, want tf with 2 members", got)
	}

	send(m, key("enter"))
	if m.nav != levelRepos || m.selColl != "tf" {
		t.Fatalf("nav = %d selColl = %q, want the collection's repo list", m.nav, m.selColl)
	}
	repos := m.visibleRepos()
	if len(repos) != 2 {
		t.Fatalf("collection list = %d repos, want 2", len(repos))
	}
	if repos[0].Provider.Name != "github" || repos[1].Provider.Name != "codeberg" {
		t.Fatalf("collection should span providers in stored order, got %v %v",
			repos[0].Provider.Name, repos[1].Provider.Name)
	}

	send(m, key("esc"))
	if m.nav != levelCollections || m.selColl != "" {
		t.Fatalf("first back should land on the collections level, nav = %d selColl = %q", m.nav, m.selColl)
	}
	send(m, key("esc"))
	if m.nav != levelProviders {
		t.Fatalf("second back should land on the provider level, nav = %d", m.nav)
	}
}

func TestCollectionViewRendersAndFilters(t *testing.T) {
	m, _ := collNavModel(t)
	m.width, m.height = 80, 20
	m.selColl = "tf"
	m.nav = levelRepos

	view := m.View()
	if !strings.Contains(view, "collections > tf") {
		t.Fatalf("breadcrumb should name the collection:\n%s", view)
	}
	// Cross-provider rows carry the provider short, as in the combined view.
	if !strings.Contains(view, "gh  mipmip/huphop") || !strings.Contains(view, "cb  pim/notes") {
		t.Fatalf("rows should be provider-disambiguated:\n%s", view)
	}

	// The fuzzy filter narrows within the collection.
	send(m, key("/"))
	send(m, key("notes"))
	if repos := m.visibleRepos(); len(repos) != 1 || repos[0].Repo.Name != "notes" {
		t.Fatalf("filter inside a collection = %v, want just notes", repos)
	}
}

func TestUnresolvedMembersAreKeptAndInert(t *testing.T) {
	m, path := collNavModel(t)
	// One member whose repo is not cached, one whose provider is not configured.
	if _, _, err := m.coll.Add("tf", []collections.Member{
		{Provider: "github", Owner: "ghost", Repo: "missing"},
		{Provider: "retired-forge", Owner: "someone", Repo: "old"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := collections.Save(path, m.coll); err != nil {
		t.Fatal(err)
	}
	m.width, m.height = 80, 24
	m.selColl = "tf"
	m.nav = levelRepos

	repos := m.visibleRepos()
	if len(repos) != 4 {
		t.Fatalf("expected all 4 members listed, got %d", len(repos))
	}
	if !repos[2].Unresolved || !repos[3].Unresolved {
		t.Fatalf("uncached members should be marked unresolved: %v", repos[2:])
	}
	if repos[2].Provider == nil || repos[2].Provider.Name != "github" {
		t.Fatalf("a configured provider should still resolve: %v", repos[2].Provider)
	}
	if repos[3].Provider != nil {
		t.Fatalf("an unconfigured provider should be nil, got %v", repos[3].Provider)
	}
	if view := m.View(); !strings.Contains(view, "(unresolved)") {
		t.Fatalf("view should mark unresolved members:\n%s", view)
	}

	// Acting on an unresolved row explains itself instead of crashing.
	m.cursor = 3
	send(m, key("space"))
	send(m, key("tab"))
	if len(m.selectedItems()) != 0 {
		t.Fatal("an unresolved row must not be selectable")
	}
	send(m, key("c"))
	if !strings.Contains(m.status, "provider not configured") {
		t.Fatalf("clone on an unconfigured member: status = %q", m.status)
	}
	m.cursor = 2
	send(m, key("enter")) // would open the detail view
	if m.nav != levelRepos || !strings.Contains(m.status, "not in the cache") {
		t.Fatalf("detail on an uncached member: nav = %d status = %q", m.nav, m.status)
	}
	send(m, key("o"))
	if !strings.Contains(m.status, "not in the cache") {
		t.Fatalf("open on an uncached member: status = %q", m.status)
	}

	// Saving the store keeps every member, resolved or not.
	if got := memberNames(t, path, "tf"); len(got) != 4 {
		t.Fatalf("persisted members = %v, want all 4 kept", got)
	}
}

func TestRemoveMembersFromCollection(t *testing.T) {
	m, path := collNavModel(t)
	m.selColl = "tf"
	m.nav = levelRepos
	m.cursor = 0

	send(m, key("d")) // remove the highlighted member
	if got := memberNames(t, path, "tf"); len(got) != 1 || !strings.Contains(got[0], "notes") {
		t.Fatalf("after removing the highlighted member: %v", got)
	}
	if !strings.Contains(m.status, "removed 1 repo from tf") {
		t.Fatalf("status = %q", m.status)
	}

	// A multi-selection wins over the highlight.
	send(m, key("space"))
	send(m, key("d"))
	if got := memberNames(t, path, "tf"); len(got) != 0 {
		t.Fatalf("selection removal left %v", got)
	}
	if len(m.selectedItems()) != 0 {
		t.Fatal("removal should clear the selection")
	}
	if _, ok := loadStore(t, path).Get("tf"); !ok {
		t.Fatal("an emptied collection must still exist")
	}
}

func TestRemoveOutsideACollectionIsInert(t *testing.T) {
	m := drillToMipmipRepos(t)
	attachStore(t, m, "tf")
	before := m.status
	send(m, key("d"))
	if m.status != before {
		t.Fatalf("d outside a collection should do nothing, status = %q", m.status)
	}
}

func TestRenameCollection(t *testing.T) {
	m, path := collNavModel(t)
	m.nav = levelCollections
	m.cursor = 0

	send(m, key("R"))
	if m.prompt == nil || m.prompt.kind != promptRename {
		t.Fatalf("R should open the rename prompt, got %v", m.prompt)
	}
	// The prompt starts from the current name; replace it.
	m.prompt.input.SetValue("terraform")
	send(m, key("enter"))

	if m.prompt != nil {
		t.Fatal("a successful rename should close the prompt")
	}
	st := loadStore(t, path)
	if _, ok := st.Get("terraform"); !ok {
		t.Fatalf("rename not persisted, store has %v", st.Names())
	}
	if got := memberNames(t, path, "terraform"); len(got) != 2 {
		t.Fatalf("rename should keep members, got %v", got)
	}
	if !strings.Contains(m.status, "renamed tf to terraform") {
		t.Fatalf("status = %q", m.status)
	}
}

func TestRenameCollisionKeepsPromptOpen(t *testing.T) {
	m, path := collNavModel(t)
	if _, err := m.coll.Create("other"); err != nil {
		t.Fatal(err)
	}
	if err := collections.Save(path, m.coll); err != nil {
		t.Fatal(err)
	}
	m.nav = levelCollections
	m.cursor = 0

	send(m, key("R"))
	m.prompt.input.SetValue("other")
	send(m, key("enter"))

	if m.prompt == nil {
		t.Fatal("a collision should keep the prompt open")
	}
	if !strings.Contains(m.prompt.err, "already exists") {
		t.Fatalf("expected a collision error, got %q", m.prompt.err)
	}
	if names := strings.Join(loadStore(t, path).Names(), ","); names != "tf,other" {
		t.Fatalf("store should be unchanged, got %q", names)
	}

	send(m, key("esc"))
	if m.prompt != nil {
		t.Fatal("esc should close the rename prompt")
	}
}

func TestDeleteCollectionNeedsConfirmation(t *testing.T) {
	m, path := collNavModel(t)
	m.nav = levelCollections
	m.cursor = 0

	send(m, key("d"))
	if m.prompt == nil || m.prompt.kind != promptDelete {
		t.Fatalf("d should open the delete confirmation, got %v", m.prompt)
	}
	send(m, key("n")) // cancel
	if m.prompt != nil {
		t.Fatal("n should close the confirmation")
	}
	if _, ok := loadStore(t, path).Get("tf"); !ok {
		t.Fatal("cancelling must not delete the collection")
	}

	send(m, key("d"))
	send(m, key("y")) // confirm
	if m.prompt != nil {
		t.Fatal("y should close the confirmation")
	}
	st := loadStore(t, path)
	if _, ok := st.Get("tf"); ok {
		t.Fatalf("collection should be gone, store has %v", st.Names())
	}
	if !strings.Contains(m.status, "deleted collection tf") {
		t.Fatalf("status = %q", m.status)
	}
}

func TestDeletingTheBrowsedCollectionLeavesItsScope(t *testing.T) {
	m, _ := collNavModel(t)
	m.selColl = "tf"
	m.nav = levelCollections
	m.cursor = 0
	send(m, key("d"))
	send(m, key("y"))
	if m.selColl != "" {
		t.Fatalf("selColl = %q, want cleared after deleting that collection", m.selColl)
	}
}

func TestPromptViewsRender(t *testing.T) {
	m, _ := collNavModel(t)
	m.width, m.height = 80, 24
	m.nav = levelCollections

	send(m, key("R"))
	if v := m.View(); !strings.Contains(v, "Rename collection tf") {
		t.Fatalf("rename modal not rendered:\n%s", v)
	}
	send(m, key("esc"))
	send(m, key("d"))
	if v := m.View(); !strings.Contains(v, "Delete collection tf") || !strings.Contains(v, "no clone is touched") {
		t.Fatalf("delete modal not rendered:\n%s", v)
	}
	send(m, key("esc"))

	m.nav = levelRepos
	m.selColl = "tf"
	send(m, key("A"))
	if v := m.View(); !strings.Contains(v, "Add 1 repo to collection") || !strings.Contains(v, newCollectionLabel) {
		t.Fatalf("picker modal not rendered:\n%s", v)
	}
}

func TestCollectionsLevelChromeAndEmptyState(t *testing.T) {
	m, _ := newTestModel(t)
	attachStore(t, m)
	m.width, m.height = 80, 20
	m.nav = levelCollections

	view := m.View()
	if !strings.Contains(view, "collections") {
		t.Fatalf("breadcrumb missing:\n%s", view)
	}
	if !strings.Contains(view, "R: rename") || !strings.Contains(view, "d: delete") {
		t.Fatalf("footer should carry the collection keys:\n%s", view)
	}
	if !strings.Contains(view, "(nothing here)") {
		t.Fatalf("an empty collections level should say so:\n%s", view)
	}

	// An unreadable store says so instead of pretending to be empty.
	m.coll = nil
	m.collErr = os.ErrPermission
	if v := m.View(); !strings.Contains(v, "collections unavailable") {
		t.Fatalf("unavailable store not reported:\n%s", v)
	}
}

func TestCollectionFooterMentionsRemove(t *testing.T) {
	m, _ := collNavModel(t)
	m.nav = levelRepos
	m.selColl = "tf"
	if foot := m.footerText(); !strings.Contains(foot, "d: remove from collection") {
		t.Fatalf("footer inside a collection: %s", foot)
	}
	m.selColl = ""
	if foot := m.footerText(); strings.Contains(foot, "d: remove from collection") {
		t.Fatalf("footer outside a collection should not offer remove: %s", foot)
	}
}

func TestLoadCollectionsFromConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, "huphop"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "huphop", "collections.yaml")
	if err := os.WriteFile(path, []byte("version: 1\ncollections:\n  - name: tf\n    members: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &Model{}
	m.loadCollections()
	if m.collErr != nil {
		t.Fatalf("loadCollections: %v", m.collErr)
	}
	if m.collPath != path {
		t.Fatalf("collPath = %q, want %q", m.collPath, path)
	}
	if names := m.coll.Names(); len(names) != 1 || names[0] != "tf" {
		t.Fatalf("loaded names = %v", names)
	}

	// A malformed file is remembered as an error, never silently replaced.
	if err := os.WriteFile(path, []byte("collections: [oops\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m2 := &Model{}
	m2.loadCollections()
	if m2.coll != nil || m2.collErr == nil {
		t.Fatalf("malformed file should leave the store nil with an error, got %v / %v", m2.coll, m2.collErr)
	}
}

func TestCollectionLaunchOption(t *testing.T) {
	m, _ := collNavModel(t)
	if err := m.applyLaunchOptions("", "tf", false, false); err != nil {
		t.Fatalf("applyLaunchOptions: %v", err)
	}
	if m.nav != levelRepos || m.selColl != "tf" || m.flatAll {
		t.Fatalf("launch should scope to the collection: nav = %d selColl = %q flatAll = %v",
			m.nav, m.selColl, m.flatAll)
	}
	if len(m.visibleRepos()) != 2 {
		t.Fatalf("expected the collection's 2 members, got %d", len(m.visibleRepos()))
	}
}

func TestCollectionLaunchComposesWithFilterAndMode(t *testing.T) {
	m, _ := collNavModel(t)
	m.cfg.Modes = map[string]config.ModeConfig{
		"management": {Header: []string{"breadcrumb"}, Footer: []string{"action_menu"}},
		"multiplex":  {Header: []string{}, Footer: []string{"switch_hint"}, SwitchCommand: "tmux -c {{.Target}}"},
	}
	if err := m.applyLaunchOptions("multiplex", "tf", false, true); err != nil {
		t.Fatalf("applyLaunchOptions: %v", err)
	}
	if m.selColl != "tf" || m.flatAll {
		t.Fatalf("collection should win over the flat view: selColl = %q flatAll = %v", m.selColl, m.flatAll)
	}
	if !m.filtering || !m.filter.Focused() {
		t.Fatal("filter should start active inside the collection")
	}
	if !m.multiplexActive() {
		t.Fatal("multiplex mode should be active")
	}
}

func TestCollectionLaunchUnknownNameErrors(t *testing.T) {
	m, _ := collNavModel(t)
	err := m.applyLaunchOptions("", "nope", false, false)
	if err == nil || !strings.Contains(err.Error(), "unknown collection") {
		t.Fatalf("expected an unknown-collection error, got %v", err)
	}

	m.coll = nil
	m.collErr = os.ErrPermission
	if err := m.applyLaunchOptions("", "tf", false, false); err == nil ||
		!strings.Contains(err.Error(), "collections unavailable") {
		t.Fatalf("expected an unavailable-store error, got %v", err)
	}
}
