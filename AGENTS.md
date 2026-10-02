# AGENTS.md

clipt is a minimal, keyboard-driven TUI note-taking app written in Go. Notes are
markdown files in a single flat directory. Full product context, tech stack and
artifact rules live in `openspec/config.yaml`.

## Workflow

- Use OpenSpec: propose a change -> apply -> archive.
- Read the relevant specs in `openspec/specs/` before changing behavior.

## Version control

- Use Jujutsu (`jj`), colocated with git: `jj commit -m`, `jj describe`, `jj split`.
- Never run git commands that change the repo (commit, rebase, reset, checkout, ...).

## Go conventions

- `gofmt`, `go vet ./...` and `go test ./...` must pass.
- Layout: entrypoint in `cmd/clipt`, packages under `internal/`.

## File safety

- Never create, overwrite or delete files outside the configured notes directory
  (the config file is the only exception).
- Tests touching file I/O use a temp dir (`t.TempDir()`), never real notes.
