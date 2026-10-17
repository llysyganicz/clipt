## Context

See proposal.md - Why. Facts about the current code and the v2 libraries
that shape the approach (verified against the v2 module sources: Bubble Tea
v2.1.0, Bubbles v2.2.1, Lip Gloss v2.0.6, Glamour v2.0.1):

- The v2 modules declare `charm.land/<lib>/v2` as their module path. The
  `github.com/charmbracelet/<lib>/v2` paths resolve to the same code but a
  `go get` on them fails the module-path check, so `charm.land` is the only
  importable path.
- Every mode package (`finder`, `preview`, `editor`, `picker`, `tmpl`) and
  the root `app` implement v1's `tea.Model`: `Update` returns `tea.Model` and
  `View` returns `string`. The root type-asserts each sub-model back after
  every `Update` (about twenty assertions), and the tests do the same.
- In v2, `tea.Model.View` returns `tea.View` (content plus declarative
  terminal state such as `AltScreen`); `tea.WithAltScreen()` is gone.
  `tea.KeyMsg` is now an interface and presses arrive as `tea.KeyPressMsg`
  with `Code rune`, `Text string` and `Mod`; `String()` returns the text of a
  printable key and otherwise a name such as `esc`, `enter`, `shift+tab`,
  `ctrl+q`, `pgup`. The `KeyCtrlQ`-style constants are gone.
- Bubbles v2: `textarea.DefaultKeyMap()` is a function; `FocusedStyle` and
  `BlurredStyle` become one `Styles{Focused, Blurred StyleState; Cursor}` set
  through `SetStyles`; `SetCursor` is renamed `SetCursorColumn`;
  `SetPromptFunc` takes `func(PromptInfo) string` where `PromptInfo.LineNumber`
  is the display row, the same value v1 passed as an `int`. `textinput` moves
  `PromptStyle` and friends into `Styles` set through `SetStyles`.
  `viewport.New` takes options and `Width`/`Height` become setters. Both
  inputs keep a virtual (drawn) cursor by default and still export `Blink`.
  The text area's unexported `wrap` is byte-identical to the copy in
  `internal/ui/editor/wrap.go`, and `Update` still repositions the viewport
  on every message, so the editor's `repositionMsg` trick keeps working.
- Bubbles v2's text area default key map adds page up/down and a text
  selection set (`shift+arrows`, `ctrl+g` select all, `ctrl+shift+c` copy)
  that v1 did not have. `ctrl+t` is still bound to transpose by default.
- Bubbles v2 default styles color prompts, blurred text and the cursor with
  ANSI palette entries (`7`, `240`), where v1 left them unset.
- Lip Gloss v2's `Style.Render` always emits true-color escape codes; the
  Bubble Tea renderer downsamples for the terminal. Lip Gloss v1 detected the
  profile itself and rendered plain text under `go test`, which several
  tests rely on (`strings.HasPrefix(row, "  9 ")`, `strings.Index` for the
  text column, `HasPrefix(lines[0], "template> ")`). Two test files already
  carry a regex stripper for SGR codes. Glamour v2 is likewise pure and emits
  colors unconditionally.
- `textarea.SetWidth` in v2 reserves line-number columns only while
  `ShowLineNumbers` is on; v1 always reserved them. The editor turns them off,
  so the text column becomes a few cells wider at the same terminal width.
- `github.com/charmbracelet/x/ansi` is already an indirect dependency and
  offers `ansi.Strip`.

## Goals / Non-Goals

**Goals:**

- A mechanical, behavior-preserving port: every current key, layout, color
  and file operation works as before, and every existing test keeps its
  intent.
- Less glue, not more: drop the sub-model type assertions the v1 interface
  forced on the root and the tests.
- One place for test-side knowledge of v2 key names and ANSI stripping.

**Non-Goals:**

- Adopting any new v2 capability (real cursor, selection, soft wrap, light
  theme). Where v2 turns one on by default, this design turns it off to keep
  behavior identical.
- Refactoring beyond what the API change requires.

## Decisions

### Only the root model implements `tea.Model`

