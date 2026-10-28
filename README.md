# clipt

clipt is a minimal, keyboard-driven TUI note-taking app. Notes are plain
markdown files in a single flat directory: the file name is the note name,
there are no folders or sync. Notes can carry tags in YAML front matter. You
fuzzy-find a note by name or tag, read it rendered, and edit it in a plain
text area that saves when you leave. Templates with small forms
can be inserted into a note while editing.

## Modes

clipt has four modes (finder, preview, edit and template insert) plus a
small directory picker. Every mode shows its
keys in the status bar at the bottom. `ctrl+q` quits from anywhere, saving
the note first when you are editing one; `ctrl+c` quits without saving.

### Finder

The home screen. Lists the notes in the notes directory by modification time,
newest first, hiding the note that is currently open, with each note's tags
shown dimmed after its name. Typing filters the list with a case-insensitive
subsequence match against the note names, and when no note has exactly the
typed name a last row offers to create it.

Words in the query that start with `#` filter by tag instead: `#wo` keeps
the notes with a tag starting with `wo`, case-insensitively, and several
`#` words must all match. The rest of the query is matched against the name,
so `#work sta` lists the `work` notes whose name matches `sta`. Creating a
note from such a query writes the tags into it: `#work standup` creates
`standup.md` holding `tags: [work]` and opens it with the cursor below the
front matter.

