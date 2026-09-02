package collections

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mem is a terse Member constructor for tests.
func mem(provider, owner, repo string) Member {
	return Member{Provider: provider, Owner: owner, Repo: repo}
}

func TestDefaultPathUsesXDGThenHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if want := filepath.Join("/tmp/xdg", "huphop", "collections.yaml"); got != want {
		t.Fatalf("DefaultPath with XDG = %q, want %q", got, want)
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	got, err = DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath without XDG: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}
	if want := filepath.Join(home, ".config", "huphop", "collections.yaml"); got != want {
		t.Fatalf("DefaultPath without XDG = %q, want %q", got, want)
	}
}

func TestLoadMissingFileYieldsEmptyStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "collections.yaml")

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load missing file: %v", err)
	}
	if len(s.Collections) != 0 {
		t.Fatalf("expected no collections, got %d", len(s.Collections))
	}
	if s.Version != Version {
		t.Fatalf("Version = %d, want %d", s.Version, Version)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Load must not create the file, stat err = %v", err)
	}
}

func TestLoadMalformedFileErrorsWithPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "collections.yaml")
	if err := os.WriteFile(path, []byte("collections: [oops\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected a parse error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("error %q should name the path %q", err, path)
	}
}

func TestLoadRejectsDuplicateAndEmptyNames(t *testing.T) {
	dir := t.TempDir()

	dup := filepath.Join(dir, "dup.yaml")
	if err := os.WriteFile(dup, []byte(`version: 1
collections:
  - name: terraform
    members: []
  - name: Terraform
    members: []
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dup); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate names should error, got %v", err)
	}

	blank := filepath.Join(dir, "blank.yaml")
	if err := os.WriteFile(blank, []byte("version: 1\ncollections:\n  - name: \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(blank); err == nil || !strings.Contains(err.Error(), "empty name") {
		t.Fatalf("empty name should error, got %v", err)
	}

	badMember := filepath.Join(dir, "member.yaml")
	if err := os.WriteFile(badMember, []byte(`version: 1
collections:
  - name: broken
    members:
      - provider: gh
        owner: org1
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(badMember); err == nil || !strings.Contains(err.Error(), "missing provider") {
		t.Fatalf("member without repo should error, got %v", err)
	}
}

func TestLoadDefaultsMissingVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "collections.yaml")
	if err := os.WriteFile(path, []byte("collections:\n  - name: hand-written\n    members: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Version != Version {
		t.Fatalf("Version = %d, want %d", s.Version, Version)
	}
}

func TestSaveLoadRoundTripPreservesOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "collections.yaml")

	s := &Store{}
	for _, name := range []string{"terraform-modules", "dotfiles", "client-acme"} {
		if _, err := s.Create(name); err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
	}
	members := []Member{
		mem("github-mipmip", "org1", "terraform-aws"),
		mem("codeberg-pim", "pim", "tf-modules-dns"),
		mem("gitlab-work", "infra", "tf-network"),
	}
	if _, _, err := s.Add("terraform-modules", members); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := Save(path, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if want := []string{"terraform-modules", "dotfiles", "client-acme"}; strings.Join(got.Names(), ",") != strings.Join(want, ",") {
		t.Fatalf("collection order = %v, want %v", got.Names(), want)
	}
	if got.Version != Version {
		t.Fatalf("Version = %d, want %d", got.Version, Version)
	}
	c, ok := got.Get("terraform-modules")
	if !ok {
		t.Fatal("terraform-modules missing after round-trip")
	}
	if len(c.Members) != len(members) {
		t.Fatalf("member count = %d, want %d", len(c.Members), len(members))
	}
	for i, m := range members {
		if !c.Members[i].SameAs(m) {
			t.Fatalf("member %d = %v, want %v (insertion order must be preserved)", i, c.Members[i], m)
		}
	}
}

func TestSaveWritesReadableYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "collections.yaml")
	s := &Store{}
	if _, err := s.Create("terraform-modules"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Add("terraform-modules", []Member{mem("github-mipmip", "org1", "terraform-aws")}); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, s); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"version: 1", "name: terraform-modules", "provider: github-mipmip"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("collections.yaml missing %q:\n%s", want, data)
		}
	}
}

func TestSaveRefusesInvalidStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "collections.yaml")
	s := &Store{Collections: []Collection{{Name: "a"}, {Name: "A"}}}
	if err := Save(path, s); err == nil {
		t.Fatal("expected Save to refuse a duplicate-name store")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("refused Save must not create the file, stat err = %v", err)
	}
}

func TestAddIsIdempotentAndCounts(t *testing.T) {
	s := &Store{}
	if _, err := s.Create("tf"); err != nil {
		t.Fatal(err)
	}
	first := []Member{
		mem("gh", "org1", "a"),
		mem("gh", "org1", "b"),
	}
	added, existing, err := s.Add("tf", first)
	if err != nil || added != 2 || existing != 0 {
		t.Fatalf("first Add = (%d, %d, %v), want (2, 0, nil)", added, existing, err)
	}

	mixed := []Member{
		mem("gh", "org1", "b"), // already a member
		mem("gh", "org2", "c"),
	}
	added, existing, err = s.Add("tf", mixed)
	if err != nil || added != 1 || existing != 1 {
		t.Fatalf("second Add = (%d, %d, %v), want (1, 1, nil)", added, existing, err)
	}

	c, _ := s.Get("tf")
	if len(c.Members) != 3 {
		t.Fatalf("member count = %d, want 3 (no duplicates)", len(c.Members))
	}

	if _, _, err := s.Add("nope", first); err == nil {
		t.Fatal("Add to an unknown collection should error")
	}
	if _, _, err := s.Add("tf", []Member{{Provider: "gh", Owner: "org1"}}); err == nil {
		t.Fatal("Add with an incomplete member should error")
	}
}

