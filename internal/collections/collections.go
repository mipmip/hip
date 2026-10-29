package collections

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Version is the current collections.yaml schema version. A file without a
// version is read as version 1 so hand-written files stay valid.
const Version = 1

// Member identifies one repository in a collection. Provider is the unique
// providers[].name key from the configuration (not the display short), so a
// collection can hold repositories from different owners and providers.
type Member struct {
	// Provider is the configured provider name owning this repository.
	Provider string `yaml:"provider"`
	// Owner is the owning user or organization.
	Owner string `yaml:"owner"`
	// Repo is the repository name.
	Repo string `yaml:"repo"`
}

// SameAs reports whether two members identify the same repository. Provider
// names are compared exactly (they are config keys); owner and repository names
// are compared case-insensitively, because forges differ in case handling.
func (m Member) SameAs(other Member) bool {
	return m.Provider == other.Provider &&
		strings.EqualFold(m.Owner, other.Owner) &&
		strings.EqualFold(m.Repo, other.Repo)
}

// Key renders the member as "provider owner/repo" for display and logging.
func (m Member) Key() string { return m.Provider + " " + m.Owner + "/" + m.Repo }

// Collection is one named, user-defined grouping of repositories. Members are
// held in insertion order.
type Collection struct {
	// Name is the collection's unique (case-insensitive) name.
	Name string `yaml:"name"`
	// Members are the collection's repositories, in the order they were added.
	Members []Member `yaml:"members"`
}

// Has reports whether the collection already contains the given member.
func (c *Collection) Has(m Member) bool {
	for _, existing := range c.Members {
		if existing.SameAs(m) {
			return true
		}
	}
	return false
}

// Store is the persisted set of collections, held as a list so the file keeps a
// stable order across writes (a map would be marshaled key-sorted).
type Store struct {
	// Version is the schema version of the persisted file.
	Version int `yaml:"version"`
	// Collections are the user's collections, in file order.
	Collections []Collection `yaml:"collections"`
}

// DefaultPath returns the resolved collections path using $XDG_CONFIG_HOME when
// set, otherwise ~/.config/hip/collections.yaml. It sits beside config.yaml
// in the configuration directory, but is writable: the TUI edits it at runtime.
func DefaultPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "hip", "collections.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "hip", "collections.yaml"), nil
}

// Load reads the collections file. A missing file yields an empty store and no
// error (and no file is created); a malformed or internally inconsistent file
// returns an error naming the path.
func Load(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Store{Version: Version}, nil
		}
		return nil, fmt.Errorf("reading collections %s: %w", path, err)
	}
	var s Store
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing collections %s: %w", path, err)
	}
	if s.Version == 0 {
		s.Version = Version
	}
	if err := s.validate(); err != nil {
		return nil, fmt.Errorf("invalid collections %s: %w", path, err)
	}
	return &s, nil
}

// validate rejects a store that cannot round-trip meaningfully: blank or
// duplicate collection names and members missing an identifying field.
func (s *Store) validate() error {
	seen := make(map[string]bool, len(s.Collections))
	for i := range s.Collections {
		c := &s.Collections[i]
		name := strings.TrimSpace(c.Name)
		if name == "" {
			return fmt.Errorf("collection %d has an empty name", i+1)
		}
		key := strings.ToLower(name)
		if seen[key] {
			return fmt.Errorf("duplicate collection name %q", name)
		}
		seen[key] = true
		for j, m := range c.Members {
			if m.Provider == "" || m.Owner == "" || m.Repo == "" {
				return fmt.Errorf("collection %q member %d is missing provider, owner or repo", name, j+1)
			}
		}
	}
	return nil
}

// Save writes the store atomically: marshal, write a temporary file in the
// target directory, sync, then rename into place, so an interrupted write can
// never leave a partial collections.yaml. The parent directory is created as
// needed.
func Save(path string, s *Store) error {
	if s.Version == 0 {
		s.Version = Version
	}
	if err := s.validate(); err != nil {
		return fmt.Errorf("refusing to save invalid collections: %w", err)
	}

	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("marshaling collections: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("marshaling collections: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, "collections.*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp collections file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.WriteString(buf.String()); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp collections file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing temp collections file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp collections file: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("setting collections file mode: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("renaming collections file into place: %w", err)
	}
	return nil
}

