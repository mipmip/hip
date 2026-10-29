package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/hip/internal/config"
)

// Run builds the model from cfg and runs the Bubble Tea program until the user
// quits. It is the entrypoint used by the `hip tui` command. When mode is
// non-empty it overrides the configured default_mode; when collection is
// non-empty the TUI starts in that collection's repository list; when flatList
// is true it starts in the combined cross-provider flat view; when filter is
// true it starts with the fuzzy filter already active and focused (filter
// implies the flat view unless a collection scopes the list instead).
func Run(cfg *config.Config, mode, collection string, flatList, filter bool) error {
	m := New(cfg)
	if err := m.applyLaunchOptions(mode, collection, flatList, filter); err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// applyLaunchOptions configures a freshly built model for the requested launch
// state: mode override, collection scope, flat combined view, and filter-active
// start. The filter option implies the flat view (the fuzzy filter only searches
// repositories at the repos level), so it starts flat even when flatList is
// false — unless a collection was named, which scopes the repository list
// instead and takes precedence over the flat view.
func (m *Model) applyLaunchOptions(mode, collection string, flatList, filter bool) error {
	if mode != "" {
		if err := m.SetMode(mode); err != nil {
			return err
		}
	}
	switch {
	case collection != "":
		if err := m.startCollection(collection); err != nil {
			return err
		}
	case flatList || filter:
		m.startFlat()
	}
	if filter {
		m.startFiltering()
	}
	return nil
}
