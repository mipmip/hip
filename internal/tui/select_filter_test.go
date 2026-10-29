package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/mipmip/hip/internal/config"
	"github.com/mipmip/hip/internal/provider"
)

// selectedKeys returns the "provider owner/name" keys of every selected item,
// so assertions do not depend on iteration order of the model maps.
func selectedKeys(m *Model) []string {
	var out []string
	for _, it := range m.selectedItems() {
		out = append(out, it.Provider.Name+" "+it.key())
	}
	return out
}

func hasKey(keys []string, want string) bool {
	for _, k := range keys {
		if k == want {
			return true
		}
	}
	return false
}

func TestTabSelectsWhileFilteringWithoutTouchingQuery(t *testing.T) {
	m := drillToMipmipRepos(t)
	m.height = 10
	send(m, key("/"))
	send(m, key("o")) // matches huphop and dotfiles
	if got := m.filter.Value(); got != "o" {
		t.Fatalf("precondition: query = %q, want %q", got, "o")
	}
	before := m.offset

	send(m, key("tab"))

	if !m.filtering {
		t.Fatal("Tab must not leave filter-edit mode")
	}
	if got := m.filter.Value(); got != "o" {
		t.Fatalf("Tab changed the query to %q; it must stay %q", got, "o")
	}
	if m.offset != before {
		t.Fatalf("Tab moved the scroll window: offset = %d, want %d", m.offset, before)
	}
	if m.cursor != 1 {
		t.Fatalf("Tab should advance the cursor: cursor = %d, want 1", m.cursor)
	}
	sel := selectedKeys(m)
	if len(sel) != 1 || !hasKey(sel, "github mipmip/huphop") {
		t.Fatalf("expected huphop selected, got %v", sel)
	}
}

func TestShiftTabTogglesAndRetreats(t *testing.T) {
	m := drillToMipmipRepos(t)
	send(m, key("/"))
	send(m, key("down")) // cursor -> 1 (dotfiles)
	if m.cursor != 1 {
		t.Fatalf("precondition: cursor = %d, want 1", m.cursor)
	}

	send(m, key("shift+tab"))

	if m.cursor != 0 {
		t.Fatalf("shift+Tab should retreat: cursor = %d, want 0", m.cursor)
	}
	if sel := selectedKeys(m); len(sel) != 1 || !hasKey(sel, "github mipmip/dotfiles") {
		t.Fatalf("expected dotfiles selected, got %v", sel)
	}
	if !m.filtering {
		t.Fatal("shift+Tab must not leave filter-edit mode")
	}
}

func TestTabTogglesOffWhenPressedTwice(t *testing.T) {
	m := drillToMipmipRepos(t)
	send(m, key("/"))
	send(m, key("tab")) // select row 0, cursor -> 1
	send(m, key("up"))  // back to row 0
	send(m, key("tab")) // deselect row 0
	if sel := selectedKeys(m); len(sel) != 0 {
		t.Fatalf("expected nothing selected after toggling twice, got %v", sel)
	}
}

func TestSpaceStaysAFilterCharacterWhileFiltering(t *testing.T) {
	m := drillToMipmipRepos(t)
	send(m, key("/"))
	send(m, key("o"))
	send(m, key("space"))

	if got := m.filter.Value(); got != "o " {
		t.Fatalf("space should extend the query: got %q, want %q", got, "o ")
	}
	if sel := selectedKeys(m); len(sel) != 0 {
		t.Fatalf("space must not select while filtering, got %v", sel)
	}
}

