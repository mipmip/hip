package tui

// collections.go wires the user-defined collection store into the TUI: the
// add-to-collection modal, the collections navigation level and the
// member/collection mutations. Every mutation writes the store immediately, and
// a failed write rolls the in-memory store back to its pre-mutation snapshot so
// what the TUI shows always matches collections.yaml.

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mipmip/hip/internal/collections"
)

// newCollectionRow is the sentinel cursor row in the picker that creates a
// collection instead of choosing an existing one.
const newCollectionLabel = "+ new collection"

// collPicker is the add-to-collection modal state: the repositories being
// added, the choosable collection names, and the name input used by the
// create-a-collection row.
type collPicker struct {
	// items are the repositories to add.
	items []repoItem
	// names are the existing collection names in file order; the cursor may also
	// sit one past the end, on the create-a-collection row.
	names []string
	// counts are the member counts parallel to names.
	counts []int
	// cursor is the highlighted row (len(names) == the create row).
	cursor int
	// naming is true while the user types a new collection's name.
	naming bool
	name   textinput.Model
	// err is the last error to report inside the modal (collision, write failure).
	err string
}

// atNewRow reports whether the cursor sits on the create-a-collection row.
func (p *collPicker) atNewRow() bool { return p.cursor >= len(p.names) }

// memberOf maps a repo item to its store identity: the configured provider name
// plus owner and repository name.
func memberOf(it repoItem) collections.Member {
	return collections.Member{
		Provider: it.Provider.Name,
		Owner:    it.Repo.Owner,
		Repo:     it.Repo.Name,
	}
}

// membersOf maps repo items to store members, skipping items with no provider.
func membersOf(items []repoItem) []collections.Member {
	out := make([]collections.Member, 0, len(items))
	for _, it := range items {
		if it.Provider == nil {
			continue
		}
		out = append(out, memberOf(it))
	}
	return out
}

// collectionsReady reports whether the store is usable, setting a status message
// when it is not (a malformed collections.yaml is reported, never overwritten).
func (m *Model) collectionsReady() bool {
	if m.coll != nil {
		return true
	}
	if m.collErr != nil {
		m.status = fmt.Sprintf("collections unavailable: %v", m.collErr)
	} else {
		m.status = "collections unavailable"
	}
	return false
}

// saveCollections persists the store, rolling back to snap when the write fails
// so the in-memory store never drifts from the file.
func (m *Model) saveCollections(snap *collections.Store) error {
	if err := collections.Save(m.collPath, m.coll); err != nil {
		m.coll = snap
		return err
	}
	return nil
}

// clearSelection drops the multi-selection across all providers.
func (m *Model) clearSelection() {
	for _, p := range m.providers {
		items := m.reposByProvider[p.Name]
		for i := range items {
			items[i].Selected = false
		}
	}
}

// addToCollection opens the add-to-collection modal for the multi-selection,
// falling back to the highlighted repository when nothing is selected. It is
// inert in multiplex mode, where the primary action is single-repo.
func (m *Model) addToCollection() (tea.Model, tea.Cmd) {
	if m.multiplexActive() {
		return m, nil
	}
	if !m.collectionsReady() {
		return m, nil
	}
	items := m.selectedItems()
	if len(items) == 0 {
		it, ok := m.currentRepo()
		if !ok {
			m.status = "nothing to add to a collection"
			return m, nil
		}
		items = []repoItem{it}
	}

	ti := textinput.New()
	ti.Placeholder = "collection name"
	ti.Prompt = "name: "
	ti.Focus()

	p := &collPicker{items: items, name: ti}
	for _, c := range m.coll.Collections {
		p.names = append(p.names, c.Name)
		p.counts = append(p.counts, len(c.Members))
	}
	m.picker = p
	return m, nil
}

// handlePickerKey is the modal key map for the add-to-collection picker: it
// swallows every key so the list underneath cannot move while the modal is open.
func (m *Model) handlePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	p := m.picker
	if p.naming {
		switch msg.Type {
		case tea.KeyEnter:
			return m.commitAddToCollection(strings.TrimSpace(p.name.Value()), true)
		case tea.KeyEsc:
			// Back to the collection list, keeping the modal open.
			p.naming = false
			p.err = ""
			p.name.SetValue("")
			return m, nil
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		}
		var cmd tea.Cmd
		p.name, cmd = p.name.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc", "q":
		m.picker = nil
		return m, nil
	case "up", "k":
		if p.cursor > 0 {
			p.cursor--
		}
		return m, nil
	case "down", "j":
		if p.cursor < len(p.names) {
			p.cursor++
		}
		return m, nil
	case "enter":
		if p.atNewRow() {
			p.naming = true
			p.err = ""
			p.name.Focus()
			return m, nil
		}
		return m.commitAddToCollection(p.names[p.cursor], false)
	}
	return m, nil
}