// Clone returns a deep copy of the store, so a caller can mutate it and roll
// back to the snapshot when persisting fails.
func (s *Store) Clone() *Store {
	out := &Store{Version: s.Version}
	if s.Collections == nil {
		return out
	}
	out.Collections = make([]Collection, len(s.Collections))
	for i, c := range s.Collections {
		out.Collections[i] = Collection{Name: c.Name}
		if c.Members != nil {
			out.Collections[i].Members = append([]Member(nil), c.Members...)
		}
	}
	return out
}

// Names returns the collection names in file order.
func (s *Store) Names() []string {
	out := make([]string, 0, len(s.Collections))
	for _, c := range s.Collections {
		out = append(out, c.Name)
	}
	return out
}

// indexOf returns the index of the named collection, or -1. Names are matched
// case-insensitively so "Terraform" and "terraform" are the same collection.
func (s *Store) indexOf(name string) int {
	want := strings.ToLower(strings.TrimSpace(name))
	for i, c := range s.Collections {
		if strings.ToLower(strings.TrimSpace(c.Name)) == want {
			return i
		}
	}
	return -1
}

// Get returns the named collection. The pointer aliases the store and stays
// valid until the next mutation.
func (s *Store) Get(name string) (*Collection, bool) {
	i := s.indexOf(name)
	if i < 0 {
		return nil, false
	}
	return &s.Collections[i], true
}

// Create adds an empty collection, appending it after the existing ones. An
// empty or already-used name is an error: a collision is a typo, not a request
// to merge two collections.
func (s *Store) Create(name string) (*Collection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("collection name must not be empty")
	}
	if s.indexOf(name) >= 0 {
		return nil, fmt.Errorf("collection %q already exists", name)
	}
	s.Collections = append(s.Collections, Collection{Name: name})
	return &s.Collections[len(s.Collections)-1], nil
}

// Add appends members to a collection, skipping repositories that are already
// members. It reports how many were added and how many were already present, so
// the caller can say "added 5, 2 already members".
func (s *Store) Add(name string, members []Member) (added, existing int, err error) {
	i := s.indexOf(name)
	if i < 0 {
		return 0, 0, fmt.Errorf("no such collection %q", name)
	}
	c := &s.Collections[i]
	for _, m := range members {
		if m.Provider == "" || m.Owner == "" || m.Repo == "" {
			return added, existing, fmt.Errorf("member %q is missing provider, owner or repo", m.Key())
		}
		if c.Has(m) {
			existing++
			continue
		}
		c.Members = append(c.Members, m)
		added++
	}
	return added, existing, nil
}

// Remove drops members from a collection, preserving the order of the rest. A
// collection that becomes empty still exists. It reports how many members were
// actually removed.
func (s *Store) Remove(name string, members []Member) (removed int, err error) {
	i := s.indexOf(name)
	if i < 0 {
		return 0, fmt.Errorf("no such collection %q", name)
	}
	c := &s.Collections[i]
	kept := c.Members[:0:0]
	for _, have := range c.Members {
		drop := false
		for _, m := range members {
			if have.SameAs(m) {
				drop = true
				break
			}
		}
		if drop {
			removed++
			continue
		}
		kept = append(kept, have)
	}
	c.Members = kept
	return removed, nil
}

// Rename changes a collection's name in place, keeping its members and its
// position in the file. Renaming onto a name another collection already uses is
// an error; changing only the case of the collection's own name is allowed.
func (s *Store) Rename(oldName, newName string) error {
	i := s.indexOf(oldName)
	if i < 0 {
		return fmt.Errorf("no such collection %q", oldName)
	}
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return errors.New("collection name must not be empty")
	}
	if j := s.indexOf(newName); j >= 0 && j != i {
		return fmt.Errorf("collection %q already exists", newName)
	}
	s.Collections[i].Name = newName
	return nil
}

// Delete removes a collection. It never touches any clone on disk: a collection
// is a view, not storage.
func (s *Store) Delete(name string) error {
	i := s.indexOf(name)
	if i < 0 {
		return fmt.Errorf("no such collection %q", name)
	}
	s.Collections = append(s.Collections[:i], s.Collections[i+1:]...)
	return nil
}
