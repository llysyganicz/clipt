## Why

clipt exists only as a local jj/git repository: there is no remote, no `main`
bookmark and no way to hand a build to another machine except copying the
binary by hand. With the editor done the app is usable, so now is the time to
host it on GitHub and let a version tag produce downloadable builds for the two
platforms it is used on day to day.

## What Changes

- Host the repository as the private GitHub repository
  `llysyganicz/clipt` with `main` as the default branch. Create the `main`
  bookmark in jj at the current head, register the remote as `origin` and
  push through jj.
- Add a `README.md` at the repository root describing what clipt is, the
  three modes and their keys, where the config file lives, how to install a
  release build and how to build and test from source.
- Add a release pipeline as a GitHub Actions workflow that runs only when a
  tag of the form `vX.Y.Z` is pushed. The workflow refuses to release a tag
  whose commit is not on `main`, runs `gofmt`, `go vet` and `go test` as a
  gate, then builds and publishes a GitHub Release.
- Each release carries two archives, Linux `amd64` (`.tar.gz`) and Windows
  `amd64` (`.zip`), each holding the `clipt` binary and the README, plus a
  `checksums.txt` with SHA-256 sums. Release notes list the
  commits since the previous tag.
- Add `dist/` to `.gitignore` for local snapshot builds of the pipeline.

## Capabilities

### New Capabilities

- `release`: when a release is produced (tag pattern, `main`-only rule, test
  gate), which artifacts it contains (platforms, archive formats, names,
  checksums) and what the published GitHub Release looks like.

### Modified Capabilities

None. The app's behaviour does not change; hosting and the README are
one-time setup and documentation without spec-level behaviour.

## Non-goals

- A CI workflow on every push or pull request. Tests run in the release
  workflow as a gate; continuous checks are a later change once there is a
  reason for them (collaborators, pull requests).
- Pre-release tags such as `v1.0.0-rc1`; only `vX.Y.Z` triggers a release.
- macOS builds. The context lists macOS as a supported platform, but the
  request asks for Linux and Windows; adding `darwin` is a one-line follow-up
  once macOS has had its manual pass.
- `arm64` and 32-bit builds for either platform; `amd64` is the only
  architecture released.
- A `--version` flag or any version string compiled into the binary; the
  release tag is the only version identifier.
- A `LICENSE` file; the repository is private.
- Package-manager distribution (Homebrew, Scoop, winget, AUR, `.deb`).
- Code signing or notarisation of binaries.
- Branch protection rules and other GitHub repository settings beyond
  visibility and default branch.

## Impact

- New files: `README.md`, `.github/workflows/release.yml`,
  `.goreleaser.yaml`; `.gitignore` gains `dist/`.
- Repository state: a `main` bookmark, an `origin` remote and a private
  GitHub repository under the user's account. Requires an authenticated `gh`
  with the `repo` scope (already present on this machine).
- No Go code changes; `go.mod` is unchanged. GoReleaser runs only in CI and
  is not a module dependency.
- Releasing becomes a two-command routine in jj: `jj tag set vX.Y.Z -r main`
  then `jj git push --tag vX.Y.Z`.