func TestSelectionsSurviveRefilteringAcrossProviders(t *testing.T) {
	m, _ := newTestModel(t)
	m.startFlat()

	// Query A: pick the one codeberg repo.
	send(m, key("/"))
	send(m, key("n"))
	send(m, key("o"))
	send(m, key("t")) // "not" matches notes only
	if repos := m.visibleRepos(); len(repos) != 1 || repos[0].Repo.Name != "notes" {
		t.Fatalf("query A should show only notes, got %v", repos)
	}
	send(m, key("tab"))

	// Replace the query entirely and pick from a different provider.
	for range 3 {
		send(m, key("backspace"))
	}
	send(m, key("w")) // widgets
	if repos := m.visibleRepos(); len(repos) != 1 || repos[0].Repo.Name != "widgets" {
		t.Fatalf("query B should show only widgets, got %v", repos)
	}
	send(m, key("tab"))

	sel := selectedKeys(m)
	if len(sel) != 2 {
		t.Fatalf("expected 2 selections across queries and providers, got %v", sel)
	}
	if !hasKey(sel, "codeberg pim/notes") || !hasKey(sel, "github acme/widgets") {
		t.Fatalf("selection lost across re-filtering: %v", sel)
	}
}

func TestSelectionKeysInertInMultiplexAndAboveRepos(t *testing.T) {
	// Multiplex mode: the primary action is single-repo, so nothing selects.
	mm := modeModel(t, `tmux -c {{.Target}}`, true, func([]string) error { return nil }, &recordCloner{})
	send(mm, key("tab"))
	send(mm, key("shift+tab"))
	send(mm, key("space"))
	if sel := selectedKeys(mm); len(sel) != 0 {
		t.Fatalf("multiplex mode must not select, got %v", sel)
	}

	// Provider and owner levels have no repository to select.
	m, _ := newTestModel(t)
	send(m, key("tab"))
	if m.cursor != 0 {
		t.Fatalf("Tab at the provider level moved the cursor to %d", m.cursor)
	}
	send(m, key("enter")) // owners
	send(m, key("tab"))
	if sel := selectedKeys(m); len(sel) != 0 {
		t.Fatalf("owner level must not select, got %v", sel)
	}
	if m.cursor != 0 {
		t.Fatalf("Tab at the owner level moved the cursor to %d", m.cursor)
	}
}

func TestSelectedFacetNarrowsToTheBasket(t *testing.T) {
	m := drillToMipmipRepos(t)
	send(m, key("space")) // select huphop, cursor stays
	if sel := selectedKeys(m); len(sel) != 1 {
		t.Fatalf("precondition: expected 1 selection, got %v", sel)
	}

	send(m, key("x")) // selected:only
	repos := m.visibleRepos()
	if len(repos) != 1 || repos[0].Repo.Name != "huphop" {
		t.Fatalf("selected:only should show just the basket, got %v", repos)
	}
	if !strings.Contains(m.facets.status(), "selected:only") {
		t.Fatalf("facet status should report the basket facet, got %q", m.facets.status())
	}

	send(m, key("x")) // selected:hide
	repos = m.visibleRepos()
	if len(repos) != 1 || repos[0].Repo.Name != "dotfiles" {
		t.Fatalf("selected:hide should exclude the basket, got %v", repos)
	}

	send(m, key("x")) // back to all
	if m.facets.selected != triAll || m.facets.active() {
		t.Fatalf("third press should return to all, got %v", m.facets.selected)
	}
	if len(m.visibleRepos()) != 2 {
		t.Fatalf("expected both repos again, got %d", len(m.visibleRepos()))
	}
}

func TestSelectedFacetComposesWithQueryAndOtherFacets(t *testing.T) {
	m, _ := newTestModel(t)
	// Make one repo a fork so another facet has something to bite on.
	m.reposByProvider["github"][1].Repo.Fork = true
	m.startFlat()

	// Select huphop and dotfiles (rows 0 and 1 of the flat list).
	send(m, key("tab"))
	send(m, key("tab"))
	if sel := selectedKeys(m); len(sel) != 2 {
		t.Fatalf("precondition: expected 2 selections, got %v", sel)
	}

	send(m, key("x")) // selected:only
	send(m, key("f")) // fork:only -> dotfiles is the only selected fork
	repos := m.visibleRepos()
	if len(repos) != 1 || repos[0].Repo.Name != "dotfiles" {
		t.Fatalf("selected:only AND fork:only should leave dotfiles, got %v", repos)
	}
	status := m.facets.status()
	if !strings.Contains(status, "fork:only") || !strings.Contains(status, "selected:only") {
		t.Fatalf("status should show both facets, got %q", status)
	}

	// A fuzzy query narrows further and can empty the result.
	send(m, key("/"))
	send(m, key("h"))
	send(m, key("u"))
	if repos := m.visibleRepos(); len(repos) != 0 {
		t.Fatalf("query 'hu' AND fork:only AND selected:only should match nothing, got %v", repos)
	}
}

