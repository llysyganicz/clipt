## Context

See proposal.md for motivation. Current state that shapes the approach:

- The repository is jj colocated with git. There is no bookmark at all (git
  reports a detached `HEAD`), no remote and no `.github/` directory. The
  working copy's `clipt` binary is already ignored.
- `go.mod` declares `go 1.26.8`; the local toolchain is Go 1.27.1. The
  module path is `github.com/llysyganicz/clipt`, which already matches the
  intended GitHub location.
- The entrypoint is `cmd/clipt`; nothing in the binary knows its version.
- `gh` 2.101 is installed and authenticated as `llysyganicz` with the `repo`
  scope. jj 0.45 creates tags with `jj tag set <name> -r <rev>` (there is no
  `jj git tag`; `jj git` only has clone, fetch, push, remote and friends) and
  pushes them with `jj git push --tag <name>`. Both were tried on a scratch
  tag and deleted again. GoReleaser is not installed locally.
- AGENTS.md forbids running git commands that change the repository; every
  repository mutation here goes through `jj` or `gh`.
- Latest versions at planning time: `actions/checkout` v7, `actions/setup-go`
  v7, `goreleaser/goreleaser-action` v7, GoReleaser v2.18.

## Goals / Non-Goals

**Goals:**

- One tag push produces a complete release with no manual steps on GitHub.
- A tag on the wrong commit fails loudly instead of releasing.
- Everything about the pipeline is in two small declarative files that can be
  validated locally without pushing a tag.

**Non-Goals:**

- See proposal.md Non-goals. In addition: no caching of Go modules in CI
  (the build takes under two minutes and a cache adds a failure mode), and no
  reusable workflows or composite actions.

## Decisions

### Build and publish with GoReleaser rather than a hand-written matrix

A `.goreleaser.yaml` declares the OS/arch matrix, archive formats, naming,
checksums, changelog and the GitHub Release in about forty lines, and the
same file drives a local dry run (`release --snapshot --clean`). The
alternative, a GitHub Actions matrix job with `go build` per target plus a
release step uploading assets, needs shell for naming, tar/zip, checksums
and release notes and is harder to test locally. GoReleaser is pinned to the
v2 series through the action's `version: "~> v2"` input and runs only in CI;
it never enters `go.mod`.

Config outline (`version: 2`):

- `project_name: clipt`; one build with `main: ./cmd/clipt`, `binary: clipt`,
  `env: [CGO_ENABLED=0]`, `goos: [linux, windows]`, `goarch: [amd64]`,
  `ldflags: ["-s -w"]`. The
  explicit ldflags replace GoReleaser's default `-X main.version=...` flags,
  which would otherwise silently point at variables the binary does not have.
- `archives`: `formats: [tar.gz]`, `format_overrides: [{goos: windows,
  formats: [zip]}]`, `name_template:
  "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"`, `files:
  [README.md]`. `{{ .Version }}` is the tag without `v`, giving the names in
  the spec.
- `checksum: {name_template: checksums.txt}` (SHA-256 is the default).
- `changelog: {use: git, sort: asc, filters: {exclude: ["^Merge "]}}`.
- `release: {github: {owner: llysyganicz, name: clipt}, mode: replace,
  prerelease: false, draft: false}`. The repository is named explicitly
  because `goreleaser check` and snapshot builds otherwise need a git remote
  to infer it, and the remote only exists after hosting. `replace`
  satisfies the re-run requirement; `prerelease: false` is explicit because
  the default `auto` is irrelevant once only `vX.Y.Z` tags reach the run.

### Workflow shape

`.github/workflows/release.yml`, triggered by
`on: push: tags: ["v[0-9]+.[0-9]+.[0-9]+"]` (GitHub's filter syntax; `+`
is supported). `permissions: contents: write` is the only grant. One job on
`ubuntu-latest` with these steps:

1. `actions/checkout@v7` with `fetch-depth: 0` so the changelog and the
   `main` check see history.
2. **Main check**: `git fetch origin main` then
   `git merge-base --is-ancestor "$GITHUB_SHA" origin/main || { echo "::error::tag
   $GITHUB_REF_NAME is not on main"; exit 1; }`. Tag push events carry no
   branch, so ancestry is the only way to express "tagged on main". This step
   runs before setup-go so a wrong tag fails in seconds.
