// Package collections persists named, user-defined repository collections in
// ~/.config/huphop/collections.yaml: hand-curated groupings of repositories that
// may span owners, organizations and providers.
//
// A collection is a view, never storage: it holds references to repositories
// (provider name + owner + repo), so adding or removing members never touches a
// clone on disk. Membership is insertion-ordered and the file is written
// atomically, keeping it hand-editable and diff-stable.
//
// Implemented in milestone "07 Workspaces and more". See openspec change
// add-repo-collections.
package collections
