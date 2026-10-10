## 1. Repository hygiene

- [x] 1.1 Add `dist/` to `.gitignore` and verify `jj status` does not list a freshly created empty `dist/` directory (remove the directory afterwards)

## 2. README

- [x] 2.1 Write `README.md` with the sections from design.md (description, Modes with a key table per mode, Notes directory and config, Install from Releases, Build from source, Development, Releasing) and verify every key in the tables matches the requirement text in `openspec/specs/finder`, `preview`, `editor`, `notes-dir-selection` and the config paths match `openspec/specs/config`
- [x] 2.2 Verify the Install section states that release assets of the private repository need a signed-in GitHub account and names the three archives exactly as the `release` spec does

## 3. GoReleaser configuration

- [x] 3.1 Write `.goreleaser.yaml` (version 2) with the build, archives, checksum, changelog and release blocks from design.md and verify `go run github.com/goreleaser/goreleaser/v2@v2.18.3 check` reports no errors
- [x] 3.2 Run `go run github.com/goreleaser/goreleaser/v2@v2.18.3 release --snapshot --clean` and verify `dist/` contains exactly `clipt_*_linux_amd64.tar.gz`, `clipt_*_windows_amd64.zip` and `checksums.txt` as release assets, and that `sha256sum -c checksums.txt` passes inside `dist/`
- [x] 3.3 Extract the snapshot Linux amd64 archive into a temp dir and verify it holds `clipt` and `README.md`, that `file clipt` reports a statically linked executable and `ldd clipt` reports it is not dynamic; extract the Windows zip and verify it holds `clipt.exe` and `README.md`

## 4. Release workflow

- [x] 4.1 Write `.github/workflows/release.yml` with the `vX.Y.Z` tag trigger, `contents: write` permission and the five steps from design.md (checkout with full history, `main` ancestry check, setup-go from `go.mod`, gofmt/vet/test gate, GoReleaser action) and verify `go run github.com/rhysd/actionlint/cmd/actionlint@latest` passes on the file
- [x] 4.2 Verify the `main` check step locally by running its `git merge-base --is-ancestor` command (read-only) against the current head and a hypothetical `origin/main` equal to it, confirming exit code 0, and against an unrelated commit such as the root commit's parent, confirming a non-zero exit

## 5. Commit and host on GitHub

- [ ] 5.1 Run `gofmt -l .`, `go vet ./...` and `go test ./...` to confirm nothing regressed, then `jj commit -m "Add README, release pipeline and GitHub hosting"` and verify `jj status` shows a clean working copy and `jj log` shows the new commit as `@-`
- [ ] 5.2 Run `jj bookmark create main -r @-` and verify `jj bookmark list` shows `main` at that commit
- [ ] 5.3 Run `gh repo create llysyganicz/clipt --private --description "Minimal, keyboard-driven TUI note-taking app"` and verify `gh repo view llysyganicz/clipt --json visibility,isEmpty` reports `PRIVATE` and empty
- [ ] 5.4 Run `jj git remote add origin https://github.com/llysyganicz/clipt.git` then `jj git push --bookmark main --allow-new` and verify `gh repo view llysyganicz/clipt --json defaultBranchRef` reports `main`, `gh api repos/llysyganicz/clipt/contents/README.md` succeeds and `gh workflow list --repo llysyganicz/clipt` lists the release workflow

## 6. First release

- [ ] 6.1 With the user's go-ahead on the version number, run `jj tag set v0.1.0 -r main` and `jj git push --tag v0.1.0`, then watch the run with `gh run watch` and verify `gh run list --workflow release.yml` shows it succeeded
- [ ] 6.2 Verify `gh release view v0.1.0 --json assets,isDraft,isPrerelease` lists `clipt_0.1.0_linux_amd64.tar.gz`, `clipt_0.1.0_windows_amd64.zip` and `checksums.txt`, with draft and pre-release both false, and that the release notes list the repository's commits
- [ ] 6.3 Download the Linux amd64 asset and `checksums.txt` with `gh release download v0.1.0` into a temp dir, verify `sha256sum -c` passes for it, extract and confirm the `clipt` binary is executable and statically linked
