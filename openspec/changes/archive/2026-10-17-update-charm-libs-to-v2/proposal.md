## Why

clipt still builds on the Charm v1 libraries (Bubble Tea 1.3.10, Bubbles
1.0.0, a pre-release Lip Gloss 1.1.1 snapshot and Glamour 1.0.0). Charm v2 is
the current stable major for all four, published under the `charm.land`
module paths; v1 only receives maintenance. The editor-ux-fixes change
deliberately deferred this migration so it could land on its own, and the UI
is small enough today that moving every package in one pass is still cheap.
Later work (a real terminal cursor, light/dark detection, viewport soft wrap)
is only available on v2.

## What Changes

- Replace the four Charm dependencies with their v2 majors at the canonical
  import paths: `charm.land/bubbletea/v2` v2.1.0, `charm.land/bubbles/v2`
  v2.2.1, `charm.land/lipgloss/v2` v2.0.6 and `charm.land/glamour/v2` v2.0.1
  (newest stable releases at the time of writing).
- Adapt every UI package to the v2 APIs: the root model's `View` returns a
  `tea.View` that declares the alternate screen (the `WithAltScreen` program
  option is gone); key presses arrive as `tea.KeyPressMsg` and are matched by
  name (`"esc"`, `"ctrl+q"`, ...) instead of `KeyType` constants; mode models
  return their own type from `Update` since they no longer need to satisfy
  `tea.Model`; Bubbles styles move into `Styles` structs, the viewport takes
  options and setters, and the text area's cursor setter and prompt callback
  change signature.
- Adapt the tests: key messages are built as `tea.KeyPressMsg`, and view
  assertions strip ANSI escape codes, because Lip Gloss v2 always emits them
  (v1 rendered plain text when no terminal was attached, which the current
  tests rely on).
- No user-visible change: same keys, same layout, same colors, same files on
  disk. The virtual (drawn) cursor is kept, so no cursor plumbing is added.

Non-goals:

- Using the real terminal cursor, keyboard enhancements, mouse support or any
  other new v2 feature.
- Light-background detection or a light theme (the theme stays gruvbox dark).
- Replacing the copied text-area `wrap` function: v2 still does not export
  it and the copy is byte-identical to v2.2.1.
- Upgrading non-Charm dependencies beyond what `go mod tidy` requires.
- Changing README content: it does not mention the libraries.
- Windows and macOS verification (remains a manual pass after v1).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. No requirement changes: the migration keeps every documented behavior,
so this change sets `skip_specs: true`.

## Impact

- `go.mod` / `go.sum`: the four Charm requirements change module path and
  major; the indirect tree shifts to `charmbracelet/ultraviolet`,
  `colorprofile` and `x/ansi`, dropping `termenv`, the `muesli/*` helpers and
  others as `go mod tidy` dictates. `mattn/go-runewidth` and `rivo/uniseg`
  stay direct (used by the copied `wrap`).
- Code: `cmd/clipt/main.go` and every file under `internal/ui/` that imports
  a Charm package (`app`, `editor`, `finder`, `picker`, `preview`, `theme`,
  `tmpl`). Files in `internal/config`, `internal/notes` and
  `internal/templates` are untouched.
- Tests: every `*_test.go` under `internal/ui/`; a small shared test helper
  package is added for key construction and ANSI stripping.
- Release pipeline and CI: unchanged. The same `gofmt`, `go vet` and
  `go test` gate applies.
- Risk: rendering differences from the new renderer and from Glamour v2's
  improved wrapping; covered by the existing view tests plus a manual pass in
  a Linux terminal before release.