| Key                 | Action                                                        |
|---------------------|---------------------------------------------------------------|
| any printable key   | Edit the query                                                |
| `Up` / `ctrl+p`     | Move the selection up                                         |
| `Down` / `ctrl+n`   | Move the selection down                                       |
| `Enter`             | Open the selected note in preview, or create `<query>.md` on the create row and open it in edit mode |
| `Esc`               | Back to the open note's preview, or quit when no note is open |
| `ctrl+t`            | Switch to the template list (see [Templates](#templates))     |
| `ctrl+o`            | Change the notes directory                                    |
| `ctrl+q`            | Quit                                                          |
| `ctrl+c`            | Quit                                                          |

### Preview

Shows the open note as rendered markdown, word-wrapped to the terminal width.
The front matter is not rendered; when the note has tags they are shown in a
header above the content that stays in place while scrolling. Invalid front
matter is reported in that header instead.

| Key              | Action                                                  |
|------------------|---------------------------------------------------------|
| `Up` / `Down`    | Scroll one line                                         |
| `PgUp` / `PgDn`  | Scroll one page                                         |
| `e`              | Edit the note                                           |
| `r`              | Rename the note (`Enter` confirms, `Esc` cancels)       |
| `d`              | Delete the note (`y` confirms, `n` or `Esc` cancels)    |
| `Esc`            | Back to the finder                                      |
| `ctrl+q`         | Quit                                                    |
| `ctrl+c`         | Quit                                                    |

A rename is refused when a note with the new name already exists. Deleting
returns to the finder.

### Edit

The note's raw markdown in a text area with line numbers. There is no modal
editing: every key inserts, moves the cursor or selects, like a plain text
box. There is no save command; leaving edit mode writes the note if it
changed.

| Key                                   | Action                                   |
|---------------------------------------|------------------------------------------|
| printable keys, `Enter`, `Backspace`, `Delete` | Edit the text (replacing or removing the selection) |
| arrow keys, `Home`, `End`             | Move the cursor (clearing the selection) |
| `Shift+Left` / `Shift+Right`          | Select by character                      |
| `Shift+Up` / `Shift+Down`             | Select by line                           |
| `Ctrl+Shift+Left` / `Ctrl+Shift+Right` | Select by word                          |
| `ctrl+g`                              | Select all                               |
| `ctrl+shift+c`                        | Copy the selection to the system clipboard |
| `Tab`                                 | Insert two spaces (replacing the selection) |
| `Shift+Tab`                           | Remove up to two leading spaces          |
| `ctrl+t`                              | Insert a template at the cursor (replacing the selection) |
| `Esc`                                 | Save (if changed) and back to preview    |
| `ctrl+q`                              | Save (if changed) and quit               |
| `ctrl+c`                              | Quit without saving                      |

Copying uses the system clipboard: on Linux one of `xclip`, `xsel` or
`wl-copy` must be installed, Windows and macOS need nothing. When the
clipboard cannot be written the error is shown above the status bar. Some
terminals handle `ctrl+shift+c` themselves or report it as `ctrl+c`, which
quits without saving; the status bar shows the copy key while a selection
exists.

### Template insert

Opened with `ctrl+t` while editing a note. A picker lists the templates,
filtered like the finder (name and `#tag` words) and showing their tags.
Picking a template without fields inserts its body at once; otherwise a form
opens with one row per field. The body is inserted at the cursor and the
template's tags are added to the note (see [Tags](#tags)).

| Key                         | Action                                               |
|-----------------------------|------------------------------------------------------|
| any printable key           | Edit the picker query or the focused form value      |
| `Up` / `Down`, `ctrl+p` / `ctrl+n` | Move the picker selection                     |
| `Enter`                     | Pick the template; in the form, next row or submit on the last row |
| `Tab` / `Shift+Tab`, `Up` / `Down` | Move between form rows                        |
| `Left` / `Right`            | Change an option field                               |
| `Esc`                       | Back to edit mode with the note unchanged            |
| `ctrl+q`                    | Save the note (if changed) and quit                  |
| `ctrl+c`                    | Quit without saving                                  |

### Notes directory picker

Opens on the first run, when the configured directory no longer exists, and
on `ctrl+o` from the finder. It is a single path input prefilled with
`~/notes` on the first run. A leading `~` is expanded to your home directory.

| Key       | Action                                                                 |
|-----------|------------------------------------------------------------------------|
| `Enter`   | Accept the path; a missing directory is created after a `y`/`n` prompt |
| `Esc`     | Quit on the first run, or back to the finder when opened via `ctrl+o`  |
| `ctrl+q`  | Quit                                                                   |
| `ctrl+c`  | Quit                                                                   |

## Tags

A note may start with YAML front matter: the first line is exactly `---` and
the block ends at the next `---` line. Its `tags` key lists the note's tags,
in either YAML list form:

```markdown
---
tags: [work, meeting]
---
# Standup
```

```yaml
tags:
  - work
  - meeting
```

Tags keep the order and case you write; blank entries and repeats (compared
case-insensitively) are ignored, and other front matter keys are left alone.
Front matter that is not closed, is not valid YAML or whose `tags` is not a
list counts as invalid: the note is still listed, just without tags. clipt
never rewrites front matter on its own; the only change it makes is adding a
template's tags when the template is inserted, creating the front matter when
the note has none. Tags are shown as `#work #meeting`.

## Templates

Templates are markdown files in the `.templates` directory inside the notes
directory; they are never listed as notes. In the finder, `ctrl+t` switches to
the template list: `Enter` opens a template in edit mode, the create row
creates a new template from a starter skeleton (with the query's `#tags`
filled in) and `Esc` returns to the notes.

A template may start with front matter declaring `tags` and a `form`:

```markdown
---
tags: [meeting]
form:
  - name: topic
    label: Topic
    type: text
  - name: prio
    type: option
    options: [low, high]
---
## {{topic}} ({{prio}})

- agenda
```

Each field has a `name`, a `type` and an optional `label` (defaults to the
name). Types are `text` (starts empty), `number` (starts at `0`, must be a
decimal number) and `option` (starts at the first entry of `options`). Every
`{{name}}` in the body is replaced by the field's value; placeholders that
name no field are kept as written. Template tags are taken literally.

## Notes directory and config

Settings live in a TOML file named `clipt/config.toml` under your platform's
config directory:

| Platform | Config file                                                        |
|----------|--------------------------------------------------------------------|
| Linux    | `$XDG_CONFIG_HOME/clipt/config.toml` (default `~/.config/clipt/config.toml`) |
| Windows  | `%AppData%\clipt\config.toml`                                      |
| macOS (untested) | `~/Library/Application Support/clipt/config.toml`          |

| Key         | Required | Meaning                                                   |
|-------------|----------|-----------------------------------------------------------|
| `notes_dir` | yes      | Absolute path of the notes directory                      |
| `theme`     | no       | Name of the color theme (see [Themes](#themes)); default `gruvbox_dark` |

```toml
notes_dir = "/home/user/notes"
theme = "gruvbox_light"
```

Unknown keys are ignored. clipt writes nothing outside the notes directory
except this file. The notes directory is chosen in the app, which rewrites the
file keeping the `theme` key; the theme is the one setting you edit by hand.

## Themes

clipt draws every screen, including rendered markdown, from one terminal
color palette. The palette comes from the theme named by the `theme` key;
without the key clipt uses `gruvbox_dark` and looks exactly as it did before
themes existed. Two themes are built into the binary:

| Theme           | Palette                 |
|-----------------|-------------------------|
| `gruvbox_dark`  | gruvbox dark (default)  |
| `gruvbox_light` | gruvbox light           |

A theme named `<name>` is looked up in this order:

1. `<config dir>/clipt/themes/<name>.toml`, next to `config.toml` (for example
   `~/.config/clipt/themes/<name>.toml` on Linux);
2. the built-in theme of that name.

A file in the themes directory wins, so saving your own `gruvbox_dark.toml`
there replaces the built-in one. clipt never writes to that directory; the
release archives ship both built-in themes in a `themes/` folder, so copy one
of them in as a starting point.

A theme file is flat TOML with one `"#rrggbb"` color per key: `background`,
`foreground`, the sixteen standard ANSI colors and an optional `selection`.
This is the built-in `gruvbox_dark`:

```toml
background = "#282828"
foreground = "#ebdbb2"
selection = "#504945"

black = "#282828"
red = "#cc241d"
green = "#98971a"
yellow = "#d79921"
blue = "#458588"
magenta = "#b16286"
cyan = "#689d6a"
white = "#a89984"

bright_black = "#928374"
bright_red = "#fb4934"
bright_green = "#b8bb26"
bright_yellow = "#fabd2f"
bright_blue = "#83a598"
bright_magenta = "#d3869b"
bright_cyan = "#8ec07c"
bright_white = "#ebdbb2"
```

Every key except `selection` is required. `selection` is the background of
the selected text in edit mode and defaults to `bright_black` when omitted.
Unknown keys are ignored, so a file with extra entries (say a `cursor` color
from your terminal's config) loads fine.

While clipt runs the whole terminal screen is on the theme's `background`
and `foreground` is the default text color, in every mode. Nothing else sets
a background: rows are drawn straight on the screen, and only the editor's
selected text sits on `selection`. How the colors get there depends on the
platform:

- On Linux and macOS clipt sets the terminal's own background and foreground
  colors and restores them when it exits. A terminal or multiplexer that
  ignores the request shows the palette's text colors on its own background
  instead.
- On Windows clipt paints every cell of its screen itself and never changes
  the terminal's colors, because the console that sits between clipt and the
  terminal applies the change but drops the request to restore the original
  colors on exit. A column or row the terminal keeps for itself (one that no
  program can draw into) shows the terminal's own background rather than the
  theme's.

The slots are used like this: ordinary text, including what you type into
the finder and the directory prompt, in `foreground`; the selected row,
prompts and the cursor's line number in `bright_yellow`; the create row and
tags in `bright_cyan`; errors in `bright_red`; hints, dimmed tags and line
numbers in `bright_black`. In rendered markdown, headings and table headers
use `bright_yellow`, links and code `bright_cyan`, quotes and rules
`bright_black`, list markers `bright_blue` and ticked task boxes
`bright_green`. Fenced code blocks are not syntax highlighted.

A theme that cannot be used stops clipt before it starts, with a message on
stderr and exit status 1: an unknown name reports the path that was looked
for, and a file that is not valid TOML, lacks a key or holds a value that is
not a `#rrggbb` color reports the file and, where there is one, the key.
Nothing is written in that case.

## Install

Builds are published as [GitHub Releases](https://github.com/llysyganicz/clipt/releases).
The repository is private, so you need to be signed in to GitHub with an
account that has access to it to download the assets.

Each release carries:

| Asset                             | Contents                                        |
|-----------------------------------|-------------------------------------------------|
| `clipt_X.Y.Z_linux_amd64.tar.gz`  | `clipt`, `README.md` and the built-in `themes/` |
| `clipt_X.Y.Z_windows_amd64.zip`   | `clipt.exe`, `README.md` and the built-in `themes/` |
| `checksums.txt`                   | SHA-256 sums of both archives                   |

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

Requires the Go version declared in `go.mod` (or newer). This is also the
way to run clipt on macOS, which has no release build. clipt only uses
cross-platform standard library APIs, so it should work there, but macOS is
untested.

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