The root `app.Model` keeps the `tea.Model` trio and its `View` builds a
`tea.View` from the active mode's string with `AltScreen = true`; `main.go`
drops `tea.WithAltScreen()`. Each mode model keeps `Init`, `Update` and
`View`, but `Update` returns its own `Model` and `View` keeps returning
`string`. The root assigns the result directly and the type assertions go
away, in `app.go` and in the tests.

This is how Bubbles' own components are shaped in v2, and it is the smaller
change: the alternative, making every mode return `tea.View` so it still
satisfies `tea.Model`, would have the root unwrap `.Content` from each child
and re-wrap it, and would keep all the assertions. The project rule "model
each mode as its own Bubble Tea model" is still met: each mode is a
self-contained Elm-style model with the same three methods; it just does not
need to be addressable through the interface.

### Keys are matched by name

Each `case tea.KeyMsg:` becomes `case tea.KeyPressMsg:` and the switches on
`msg.Type` become switches on `msg.String()` using the names Bubble Tea
produces: `esc`, `enter`, `up`, `down`, `left`, `right`, `tab`, `shift+tab`,
`home`, `end`, `pgup`, `pgdown`, `backspace`, `ctrl+c`, `ctrl+q`, `ctrl+t`,
`ctrl+o`, `ctrl+p`, `ctrl+n`. Printable keys yield their text, so the
existing `msg.String() == "e"` / `"y"` / `"N"` checks are unchanged.

Alternative considered: matching on `Code` and `Mod` fields
(`msg.Code == 'q' && msg.Mod == tea.ModCtrl`). It is what the upgrade guide
calls "more foolproof", but it doubles every control-key case and diverges
from the string form Bubbles' own `key.Matches` uses. The names are stable
API (they are what `key.WithKeys` takes), and a test per mode already covers
each binding.

### Text area adaptation preserves v1 behavior

- Key map: start from `textarea.DefaultKeyMap()`, disable
  `TransposeCharacterBackward` as today, and also disable the six selection
  bindings plus `SelectAll` and `CopySelection`, so `shift+arrows`, `ctrl+g`
  and `ctrl+shift+c` do nothing, as in v1. Page up/down stay enabled: they
  only move the cursor and do not change any documented behavior. `Paste`
  (`ctrl+v`) existed in v1 and stays.
- Styles: `theme.EditorStyles()` returns one `textarea.Styles` built from the
  same eight styles it sets today (`Selection` left empty, since selection is
  disabled), with `Cursor` set to a blinking block in the theme foreground so
  the cursor color is owned by the theme rather than by Bubbles' ANSI `7`
  default.
- `SetCursor` becomes `SetCursorColumn`; the prompt callbacks take
  `textarea.PromptInfo` and read `LineNumber` where they read the `int`.
- The copied `wrap` stays; its comment is updated to name v2.2.1 as the
  source. The existing equivalence test keeps guarding the copy.
- Resize: v2's `CursorUp` scrolls the viewport on every call (v1 only did so
  in `Update`), so the editor's reflow applies the new size right after
  reloading the buffer and before walking the cursor back to its line.
  Otherwise the walk scrolls against the old height and the grown terminal
  shows blank rows below the content. Found by the existing resize test.

### Text inputs take their styles from the theme

`theme.InputStyles()` returns a `textinput.Styles` whose focused and blurred
states both use `theme.Prompt` for the prompt, an unstyled `Text`, and
`theme.Hint` for placeholder and suggestion, with the same cursor style as
the editor. Every `textinput.New()` call site (`finder`, `picker`, `preview`
rename, `tmpl` picker and form) replaces `ti.PromptStyle = theme.Prompt` with
`ti.SetStyles(theme.InputStyles())`. Without this, v2's default styles would
dim the blurred form rows in the template form and recolor the cursor.

### Viewport and Lip Gloss and Glamour are path and signature changes only

- `viewport.New(viewport.WithWidth(w), viewport.WithHeight(h))`; `render`
  uses `SetWidth`/`SetHeight`. The key map literal is unchanged; the new
  `Left`/`Right`/half-page bindings stay zero so they never match, and
  `SoftWrap` stays off because Glamour already wraps to the width.
- Lip Gloss: `lipgloss.Color(hex)` now returns `color.Color`; the style
  calls, `Width`, `PlaceHorizontal` and `Right` are unchanged.
  `GetBackground()` still returns `NoColor{}` when unset, so the theme test
  comparison holds.