3. `actions/setup-go@v7` with `go-version-file: go.mod`, so CI tracks the
   module's Go version instead of a second pinned copy.
4. **Gate**: `test -z "$(gofmt -l .)"`, `go vet ./...`, `go test ./...`.
5. `goreleaser/goreleaser-action@v7` with `distribution: goreleaser`,
   `version: "~> v2"`, `args: release --clean`, `GITHUB_TOKEN:
   ${{ secrets.GITHUB_TOKEN }}`.

Alternative considered: two jobs (test, then release) for a clearer UI. One
job avoids a second checkout and setup-go for a pipeline this short.

### GoReleaser's own `main` filter is not used

GoReleaser can refuse tags not on a branch, but only via `git.ignore_tags` /
hooks that are less direct than the one-line `merge-base` step, and the step
fails before Go is even installed. Keep the check in the workflow.

### Hosting through jj and gh

- `gh repo create llysyganicz/clipt --private --description "..."` creates an
  empty repository (no `--source`/`--push`, which would run git). The
  description is the first README sentence.
- `jj bookmark create main -r <head>` puts `main` on the current head.
- `jj git remote add origin https://github.com/llysyganicz/clipt.git`
  matches the `https` protocol `gh` is configured for.
- `jj git push --bookmark main --allow-new` pushes; GitHub makes the first
  pushed branch the default, so `main` becomes the default branch without a
  settings call. Verify with `gh repo view --json defaultBranchRef`.

Alternative: `gh repo create --source . --push` is shorter but shells out to
`git push`, which AGENTS.md rules out.

### Releasing routine

`jj tag set vX.Y.Z -r main` then `jj git push --tag vX.Y.Z`. jj creates
lightweight tags, which GoReleaser accepts. The README documents the routine
and the "tag must be on main" rule.

### README structure

Sections, in order: one-paragraph description; Modes (finder, preview,
edit) with a key table per mode taken from `openspec/specs/*/spec.md`;
Notes directory and config (file location per platform, single
`notes_dir` key); Install (download from Releases, extract, put `clipt` on
`PATH`; `go install` is omitted because the repository is private and would
need `GOPRIVATE` plus credentials); Build from source (`go build
./cmd/clipt`); Development (jj, OpenSpec workflow, `gofmt`/`go vet`/`go
test`); Releasing (the two jj commands and what the pipeline does). No
badges and no screenshots in this change.

### Local validation without pushing a tag

`go run github.com/goreleaser/goreleaser/v2@v2.18.3 check` validates the
config and `... release --snapshot --clean` builds every archive into
`dist/` without touching GitHub. `dist/` is added to `.gitignore`. The
workflow file is checked with `gh workflow view` only after the push; before
that, a YAML parse (`python3 -c 'import yaml,...'`) catches syntax errors.
Full end-to-end proof is the first real tag, which is the user's call.

## Risks / Trade-offs

- [GoReleaser action or schema changes break the run] → the action is pinned
  to a major and GoReleaser to `~> v2`; `goreleaser check` in the tasks
  pins v2.18.3 locally so a schema problem shows up before the push.
- [`fetch-depth: 0` on a repository that grows large slows checkout] → the
  repository is tiny; revisit if checkout ever exceeds a few seconds.
- [Tag pushed to a commit that is on `main` locally but `main` was not pushed
  first] → the ancestry check fails with a clear message; push `main`, then
  re-push the tag (the `replace` release mode handles the retry).
- [`go-version-file` requests `1.26.8`, which setup-go may need to download
  rather than find preinstalled] → adds a few seconds; acceptable.
- [The tag filter lets an unintended name through] → GitHub filter patterns
  match the whole ref name and support `+`, so `v[0-9]+.[0-9]+.[0-9]+`
  rejects `v1.0.0-rc1`, `1.0.0` and `v1.2.3.4`. A negative test cannot run
  in CI without pushing junk tags, so `actionlint` checks the syntax and the
  design relies on the documented semantics.
- [Private repository: release assets are only downloadable while
  authenticated] → intended; the README's install section says so.

## Migration Plan

All additive. Rollback is deleting the three new files and, if wanted, the
GitHub repository (`gh repo delete`, user-confirmed). A failed release run
publishes nothing, so there is nothing to undo on GitHub after a failure.

## Open Questions

None.