func TestMembershipIsCaseInsensitiveButProviderSensitive(t *testing.T) {
	s := &Store{}
	if _, err := s.Create("tf"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Add("tf", []Member{mem("github-mipmip", "Org1", "Terraform-AWS")}); err != nil {
		t.Fatal(err)
	}

	// Same repository in different case is the same member.
	added, existing, err := s.Add("tf", []Member{mem("github-mipmip", "org1", "terraform-aws")})
	if err != nil || added != 0 || existing != 1 {
		t.Fatalf("case-different Add = (%d, %d, %v), want (0, 1, nil)", added, existing, err)
	}

	// Same owner/repo on another provider is a different member.
	added, existing, err = s.Add("tf", []Member{mem("codeberg-pim", "org1", "terraform-aws")})
	if err != nil || added != 1 || existing != 0 {
		t.Fatalf("other-provider Add = (%d, %d, %v), want (1, 0, nil)", added, existing, err)
	}

	c, _ := s.Get("tf")
	if len(c.Members) != 2 {
		t.Fatalf("member count = %d, want 2", len(c.Members))
	}
	if !c.Has(mem("github-mipmip", "ORG1", "terraform-AWS")) {
		t.Fatal("Has should match case-insensitively")
	}
	if c.Members[0].Key() != "github-mipmip Org1/Terraform-AWS" {
		t.Fatalf("Key = %q, want the stored casing", c.Members[0].Key())
	}
}

func TestCreateRejectsEmptyAndDuplicateNames(t *testing.T) {
	s := &Store{}
	if _, err := s.Create("  tf  "); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if names := s.Names(); names[0] != "tf" {
		t.Fatalf("name = %q, want the trimmed %q", names[0], "tf")
	}
	if _, err := s.Create("   "); err == nil {
		t.Fatal("blank name should error")
	}
	if _, err := s.Create("TF"); err == nil {
		t.Fatal("case-different duplicate should error")
	}
	if _, ok := s.Get("missing"); ok {
		t.Fatal("Get of an unknown collection should report false")
	}
}

func TestRemovePreservesOrderAndKeepsEmptyCollection(t *testing.T) {
	s := &Store{}
	if _, err := s.Create("tf"); err != nil {
		t.Fatal(err)
	}
	all := []Member{
		mem("gh", "o", "a"),
		mem("gh", "o", "b"),
		mem("gh", "o", "c"),
	}
	if _, _, err := s.Add("tf", all); err != nil {
		t.Fatal(err)
	}

	removed, err := s.Remove("tf", []Member{mem("gh", "O", "B"), mem("gh", "o", "zz")})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	c, _ := s.Get("tf")
	if len(c.Members) != 2 || c.Members[0].Repo != "a" || c.Members[1].Repo != "c" {
		t.Fatalf("members = %v, want a then c", c.Members)
	}

	if _, err := s.Remove("tf", all); err != nil {
		t.Fatalf("Remove all: %v", err)
	}
	c, ok := s.Get("tf")
	if !ok {
		t.Fatal("an emptied collection must still exist")
	}
	if len(c.Members) != 0 {
		t.Fatalf("expected no members, got %d", len(c.Members))
	}
	if _, err := s.Remove("nope", all); err == nil {
		t.Fatal("Remove from an unknown collection should error")
	}
}

func TestRenameInPlaceAndCollision(t *testing.T) {
	s := &Store{}
	for _, n := range []string{"a", "b", "c"} {
		if _, err := s.Create(n); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := s.Add("b", []Member{mem("gh", "o", "r")}); err != nil {
		t.Fatal(err)
	}

	if err := s.Rename("b", "beta"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if names := strings.Join(s.Names(), ","); names != "a,beta,c" {
		t.Fatalf("names = %q, want a,beta,c (position preserved)", names)
	}
	c, ok := s.Get("beta")
	if !ok || len(c.Members) != 1 {
		t.Fatalf("renamed collection lost its members: %v %v", ok, c)
	}

	if err := s.Rename("beta", "c"); err == nil {
		t.Fatal("rename onto an existing name should error")
	}
	if names := strings.Join(s.Names(), ","); names != "a,beta,c" {
		t.Fatalf("a refused rename must change nothing, got %q", names)
	}
	// Changing only the case of a collection's own name is allowed.
	if err := s.Rename("beta", "Beta"); err != nil {
		t.Fatalf("case-only rename: %v", err)
	}
	if err := s.Rename("Beta", "  "); err == nil {
		t.Fatal("rename to a blank name should error")
	}
	if err := s.Rename("nope", "x"); err == nil {
		t.Fatal("rename of an unknown collection should error")
	}
}

func TestDeleteRemovesOnlyThatCollection(t *testing.T) {
	s := &Store{}
	for _, n := range []string{"a", "b", "c"} {
		if _, err := s.Create(n); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Delete("b"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if names := strings.Join(s.Names(), ","); names != "a,c" {
		t.Fatalf("names = %q, want a,c", names)
	}
	if err := s.Delete("b"); err == nil {
		t.Fatal("deleting an unknown collection should error")
	}
}

func TestSaveIsAtomicAndIgnoresStrayTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "collections.yaml")

	s := &Store{}
	if _, err := s.Create("tf"); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// A leftover temp file from an interrupted write is not part of the store.
	stray := filepath.Join(dir, "collections.stray.tmp")
	if err := os.WriteFile(stray, []byte("version: 1\ncollections:\n  - name: garbage\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if names := strings.Join(got.Names(), ","); names != "tf" {
		t.Fatalf("names = %q, want tf (stray temp file must be ignored)", names)
	}

	// A successful Save leaves no temp files of its own behind.
	if err := Save(path, got); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	tmpCount := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "collections.") && strings.HasSuffix(e.Name(), ".tmp") {
			tmpCount++
		}
	}
	if tmpCount != 1 {
		t.Fatalf("temp files in dir = %d, want only the stray one", tmpCount)
	}
}

func TestCloneIsADeepCopy(t *testing.T) {
	s := &Store{}
	if _, err := s.Create("tf"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Add("tf", []Member{mem("gh", "o", "a")}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("empty"); err != nil {
		t.Fatal(err)
	}

	snap := s.Clone()
	if _, _, err := s.Add("tf", []Member{mem("gh", "o", "b")}); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("empty"); err != nil {
		t.Fatal(err)
	}

	c, ok := snap.Get("tf")
	if !ok || len(c.Members) != 1 {
		t.Fatalf("snapshot should keep its own members, got %v", c)
	}
	if _, ok := snap.Get("empty"); !ok {
		t.Fatal("snapshot should still hold the deleted collection")
	}
	if snap.Version != s.Version {
		t.Fatalf("snapshot version = %d, want %d", snap.Version, s.Version)
	}

	// An empty store clones without allocating a members slice.
	empty := (&Store{Version: Version}).Clone()
	if len(empty.Collections) != 0 {
		t.Fatalf("empty clone should hold no collections, got %d", len(empty.Collections))
	}
}
