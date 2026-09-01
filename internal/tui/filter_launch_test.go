package tui

import (
	"strings"
	"testing"
)

// TestFilterLaunchStartsFlatFiltering verifies --filter boots the model into the
// combined flat repo list with the fuzzy filter active and focused.
func TestFilterLaunchStartsFlatFiltering(t *testing.T) {
	m, _ := newTestModel(t)
	if err := m.applyLaunchOptions("", false, true); err != nil {
		t.Fatalf("applyLaunchOptions: %v", err)
	}
	if !m.flatAll || m.nav != levelRepos {
		t.Fatalf("filter launch should be flat at repos level: flatAll=%v nav=%d", m.flatAll, m.nav)
	}
	if !m.filtering || !m.filter.Focused() {
		t.Fatalf("filter launch should be filtering and focused: filtering=%v focused=%v",
			m.filtering, m.filter.Focused())
	}
}

// TestFilterLaunchImpliesFlat verifies --filter opens the flat view even without
// --flatlist (the fuzzy filter only searches repositories at the repos level).
func TestFilterLaunchImpliesFlat(t *testing.T) {
	m, _ := newTestModel(t)
	if err := m.applyLaunchOptions("", false, true); err != nil {
		t.Fatalf("applyLaunchOptions: %v", err)
	}
	if !m.flatAll || m.nav != levelRepos || m.selProvider != nil {
		t.Fatalf("filter alone must imply the flat view: flatAll=%v nav=%d selProvider=%v",
			m.flatAll, m.nav, m.selProvider)
	}
}

// TestFilterLaunchComposesWithMultiplex verifies --filter composes with --mode
// multiplex: filtering starts active, the switch command is live, and the
// multi-select checkbox is suppressed.
func TestFilterLaunchComposesWithMultiplex(t *testing.T) {
	m := modeModel(t, "echo {{.Repo}}", false, func([]string) error { return nil }, &recordCloner{})
	m.nav = levelProviders // simulate a fresh launch before options are applied
	if err := m.applyLaunchOptions("multiplex", false, true); err != nil {
		t.Fatalf("applyLaunchOptions: %v", err)
	}
	if !m.filtering || !m.flatAll || m.nav != levelRepos {
		t.Fatalf("filter+multiplex launch should be flat filtering: filtering=%v flatAll=%v nav=%d",
			m.filtering, m.flatAll, m.nav)
	}
	if !m.multiplexActive() {
		t.Fatal("multiplex switch command should be active")
	}
	if body := m.bodyText(); strings.Contains(body, "[ ]") || strings.Contains(body, "[x]") {
		t.Fatalf("multiplex body should have no checkbox:\n%s", body)
	}
}

// TestFilterLaunchEscUnwind verifies the two-stage Esc: the first Esc clears the
// query and stops filtering (staying in the flat list), the second backs the
// flat view out to the provider level.
func TestFilterLaunchEscUnwind(t *testing.T) {
	m, _ := newTestModel(t)
	if err := m.applyLaunchOptions("", false, true); err != nil {
		t.Fatalf("applyLaunchOptions: %v", err)
	}
	send(m, key("m")) // type into the filter

	// First Esc: stop filtering, clear the query, stay in the flat list.
	send(m, key("esc"))
	if m.filtering {
		t.Fatal("first esc should stop filtering")
	}
	if m.filter.Value() != "" {
		t.Fatalf("first esc should clear the query, got %q", m.filter.Value())
	}
	if !m.flatAll || m.nav != levelRepos {
		t.Fatalf("first esc should stay in the flat list: flatAll=%v nav=%d", m.flatAll, m.nav)
	}

	// Second Esc: back the flat view out to the provider level.
	send(m, key("esc"))
	if m.flatAll || m.nav != levelProviders {
		t.Fatalf("second esc should go up to providers: flatAll=%v nav=%d", m.flatAll, m.nav)
	}
}

// TestFilterLaunchEnterSwitchesInMultiplex verifies that from a filter-active
// multiplex start, Enter on the highlighted repo triggers the clone-then-switch
// primary action.
func TestFilterLaunchEnterSwitchesInMultiplex(t *testing.T) {
	var ran []string
	m := modeModel(t, "echo {{.Repo}}", true, func(argv []string) error { ran = argv; return nil }, &recordCloner{})
	m.nav = levelProviders
	if err := m.applyLaunchOptions("multiplex", false, true); err != nil {
		t.Fatalf("applyLaunchOptions: %v", err)
	}
	// Enter while filtering leaves edit mode then acts on the highlighted repo.
	send(m, key("enter"))
	if len(ran) == 0 {
		t.Fatal("enter should run the switch command for the highlighted repo")
	}
}
