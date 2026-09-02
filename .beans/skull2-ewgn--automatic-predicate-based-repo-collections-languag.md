---
# skull2-ewgn
title: automatic predicate-based repo collections (language, topics, cloned)
status: draft
type: feature
created_at: 2026-09-02T15:09:02Z
updated_at: 2026-09-02T15:09:02Z
parent: skull2-ok4c
---

Split off from `skull2-289n` (user-defined repo collections), which now covers
only hand-curated collections. This bean is the *automatic* side: collections
whose membership is computed from a predicate instead of enumerated by hand.

Candidates:

- language (e.g. "Nix")
- forge topics
- archived / forked / visibility — already covered by the existing facets, so
  the real question is whether they deserve to *also* appear as collections
- cloned / not-cloned locally — cheap, already computed as `repoItem.Cloned`,
  and not yet available as a facet

## The blocker

`Language` and `Topics` live in `provider.Details` (tier 2), which is fetched
lazily per repo and cached one JSON file per repo under
`~/.cache/huphop/<provider>/details/`. So a language or topic taxonomy can only
see repos whose details have already been fetched — it would be silently
incomplete, which for a navigation aid is worse than absent.

Options to resolve before this is proposable:

1. Opportunistic — only use cached details, and label the collection as partial.
2. A details sweep (`hup refresh --details`) walking every repo: hundreds of API
   calls, rate limits, but then the taxonomy is real.
3. Ship only the cheap predicates first (cloned/not-cloned, and the existing
   facet-backed ones), and leave language/topics until a sweep exists.

## Design note

Predicate-defined collections need no storage, no CRUD and no member identity —
they are computed on every render. They likely want to reuse the facet
machinery (or be expressible as saved facet+filter queries) rather than the
`collections.yaml` store from skull2-289n.