// commitAddToCollection adds the picker's repositories to the named collection,
// creating it first when create is set. On a collision or a write failure the
// modal stays open with the error and nothing changes.
func (m *Model) commitAddToCollection(name string, create bool) (tea.Model, tea.Cmd) {
	p := m.picker
	if p == nil {
		return m, nil
	}
	if !m.collectionsReady() {
		m.picker = nil
		return m, nil
	}

	snap := m.coll.Clone()
	if create {
		if _, err := m.coll.Create(name); err != nil {
			m.coll = snap
			p.err = err.Error()
			return m, nil
		}
	}
	added, existing, err := m.coll.Add(name, membersOf(p.items))
	if err != nil {
		m.coll = snap
		p.err = err.Error()
		return m, nil
	}
	if err := m.saveCollections(snap); err != nil {
		p.err = err.Error()
		return m, nil
	}

	m.picker = nil
	m.clearSelection()
	m.status = fmt.Sprintf("added %d to %s (%d already members)", added, name, existing)
	m.clampCursor()
	return m, nil
}

// pickerView renders the add-to-collection modal: a centered bordered box over
// the list, mirroring the clone popup.
func (m *Model) pickerView() string {
	p := m.picker
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Add %s to collection", plural(len(p.items), "repo", "repos"))))
	b.WriteString("\n\n")

	if p.naming {
		b.WriteString(p.name.View())
		b.WriteString("\n")
		if p.err != "" {
			b.WriteString(dimStyle.Render(p.err) + "\n")
		}
		b.WriteString("\n" + dimStyle.Render("enter: create & add  esc: back"))
		return centerBox(m, b.String())
	}

	rows := make([]string, 0, len(p.names)+1)
	for i, name := range p.names {
		rows = append(rows, fmt.Sprintf("%s %s", name, dimStyle.Render(fmt.Sprintf("(%d)", p.counts[i]))))
	}
	rows = append(rows, newCollectionLabel)
	for i, row := range rows {
		prefix := "  "
		if i == p.cursor {
			prefix = "> "
			row = cursorStyle.Render(row)
		}
		b.WriteString(prefix + row + "\n")
	}
	if p.err != "" {
		b.WriteString(dimStyle.Render(p.err) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("enter: add  esc: cancel"))
	return centerBox(m, b.String())
}

// centerBox wraps a modal body in a rounded border and centers it when the
// terminal size is known.
func centerBox(m *Model, body string) string {
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Render(body)
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}
	return box
}

// plural renders a count with a singular or plural noun.
func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

// promptKind distinguishes the two one-shot collection modals.
type promptKind int

const (
	promptRename promptKind = iota // text entry: rename a collection
	promptDelete                   // confirmation: delete a collection
)

// collPrompt is a small modal acting on one collection: a name entry (rename)
// or a yes/no confirmation (delete).
type collPrompt struct {
	kind   promptKind
	target string
	input  textinput.Model
	err    string
}

// renameCollection opens the rename modal for the highlighted collection.
func (m *Model) renameCollection() (tea.Model, tea.Cmd) {
	if !m.collectionsReady() {
		return m, nil
	}
	row, ok := m.currentCollection()
	if !ok {
		return m, nil
	}
	ti := textinput.New()
	ti.Prompt = "name: "
	ti.SetValue(row.Name)
	ti.CursorEnd()
	ti.Focus()
	m.prompt = &collPrompt{kind: promptRename, target: row.Name, input: ti}
	return m, nil
}

// deleteCollection opens the delete confirmation for the highlighted collection.
func (m *Model) deleteCollection() (tea.Model, tea.Cmd) {
	if !m.collectionsReady() {
		return m, nil
	}
	row, ok := m.currentCollection()
	if !ok {
		return m, nil
	}
	m.prompt = &collPrompt{kind: promptDelete, target: row.Name}
	return m, nil
}

// handlePromptKey is the modal key map for the rename / delete prompts.
func (m *Model) handlePromptKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	p := m.prompt
	if msg.Type == tea.KeyCtrlC {
		m.quitting = true
		return m, tea.Quit
	}
	if p.kind == promptDelete {
		switch msg.String() {
		case "y", "Y":
			return m.commitDeleteCollection()
		case "n", "N", "esc", "q":
			m.prompt = nil
			return m, nil
		}
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEnter:
		return m.commitRenameCollection(strings.TrimSpace(p.input.Value()))
	case tea.KeyEsc:
		m.prompt = nil
		return m, nil
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return m, cmd
}

