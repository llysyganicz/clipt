# clipt

clipt is a minimal, keyboard-driven TUI note-taking app. Notes are plain
markdown files in a single flat directory: the file name is the note name,
there are no folders, tags or sync. You fuzzy-find a note by name, read it
rendered, and edit it in an insert-only text area that saves when you leave.

## Modes

clipt has three modes plus a small directory picker. Every mode shows its
keys in the status bar at the bottom. `ctrl+c` quits from anywhere.

### Finder

The home screen. Lists the notes in the notes directory by modification time,
newest first, hiding the note that is currently open. Typing filters the list
with a case-insensitive subsequence match against the note names, and when
no note has exactly the typed name a last row offers to create it.

| Key                 | Action                                                        |
|---------------------|---------------------------------------------------------------|
| any printable key   | Edit the query                                                |
| `Up` / `ctrl+p`     | Move the selection up                                         |
| `Down` / `ctrl+n`   | Move the selection down                                       |
| `Enter`             | Open the selected note in preview, or create `<query>.md` on the create row and open it in edit mode |
| `Esc`               | Back to the open note's preview, or quit when no note is open |
| `ctrl+o`            | Change the notes directory                                    |
| `ctrl+c`            | Quit                                                          |

### Preview

Shows the open note as rendered markdown, word-wrapped to the terminal width.

| Key              | Action                                                  |
|------------------|---------------------------------------------------------|
| `Up` / `Down`    | Scroll one line                                         |
| `PgUp` / `PgDn`  | Scroll one page                                         |
| `e`              | Edit the note                                           |
| `r`              | Rename the note (`Enter` confirms, `Esc` cancels)       |
| `d`              | Delete the note (`y` confirms, `n` or `Esc` cancels)    |
| `Esc`            | Back to the finder                                      |
| `ctrl+c`         | Quit                                                    |

A rename is refused when a note with the new name already exists. Deleting
returns to the finder.

### Edit

The note's raw markdown in a text area with line numbers. There is no modal
editing: every key inserts or moves the cursor, like a plain text box. There
is no save command; leaving edit mode writes the note if it changed.

| Key                                   | Action                                   |
|---------------------------------------|------------------------------------------|
| printable keys, `Enter`, `Backspace`, `Delete` | Edit the text                   |
| arrow keys, `Home`, `End`             | Move the cursor                          |
| `Esc`                                 | Save (if changed) and back to preview    |
| `ctrl+c`                              | Quit without saving                      |

### Notes directory picker

Opens on the first run, when the configured directory no longer exists, and
on `ctrl+o` from the finder. It is a single path input prefilled with
`~/notes` on the first run. A leading `~` is expanded to your home directory.

| Key       | Action                                                                 |
|-----------|------------------------------------------------------------------------|
| `Enter`   | Accept the path; a missing directory is created after a `y`/`n` prompt |
| `Esc`     | Quit on the first run, or back to the finder when opened via `ctrl+o`  |
| `ctrl+c`  | Quit                                                                   |

## Notes directory and config

The only setting is the notes directory. It is stored as a single key,
`notes_dir`, in a TOML file named `clipt/config.toml` under your platform's
config directory:

| Platform | Config file                                                        |
|----------|--------------------------------------------------------------------|
| Linux    | `$XDG_CONFIG_HOME/clipt/config.toml` (default `~/.config/clipt/config.toml`) |
| Windows  | `%AppData%\clipt\config.toml`                                      |
| macOS    | `~/Library/Application Support/clipt/config.toml`                  |

```toml
notes_dir = "/home/user/notes"
```

clipt writes nothing outside the notes directory except this file. The
directory is chosen in the app; you never need to edit the file by hand.

## Install

Builds are published as [GitHub Releases](https://github.com/llysyganicz/clipt/releases).
The repository is private, so you need to be signed in to GitHub with an
account that has access to it to download the assets.

Each release carries:

| Asset                             | Contents                 |
|-----------------------------------|--------------------------|
| `clipt_X.Y.Z_linux_amd64.tar.gz`  | `clipt` and `README.md`     |
| `clipt_X.Y.Z_windows_amd64.zip`   | `clipt.exe` and `README.md` |
| `checksums.txt`                   | SHA-256 sums of both archives |

Download the archive for your platform, optionally check it against
`checksums.txt`, extract it and put the `clipt` binary somewhere on your
`PATH`. The binaries are statically linked and need no runtime.

```sh
tar -xzf clipt_X.Y.Z_linux_amd64.tar.gz
sha256sum -c --ignore-missing checksums.txt
mv clipt ~/.local/bin/
```

With the GitHub CLI signed in, `gh release download vX.Y.Z --repo llysyganicz/clipt`
fetches the assets of a release in one go.

## Build from source

Requires the Go version declared in `go.mod` (or newer).

```sh
go build ./cmd/clipt
./clipt
```

## Development

- Version control is [Jujutsu](https://jj-vcs.github.io/jj/) colocated with
  git: `jj commit -m "..."`, `jj describe`, `jj split`. Do not run git
  commands that change the repository.
- Behaviour changes follow the [OpenSpec](https://github.com/Fission-AI/OpenSpec)
  workflow: propose a change, apply it, archive it. The specs under
  `openspec/specs/` are the source of truth for what the app does.
- Before committing, these must pass:

```sh
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
```

Tests that touch the file system use a temp dir, never real notes.

## Releasing

A release is one tag push away. Tags must have the form `vX.Y.Z` and must
point to a commit on `main`:

```sh
jj tag set vX.Y.Z -r main
jj git push --tag vX.Y.Z
```

The release workflow then checks that the tagged commit is on `main`, runs
`gofmt`, `go vet` and `go test` as a gate, builds the Linux and Windows
`amd64` archives with GoReleaser and publishes a GitHub Release named after
the tag, with release notes listing the commits since the previous version
tag. A tag that is not on `main` or that fails the gate publishes nothing.
To retry a failed release, delete the tag and push it again; the existing
release is replaced.

Run the pipeline locally without publishing anything:

```sh
go run github.com/goreleaser/goreleaser/v2@v2.18.3 release --snapshot --clean
```

The archives land in `dist/`, which is ignored by version control.
