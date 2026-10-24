## Why

clipt is hard-wired to the gruvbox dark palette, which clashes with any
terminal that uses a different scheme and is unusable on a light background.
The theme package already isolates every color, so letting the user pick a
palette in the config file is the natural next step and the one the project
context has deferred until now.

## What Changes

- A new optional `theme` key in `config.toml` names the color theme; absent
  or empty it means `gruvbox_dark`, so existing installs look exactly as
  before.
- A theme is a flat TOML file with a full terminal palette: `background`,
  `foreground`, the sixteen standard ANSI color names (`black` ... `white`,
  `bright_black` ... `bright_white`) and an optional `selection` background.
- Themes are looked up by name in `<config dir>/clipt/themes/<name>.toml`
  first, then among the two themes built into the binary (`gruvbox_dark`,
  `gruvbox_light`). A user file with a built-in name shadows it.
- An unknown theme name, an unreadable or malformed theme file, a bad color
  value or a missing key is a startup error naming the file, the same way a
  malformed config is today.
- Every UI style and the rendered-markdown style is derived from the theme
  palette at startup; mode packages keep using the named styles unchanged.
- Rewriting the config file (notes directory picker) preserves the `theme`
  key instead of dropping it.
- Release archives ship the built-in theme files in a `themes/` folder so a
  user can copy one into the config dir as a starting point.
- README documents the key, the file format, the lookup order and the two
  built-in themes.

Non-goals: downloading themes over the network, writing theme files from
clipt, an in-app theme picker, per-element style overrides beyond the
palette, light/dark auto-detection from the terminal.

## Capabilities

### New Capabilities
- `theme`: theme file format, built-in themes, lookup by name from the
  config directory and the binary, error handling, and how palette slots map
  onto the UI and markdown styles.

### Modified Capabilities
- `config`: the file shape gains an optional `theme` string key; the
  "unknown key" scenario that used `theme` as its example moves to another
  key; writing the config preserves `theme`.
- `release`: the archives additionally contain a `themes/` folder with the
  built-in theme files.

## Impact

- `internal/config`: `theme` field on the file struct, returned by `Load` and
  written by `Save`.
- `internal/ui/theme`: package vars become derived from a palette through an
  `Apply` step; new theme file parsing, embedded built-ins (`go:embed`),
  lookup in the config dir, validation; the Glamour style maps every colored
  element onto palette slots.
- `cmd/clipt`: loads the theme after the config and before building the
  program; new startup error path.
- `internal/ui/picker` and `internal/ui/app`: pass the loaded theme name
  through to `Save` so a notes-dir change keeps it.
- New `internal/ui/theme/builtin/` folder with the two theme files,
  embedded in the binary and mapped to `themes/` in the `.goreleaser.yaml`
  archives.
- `README.md`: new "Themes" section.
- No new third-party dependencies: TOML via the existing BurntSushi/toml,
  colors via Lip Gloss.
