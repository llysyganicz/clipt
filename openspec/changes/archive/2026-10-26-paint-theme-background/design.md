## Context

See proposal.md - Why. Today every style in `internal/ui/theme` is
foreground-only except `Normal`, which also sets `background` and is used for
unselected finder and template picker rows, so the palette background shows
up only there. The root model's `View` already returns a `tea.View` with
`AltScreen` set. Bubble Tea v2 views also carry optional `BackgroundColor`
and `ForegroundColor` fields; the renderer sets the terminal's default colors
from them while the program runs and resets both when the program stops.
Typed text in the single-line inputs has an empty style, so it renders in
the terminal's default text color. Models capture styles when they are
constructed, after `theme.Apply` has run in `main`.

## Goals / Non-Goals

**Goals:**
- One place decides the screen colors: the root view, for every mode.
- Styles stay foreground-only so nothing but the editor selection paints a
  background; the row boxes disappear as a consequence, not as a special
  case.
- Behavior on a terminal that ignores the color requests is today's minus
  the row boxes.

**Non-Goals:**
- Querying the terminal for its colors or choosing a theme from them.
- Re-styling anything beyond what the proposal names; Glamour's markdown
  style already has no backgrounds.

## Decisions

- **Set the colors on the root `tea.View`, not per mode.** The root `View`
  wraps every mode's content; adding `BackgroundColor` and
  `ForegroundColor` next to `AltScreen` covers the picker, finder, preview,
  editor and template form in one spot and keeps mode packages ignorant of
  the terminal. Alternative: each mode returning its own view with colors -
  more code, same result, and a mode could forget.
- **Expose the palette colors from the theme package as `color.Color`.**
  Two accessors on the theme package (`Background()`, `Foreground()`)
  return the current palette's values as `lipgloss.Color`, the type Bubble
  Tea accepts. Alternative: have `app` keep its own copy of the palette -
  duplicates state that `theme` already owns through `Apply`.
- **Drop the background from `Normal` rather than keep it redundant.**
  With the screen painted, a matching row background is invisible when
  the terminal honors the request and a box when it does not. Removing it
  makes the "no element sets a background" rule true in code, not only by
  coincidence of equal colors. `finder` and `tmpl` keep using `Normal`
  unchanged.
- **Style typed input text explicitly with `foreground`.** The painted
  foreground would already recolor unstyled text, but an explicit style
  keeps the palette rule true on terminals that ignore the request and
  keeps every text style defined in `theme`. `InputStyles` gets a
  foreground text style in both focus states; `EditorStyles` already has
  one.
- **Rely on Bubble Tea's reset on exit.** The renderer emits the reset
  sequences for both colors when the program stops, so clipt adds no exit
  handling. Alternative: resetting in `main` after `Run` returns - would
  duplicate what the library does and run after the alt screen is already
  closed.
- **Test the view, not the terminal.** Tests assert that `View()` from the
  root model carries the palette background and foreground in each mode
  (same shape as the existing `AltScreen` checks), that `Normal` renders
  without a background sequence, and that input styles carry the
  foreground. Whether the terminal honors the sequence is not testable
  here and is not a requirement clipt can enforce.

## Risks / Trade-offs

- [A terminal multiplexer without passthrough ignores the color request]
  → The screen falls back to the terminal's background with palette ink,
  the same as today minus the row boxes; documented as a limitation in the
  README's Themes section.
- [A user who picked a theme matching their terminal sees no difference]
  → Intended; painting the same color is a no-op visually.
- [Painting the foreground changes how unstyled text looks in any place
  that relied on the terminal default] → The audit above found only the
  single-line inputs; they get an explicit style so the result is the same
  either way.
