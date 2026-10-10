## Purpose

Turns a version tag on the `main` branch into a published GitHub Release with
ready-to-run clipt builds for Linux and Windows, so a release is one tag push
away and never built by hand.

## ADDED Requirements

### Requirement: Release trigger
A release run SHALL start when a tag whose name is `v` followed by three
dot-separated numbers (`vX.Y.Z`) is pushed to the GitHub repository. Pushes of
branches, and of tags with any other name, SHALL NOT start a release run.

#### Scenario: Version tag pushed
- **WHEN** the tag `v0.1.0` is pushed
- **THEN** a release run starts for the commit the tag points to

#### Scenario: Branch pushed
- **WHEN** new commits are pushed to `main` without a tag
- **THEN** no release run starts

#### Scenario: Non-version tag pushed
- **WHEN** a tag named `v1.0.0-rc1`, `1.0.0` or `nightly` is pushed
- **THEN** no release run starts

### Requirement: Only commits on main are released
The release run SHALL fail, publishing nothing, when the tagged commit is not
`main` or one of its ancestors. The run SHALL report that the tag is not on
`main` as the reason.

#### Scenario: Tag on main
- **WHEN** `v0.1.0` points to the current head of `main`
- **THEN** the run continues to the test gate

#### Scenario: Tag on an older main commit
- **WHEN** `v0.1.0` points to a commit that is an ancestor of `main`
- **THEN** the run continues to the test gate

#### Scenario: Tag on a side branch
- **WHEN** `v0.1.0` points to a commit that is not reachable from `main`
- **THEN** the run fails before building, no GitHub Release is created and
  the failure message says the tag is not on `main`

### Requirement: Test gate
Before any binary is built, the release run SHALL check that `gofmt` reports
no files, `go vet ./...` passes and `go test ./...` passes, using the Go
version declared in `go.mod`. If any check fails the run SHALL fail and no
GitHub Release SHALL be created.

#### Scenario: Checks pass
- **WHEN** the tagged commit is formatted, vets clean and all tests pass
- **THEN** the run proceeds to building the artifacts

#### Scenario: A test fails
- **WHEN** `go test ./...` fails on the tagged commit
- **THEN** the run fails, no archives are built and no GitHub Release is
  created

### Requirement: Release artifacts
A release SHALL contain exactly these two archives, each holding the clipt
binary built from the tagged commit for the x86-64 architecture without cgo,
plus `README.md`:

| Archive                           | Binary      |
|-----------------------------------|-------------|
| `clipt_X.Y.Z_linux_amd64.tar.gz`  | `clipt`     |
| `clipt_X.Y.Z_windows_amd64.zip`   | `clipt.exe` |

`X.Y.Z` is the tag name without the leading `v`. The release SHALL also
contain `checksums.txt` listing the SHA-256 sum of every archive.

#### Scenario: Assets of v0.1.0
- **WHEN** the release run for `v0.1.0` succeeds
- **THEN** the GitHub Release has the assets `clipt_0.1.0_linux_amd64.tar.gz`,
  `clipt_0.1.0_windows_amd64.zip` and `checksums.txt`, and no others

#### Scenario: Linux archive content
- **WHEN** `clipt_0.1.0_linux_amd64.tar.gz` is extracted on an x86-64 Linux
  machine
- **THEN** it yields an executable `clipt` that runs without any shared
  library beyond the system's C library, and a `README.md`

#### Scenario: Windows archive content
- **WHEN** `clipt_0.1.0_windows_amd64.zip` is extracted
- **THEN** it yields `clipt.exe` and `README.md`

#### Scenario: Checksums match
- **WHEN** the SHA-256 sum of a downloaded archive is compared with its line
  in `checksums.txt`
- **THEN** they are equal

### Requirement: Published release
On success the run SHALL publish a GitHub Release on the repository that is
named after the tag, is not a draft and is not marked as a pre-release, with
release notes listing the commit messages between the previous `vX.Y.Z` tag
and this one, newest first, excluding merge commits. For the first tag the
notes SHALL list all commits.

#### Scenario: Second release
- **WHEN** `v0.2.0` is tagged three commits after `v0.1.0`
- **THEN** the release `v0.2.0` is published with notes listing those three
  commits

#### Scenario: First release
- **WHEN** `v0.1.0` is the first version tag in the repository
- **THEN** the release `v0.1.0` is published with notes listing every commit
  on `main`

### Requirement: Re-running a tag
Pushing the same `vX.Y.Z` tag again (after deleting it) SHALL replace the
assets of the existing release rather than fail or create a duplicate.

#### Scenario: Tag recreated after a failed upload
- **WHEN** `v0.1.0` is deleted and pushed again pointing to a commit on `main`
- **THEN** the run succeeds and the release `v0.1.0` holds the freshly built
  assets