func TestUnpickFromTheBasketRemovesTheRow(t *testing.T) {
	m := drillToMipmipRepos(t)
	send(m, key("space")) // select huphop
	send(m, key("down"))
	send(m, key("space")) // select dotfiles
	send(m, key("x"))     // selected:only
	if len(m.visibleRepos()) != 2 {
		t.Fatalf("precondition: expected 2 rows in the basket, got %d", len(m.visibleRepos()))
	}

	m.cursor = 0
	send(m, key("tab")) // un-pick the first row

	repos := m.visibleRepos()
	if len(repos) != 1 || repos[0].Repo.Name != "dotfiles" {
		t.Fatalf("un-picked row should leave the basket, got %v", repos)
	}
	// The following row slid into the cursor position, so Tab must not skip it.
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 (no skip when the row disappears)", m.cursor)
	}
}

func TestIndicatorReportsGlobalSelectionCount(t *testing.T) {
	m, _ := newTestModel(t)
	m.height = 20
	m.startFlat()
	send(m, key("tab")) // select the first flat row
	send(m, key("tab")) // and the second

	_ = m.View()
	if !strings.Contains(m.indicatorText, "2 selected") {
		t.Fatalf("indicator should report the selection, got %q", m.indicatorText)
	}

	// Drill into a single owner: the count stays global even though only some of
	// the selected repositories are in scope.
	m.flatAll = false
	m.nav = levelOwners
	m.selProvider = m.providers[0]
	send(m, key("j"))     // mipmip
	send(m, key("enter")) // repos of mipmip
	_ = m.View()
	if !strings.Contains(m.indicatorText, "2 selected") {
		t.Fatalf("scoped indicator should still report 2 selected, got %q", m.indicatorText)
	}
}

func TestIndicatorOmitsSelectionCountWhenNothingSelected(t *testing.T) {
	m := drillToMipmipRepos(t)
	m.height = 20
	_ = m.View()
	if strings.Contains(m.indicatorText, "selected") {
		t.Fatalf("indicator should stay clean with no selection, got %q", m.indicatorText)
	}
}

// TestFacetMatchSelectedPredicate covers the pure predicate directly, including
// the selection flag that lives on the item rather than the repo.
func TestFacetMatchSelectedPredicate(t *testing.T) {
	r := provider.Repo{Name: "a"}
	f := facets{selected: triOnly}
	if !f.match(r, true) || f.match(r, false) {
		t.Fatal("selected:only predicate wrong")
	}
	f = facets{selected: triHide}
	if f.match(r, true) || !f.match(r, false) {
		t.Fatal("selected:hide predicate wrong")
	}
}

// TestFooterAdvertisesSelectionKeys keeps the help bar honest about the new keys.
func TestFooterAdvertisesSelectionKeys(t *testing.T) {
	m := &Model{cfg: &config.Config{}, filter: textinput.New()}
	foot := m.footerText()
	for _, want := range []string{"space/tab: select", "x: selected", "A: add to collection"} {
		if !strings.Contains(foot, want) {
			t.Fatalf("footer missing %q: %s", want, foot)
		}
	}
	m.filtering = true
	if got := m.footerText(); !strings.Contains(got, "tab: select") {
		t.Fatalf("filtering footer should mention tab: %s", got)
	}
}
