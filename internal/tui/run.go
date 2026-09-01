package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/huphop/internal/config"
)

// Run builds the model from cfg and runs the Bubble Tea program until the user
// quits. It is the entrypoint used by the `hup tui` command. When mode is
// non-empty it overrides the configured default_mode; when flatList is true the
// TUI starts in the combined cross-provider flat view; when filter is true the
// TUI starts in that flat view with the fuzzy filter already active and focused
// (filter implies the flat view, so the first keystroke narrows repositories).
func Run(cfg *config.Config, mode string, flatList, filter bool) error {
	m := New(cfg)
	if err := m.applyLaunchOptions(mode, flatList, filter); err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// applyLaunchOptions configures a freshly built model for the requested launch
// state: mode override, flat combined view, and filter-active start. The filter
// option implies the flat view (the fuzzy filter only searches repositories at
// the repos level), so it starts flat even when flatList is false.
func (m *Model) applyLaunchOptions(mode string, flatList, filter bool) error {
	if mode != "" {
		if err := m.SetMode(mode); err != nil {
			return err
		}
	}
	if flatList || filter {
		m.startFlat()
	}
	if filter {
		m.startFiltering()
	}
	return nil
}
