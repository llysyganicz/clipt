## Context

See proposal.md for motivation. What shapes the approach:

- `internal/ui/theme` already owns every color: seven hex constants, eight
  named Lip Gloss styles, two gutter styles, `EditorStyles()`,
  `InputStyles()` and a Glamour `ansi.StyleConfig` built from
  `styles.DarkStyleConfig` with six color overrides. A test rejects hex
  literals in any other package. Roughly 35 call sites read `theme.X` as
  package-level values.
- Those values are initialized at Go package init, before `main` reads the
  config. Mode models capture styles when they are constructed, which
  already happens inside `run()` after the config is loaded.
- `config.Save` re-encodes a struct that only knows `notes_dir`, so any
  other key in the file is lost when the picker rewrites it.
- The file-safety rule (spec and AGENTS.md): the config file is the only
  file clipt writes outside the notes directory.
- `.goreleaser.yaml` archives already carry `README.md` through `files:`.

## Goals / Non-Goals

**Goals:**
- Zero churn in mode packages: they keep reading the same named styles.
- The default theme reproduces today's output byte for byte.
- A theme file is simple enough to transcribe from any terminal's color
  config by hand, and complete enough that a light scheme looks right
  everywhere, including rendered markdown.
- All theme failures surface before the TUI starts, as readable errors.

**Non-Goals:**
- Hot reload or an in-app theme picker.
- Per-element overrides (e.g. "headings in blue"); the palette is the only
  knob.
- Reading other programs' theme formats directly.
- Writing any file in the config dir other than `config.toml`.

## Decisions

### D1. Apply the palette by rebuilding package-level styles

`theme.Apply(Palette)` reassigns `Normal`, `Selected`, ..., `LineNumber`,
`CursorLineNumber`, the cursor color and `Markdown` from the given palette.
`EditorStyles()` and `InputStyles()` are already functions and read the
current values. The package initializes itself with the gruvbox dark palette
so tests and any code path that never calls `Apply` behave as today.

`main.run()` calls `Apply` after the config and theme are loaded and before
`app.New`/`app.NewPicker`, so every model is constructed with the final
styles.

Alternative considered: a `Theme` struct threaded through every mode model.
Rejected: ~35 call sites and every constructor change for no user-visible
benefit; the root-model-routes-between-modes rule is unaffected either way.

### D2. Palette is a struct of 19 colors with fixed semantic roles in code

```
type Palette struct {
    Background, Foreground, Selection string
    Black, Red, Green, Yellow, Blue, Magenta, Cyan, White string
    BrightBlack, ..., BrightWhite string
}
```

The fields hold validated, lower-cased `#rrggbb` strings rather than
`color.Color` values: Lip Gloss takes them through `lipgloss.Color` and
Glamour's `StyleConfig` wants hex strings, so strings avoid a lossy
round-trip and make the "every color is a palette value" tests a plain
string comparison.

Which slot each UI element uses is hard-coded in the theme package (see the
`theme` spec, "Palette roles"), not configurable. The current gruvbox
constants map onto bright yellow / bright cyan / bright red / bright black,
so the roles are chosen to make `gruvbox_dark` identical to today.

Alternative: a seven-key semantic file (`accent`, `muted`, ...). Rejected
by the user: full palettes are what people already have in their terminal
config, and sixteen slots give the markdown renderer enough variety.

### D3. Flat clipt-native TOML, not the Alacritty schema

One key per color, snake_case names matching the ANSI names (`bright_black`
not `brightBlack`), parsed with the existing BurntSushi/toml into a struct of
strings, then validated. The user chose this over the Alacritty shape so the
format needs no knowledge of another program's config.

Validation is strict about required keys and color syntax (`^#[0-9a-fA-F]{6}$`)
and lenient about unknown keys, mirroring the config file. Errors wrap the
file path and key name.

### D4. Lookup order: config dir file, then embedded

```
  theme name (config, default "gruvbox_dark")
       |
       v
  <config dir>/clipt/themes/<name>.toml exists? --yes--> parse that file
       | no
       v
  embedded FS has <name>.toml? ------------------yes--> parse embedded
       | no
       v
  error: theme "<name>" not found (looked in <path>)
```

The two built-ins live as `internal/ui/theme/builtin/gruvbox_dark.toml` and
`gruvbox_light.toml`, embedded by the theme package with
`//go:embed builtin/*.toml` (`go:embed` cannot reach outside the package
directory, and the project keeps packages under `internal/`). The same two
files are mapped into the archive's `themes/` folder via `src`/`dst` entries
in `.goreleaser.yaml`, so the archive and the binary can never disagree.

`<config dir>` is `filepath.Dir(config.Path())`; `config` exposes a
`Dir()` helper so the theme loader does not duplicate the platform logic.

Alternative: embedded-only. Rejected: the user wants files in the config
dir. Alternative: config-dir-only, seeded by clipt on first run. Rejected:
breaks the file-safety rule.

### D5. Theme errors are startup errors

Unknown theme, unreadable file, bad TOML, missing key, bad color: `run()`
returns an error naming the file (and key), `main` prints it to stderr and
exits 1. Same path and wording style as `config.ParseError`. The TUI has no
place to show such an error before the first screen, and silently falling
back to gruvbox would hide typos.

### D6. Config carries `theme` through save

`config.file` gains `Theme string \`toml:"theme,omitempty"\``. `Load`
returns it (new return value or a small struct; the tasks pick a struct
`Config{NotesDir, Theme}` to stop the growing tuple). `Save(dir, theme)`
writes both; the picker receives the theme name from the root model, which
receives it from `main`. `omitempty` keeps a file that never had `theme`
free of it.

### D7. Markdown style is fully derived from the palette

`markdownStyle(p)` starts from `styles.DarkStyleConfig` for its structural
settings (indents, margins, prefixes) and then sets every `Color` and
`BackgroundColor` field that the dark config defines, plus the element list
in the `theme` spec, from the palette. Starting from the dark config is only
a convenience for structure; since all colors are overwritten, a light
palette renders correctly. Syntax highlighting inside fenced code blocks is
dropped (the dark config's Chroma colors are not palette values): code block
text uses `bright_cyan` uniformly. This is the one visible change under
`gruvbox_dark`: fenced blocks with a language tag used to get Glamour's
default highlighting colors.

## Risks / Trade-offs

- [Package-level mutable styles] → `Apply` is called once in `main` before
  any model exists; tests that rely on defaults keep working because the
  package self-initializes to gruvbox dark. Document "call before building
  models" on `Apply`.
- [Glamour dark config has colors we miss] → a test iterates the resulting
  `StyleConfig` with reflection and asserts every non-nil `Color` /
  `BackgroundColor` is one of the palette's hex strings. This also serves
  as the "no color outside the palette" scenario for markdown.
- [Default drift] → a test asserts that parsing the embedded
  `gruvbox_dark.toml` yields exactly the previous hard-coded constants
  (`#282828`, `#504945` selection, `#ebdbb2`, `#fabd2f`, `#8ec07c`,
  `#fb4934`, `#928374`).
- [Bright black as selection fallback is unreadable on some schemes] →
  `selection` is an optional key, documented in README; both built-ins set
  it explicitly.
- [Windows path in README/spec] → the spec names the archive paths with a
  backslash for the zip scenario only; the loader uses `filepath.Join`.
- [Config-dir theme reads on every start] → one `os.Stat` plus a small file
  read; negligible.

## Migration Plan

No migration. Existing config files have no `theme` key and keep rendering
gruvbox dark. Rolling back to a previous binary leaves a `theme` key in the
file, which that binary ignores as an unknown key.