- Glamour: import paths only. `WithStyles(theme.Markdown)`, `WithWordWrap`,
  `ansi.StyleConfig` and `styles.DarkStyleConfig` are unchanged, including
  the `*string` color fields.

### A shared `uitest` helper replaces per-package test glue

A new `internal/ui/uitest` package, imported only by tests, provides:

- `Key(name string) tea.KeyPressMsg`: parses the same names the production
  code matches on (`ctrl+q`, `shift+tab`, `esc`, ...) into `Code`/`Mod`, and
  turns a single printable character into `{Code: r, Text: string(r)}`.
- `Type(s string) []tea.KeyPressMsg`: one press for a key name, otherwise
  one press per character, which is what the old per-package parsers did
  implicitly when handed a word to type.
- `Strip(s string) string`: `ansi.Strip` from `charmbracelet/x/ansi`, which
  removes CSI and OSC sequences, not only SGR.

The four `k`/`key` parsers in `finder`, `tmpl`, `picker` and the
`tea.KeyType`-based `press` helpers in `editor`, `preview`, `app` all route
through `uitest.Key`; the two regex strippers are deleted. Every assertion
that reads raw view text (gutter prefixes and `textCol` in the editor tests,
the prompt prefixes in the finder and template tests, any `Contains` on
styled text) passes the view through `uitest.Strip` first. `x/ansi` becomes a
direct dependency.

Alternatives considered: converting each helper in place (six copies of the
same two functions, and every future key name change touches six files);
forcing a no-color profile in tests (Lip Gloss v2 has no global profile to
set, the downsampling lives in the renderer that tests never run).

### Dependency pinning

`go.mod` pins the four modules at the versions above rather than `@latest`,
so the plan reproduces. `go mod tidy` removes the v1 modules and their
`termenv`/`muesli` tail and pulls `ultraviolet` at the newest pseudo-version
the four require (Bubble Tea's). `mattn/go-runewidth` and `rivo/uniseg` stay
direct for `wrap.go`.

## Risks / Trade-offs

- [Rendering tests fail on escape codes] → Every raw-text assertion goes
  through `uitest.Strip`; width checks already use `lipgloss.Width`, which
  ignores escapes.
- [Text column is wider than in v1 at the same terminal width, moving wrap
  points] → Gutter tests compare columns across rows, not absolute columns,
  and the wrap test asks the text area for its width. No test hard-codes a
  wrap column.
- [Glamour v2 wraps differently, especially CJK] → Preview tests only bound
  line widths; a visibly different wrap of ordinary text is a Glamour bug to
  report, not something to work around here.
- [Hidden behavior change from new defaults] → Selection and copy bindings
  are disabled; input and cursor styles are set explicitly by the theme. The
  only accepted additions are page up/down in the editor.
- [Esc and alt-key disambiguation differs in the new input parser] → Same
  legacy-terminal rules as v1 in practice; the manual pass presses Esc in
  every mode.
- [Mismatched `ultraviolet` pseudo-versions between the four modules] →
  Minimal version selection takes the newest; `go build ./...` and
  `go vet ./...` catch an incompatible pick immediately.
- [The `uitest` package is production-visible code with no production
  callers] → It lives under `internal/ui/` next to what it serves and its doc
  comment says it is for tests; a `go vet` pass covers it like any package.

## Migration Plan

1. Land everything as one jj change (dependency bump, code, tests, helper),
   gated by `gofmt -l .`, `go vet ./...` and `go test ./...`.
2. Manual pass in a Linux terminal before tagging: finder (type, arrows,
   `ctrl+p`/`ctrl+n`, create row, `ctrl+t`, `ctrl+o`, Esc, `ctrl+q`),
   preview (scroll with arrows and page keys, `e`, `r`, `d` with `y`/`n`,
   Esc), editor (gutter across line 10, Tab/Shift+Tab, `ctrl+t` insert,
   resize, Esc save, `ctrl+q`), notes-dir picker (`~` path, create prompt),
   and that `shift+arrows` do nothing in the editor.
3. Rollback: revert the change; no data, config or on-disk format is touched.