// commitRenameCollection renames the prompt's collection, keeping the modal open
// with an error when the name is taken or the write fails.
func (m *Model) commitRenameCollection(name string) (tea.Model, tea.Cmd) {
	p := m.prompt
	if p == nil || !m.collectionsReady() {
		m.prompt = nil
		return m, nil
	}
	snap := m.coll.Clone()
	if err := m.coll.Rename(p.target, name); err != nil {
		m.coll = snap
		p.err = err.Error()
		return m, nil
	}
	if err := m.saveCollections(snap); err != nil {
		p.err = err.Error()
		return m, nil
	}
	if m.selColl == p.target {
		m.selColl = name
	}
	m.status = fmt.Sprintf("renamed %s to %s", p.target, name)
	m.prompt = nil
	m.clampCursor()
	return m, nil
}

// commitDeleteCollection deletes the prompt's collection. It never touches a
// clone on disk: a collection is a view, not storage.
func (m *Model) commitDeleteCollection() (tea.Model, tea.Cmd) {
	p := m.prompt
	if p == nil || !m.collectionsReady() {
		m.prompt = nil
		return m, nil
	}
	snap := m.coll.Clone()
	if err := m.coll.Delete(p.target); err != nil {
		m.coll = snap
		p.err = err.Error()
		return m, nil
	}
	if err := m.saveCollections(snap); err != nil {
		p.err = err.Error()
		return m, nil
	}
	if m.selColl == p.target {
		m.selColl = ""
	}
	m.status = fmt.Sprintf("deleted collection %s", p.target)
	m.prompt = nil
	m.cursor = 0
	m.offset = 0
	m.clampCursor()
	return m, nil
}

// removeFromCollection drops the multi-selected members from the collection
// being browsed, or the highlighted member when nothing is selected (which is
// how an unresolved member is removed, since it cannot be selected).
func (m *Model) removeFromCollection() (tea.Model, tea.Cmd) {
	if m.selColl == "" || !m.collectionsReady() {
		return m, nil
	}
	items := m.selectedItems()
	if len(items) == 0 {
		it, ok := m.currentRepo()
		if !ok {
			return m, nil
		}
		items = []repoItem{it}
	}
	members := membersOf(items)
	if len(members) == 0 {
		m.status = "nothing to remove"
		return m, nil
	}

	snap := m.coll.Clone()
	removed, err := m.coll.Remove(m.selColl, members)
	if err != nil {
		m.coll = snap
		m.status = fmt.Sprintf("remove failed: %v", err)
		return m, nil
	}
	if err := m.saveCollections(snap); err != nil {
		m.status = fmt.Sprintf("remove failed: %v", err)
		return m, nil
	}
	m.clearSelection()
	m.status = fmt.Sprintf("removed %s from %s", plural(removed, "repo", "repos"), m.selColl)
	m.clampCursor()
	return m, nil
}

// promptView renders the rename / delete modal.
func (m *Model) promptView() string {
	p := m.prompt
	var b strings.Builder
	if p.kind == promptDelete {
		b.WriteString(titleStyle.Render("Delete collection " + p.target))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("This removes the collection only — no clone is touched."))
		if p.err != "" {
			b.WriteString("\n" + dimStyle.Render(p.err))
		}
		b.WriteString("\n\n" + dimStyle.Render("y: delete  n/esc: cancel"))
		return centerBox(m, b.String())
	}
	b.WriteString(titleStyle.Render("Rename collection " + p.target))
	b.WriteString("\n\n")
	b.WriteString(p.input.View())
	if p.err != "" {
		b.WriteString("\n" + dimStyle.Render(p.err))
	}
	b.WriteString("\n\n" + dimStyle.Render("enter: rename  esc: cancel"))
	return centerBox(m, b.String())
}

// unresolvedStatus explains why a collection member cannot be acted on. Such a
// member is kept in the collection (and removable with d), it just has no cached
// repository behind it yet.
func unresolvedStatus(it repoItem) string {
	if it.Provider == nil {
		return fmt.Sprintf("%s: provider not configured — fix config.yaml or remove the member", it.key())
	}
	return fmt.Sprintf("%s: not in the cache — refresh or fetch its owner first", it.key())
}

// startCollection opens a collection's repository list directly. It backs the
// --collection launch option; an unknown name is an error so the command fails
// loudly instead of opening an empty list.
func (m *Model) startCollection(name string) error {
	if m.coll == nil {
		if m.collErr != nil {
			return fmt.Errorf("collections unavailable: %w", m.collErr)
		}
		return errors.New("collections unavailable")
	}
	c, ok := m.coll.Get(name)
	if !ok {
		return fmt.Errorf("unknown collection %q", name)
	}
	m.selColl = c.Name
	m.selProvider = nil
	m.selOwner = ""
	m.flatAll = false
	m.nav = levelRepos
	m.cursor = 0
	m.offset = 0
	return nil
}
