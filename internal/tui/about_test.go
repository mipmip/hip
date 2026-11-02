package tui

import (
	"runtime"
	"strings"
	"testing"
)

func TestAboutPopupOpensFromListAndDetail(t *testing.T) {
	m, _ := newTestModel(t)

	send(m, key("?"))
	if !m.about {
		t.Fatal("? should open the about popup from the list")
	}
	send(m, key("esc"))
	if m.about {
		t.Fatal("esc should close the about popup")
	}

	m.nav = levelDetail
	send(m, key("?"))
	if !m.about {
		t.Fatal("? should open the about popup from the detail view")
	}
	send(m, key("?"))
	if m.about {
		t.Fatal("? should close the about popup")
	}
}

func TestAboutPopupSwallowsOtherKeys(t *testing.T) {
	m, _ := newTestModel(t)
	send(m, key("?"))
	nav, cursor := m.nav, m.cursor

	send(m, key("j"))
	send(m, key("enter"))
	send(m, key("q"))

	if !m.about {
		t.Fatal("about popup should stay open")
	}
	if m.nav != nav || m.cursor != cursor {
		t.Fatalf("keys leaked through: nav=%d cursor=%d", m.nav, m.cursor)
	}
	if m.quitting {
		t.Fatal("q should not quit while the about popup is open")
	}
}

func TestFilterKeepsQuestionMark(t *testing.T) {
	m, _ := newTestModel(t)
	m.startFiltering()

	send(m, key("?"))

	if m.about {
		t.Fatal("? must not open the about popup while filtering")
	}
	if got := m.filter.Value(); got != "?" {
		t.Fatalf("filter value = %q, want ?", got)
	}
}

func TestAboutHintInFooters(t *testing.T) {
	m, _ := newTestModel(t)
	if got := m.footerText(); !strings.Contains(got, "?: about") {
		t.Fatalf("list footer missing about hint: %s", got)
	}
	if got := m.detailFooterText(); !strings.Contains(got, "?: about") {
		t.Fatalf("detail footer missing about hint: %s", got)
	}
	got, ok := m.renderElement("switch_hint")
	if !ok || !strings.Contains(got, "?: about") {
		t.Fatalf("switch hint missing about hint: %q", got)
	}
}

func TestAboutPopupRenders(t *testing.T) {
	m, _ := newTestModel(t)
	m.version = "9.9.9"
	m.about = true
	m.width, m.height = 80, 24

	view := m.View()
	for _, want := range []string{
		"hip 9.9.9",
		"version   9.9.9",
		"build     " + runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH,
		"license   MIT, (c) 2026 Pim Snel",
		"github    github.com/mipmip/hip",
		"credits   Pim Snel, et al",
		"? or esc: close",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("about view missing %q:\n%s", want, view)
		}
	}
}
