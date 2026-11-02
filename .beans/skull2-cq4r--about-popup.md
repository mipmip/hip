---
# skull2-cq4r
title: about popup
status: completed
type: task
priority: normal
created_at: 2026-08-19T12:36:57Z
updated_at: 2026-11-02T21:46:34Z
parent: skull2-ok4c
---

contents:
- credits
- version
- build
- license
- github

## Summary of Changes

Shipped as `2026-11-02-add-about-popup` (commit 26e050c). The TUI now takes the resolved binary version from `cmd/hip`, and `?` opens a centered about popup with the version, Go runtime and platform, MIT copyright, GitHub URL, and credits. `?` or `esc` closes it and other keys are swallowed; while filtering, `?` stays a literal query character. The list, detail, and multiplex footers advertise the key, and the README key table documents it. Tests cover rendering, opening from list and detail, key swallowing, filter behavior, and footer hints.
