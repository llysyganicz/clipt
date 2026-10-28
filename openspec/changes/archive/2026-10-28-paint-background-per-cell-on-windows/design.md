## Context

See proposal.md - Why. The root model's `View` returns a `tea.View` with
`AltScreen`, `BackgroundColor` and `ForegroundColor` set from the theme
(`internal/ui/app/app.go`). Bubble Tea's renderer turns the two colors into
the terminal's "set default color" requests when the program starts and the
matching "reset default color" requests when it stops. Every text style in
`internal/ui/theme` is foreground-only; only the editor selection carries a
background, so apart from that the content Bubble Tea renders has no
background on any cell and the terminal's default shows through.

What was verified: clipt built from the current commit writes the set
requests on start and the reset requests on exit, in that order, after
leaving the alternate screen. Alacritty 0.17 and xterm.js both restore
their colors when fed exactly those bytes. On Windows a console host sits
between clipt and the terminal and re-interprets everything clipt writes;
the one built into Windows understands the set requests and not the resets,
and alacritty uses the built-in one unless the newer console from Windows
Terminal is placed on its path. The console also drops the "what are your
colors" queries, so clipt cannot learn the original colors and set them
back either.

Bubble Tea's view content is a string. The renderer parses it into a grid
of cells, each with its own style, and only writes cells whose style or
text changed. The ultraviolet package that Bubble Tea and Lip Gloss are
built on exposes that grid (`ScreenBuffer`, `Cell`, `Style` with `Fg` and
`Bg`) and can draw a styled string into it and render the grid back to a
string. The root model already tracks the terminal width and height from
the window size message.

## Goals / Non-Goals

**Goals:**
- On Windows the screen looks as it does today while clipt runs and the
  terminal's default colors are never touched, so there is nothing to
  restore.
- One decision point, in the root model, picks the mechanism; mode models
  stay ignorant of the terminal and of the platform.
- The per-cell paint is a pure function that tests run on Linux.

**Non-Goals:**
- Guessing whether the console host in use handles resets. The choice is by
  operating system only.
- Changing the Linux and macOS path, which works.
- Touching the mode models' views or the theme styles.

## Decisions

- **Branch on the operating system, not on the terminal.** The root view
  takes the per-cell path when the program runs on Windows and the
  terminal-color path otherwise. Alternatives: detecting Windows Terminal
  through its session variable or looking for the newer console library
  next to the terminal - the variable is absent in alacritty even when it
  has the fixed console, the library check is a guess about another
  program's files, and both leave a wrong guess recoloring the terminal.
  The per-cell path is visually the same except for the undrawable strip,
  so taking it on every Windows setup costs little.
- **Paint over the cell grid, after the content is composed.** The root
  model draws the mode's content into a `ScreenBuffer` of the terminal's
  width and height, then walks every cell: a cell without a background gets
  the palette `background`, a cell without a foreground gets the palette
  `foreground`, cells that already have one (the editor selection) are left
  alone. The buffer renders back to the string the view carries. Empty
  cells after a line's end and below the last line are painted like any
  other, which is what makes "every cell" true by construction.
  Alternatives: wrapping the content in a Lip Gloss style with a background
  and the screen's width and height - each element's own style ends with a
  reset that cancels the outer background until the next line, so the
  paint has holes after every styled element; giving every theme style a
  background and padding each row - the approach the previous change
  removed, which leaves gaps between elements and cannot reach the text
  Glamour and the textarea produce.
- **Render the grid with ultraviolet's own renderer, not Lip Gloss's
  canvas.** Lip Gloss's `Canvas.Render` trims trailing spaces from every
  line, which would strip the painted cells after each line's text. The
  underlying buffer's `Render` keeps them. Cost: `ultraviolet` moves from
  an indirect to a direct dependency, at the version Bubble Tea already
  pins. Alternative: assembling the string by hand from the cells - more
  code that the library already has.
- **Do not set the view's foreground on Windows either.** The reset for the
  foreground is dropped by the same console, and the cell walk gives every
  uncolored cell the palette `foreground`, so the view sets neither color
  on Windows. The textarea's virtual cursor keeps its own color from the
  theme.
- **Keep the per-cell paint in the theme package.** A function in
  `internal/ui/theme` takes the composed content and the screen size and
  returns the painted string; it owns the palette colors already through
  `Background()` and `Foreground()`. The root model calls it from `View`
  when the Windows path is active. Alternative: a new package - one
  function does not justify it.
- **Make the platform choice overridable for tests.** The root model reads
  the choice from a package-level value initialised from the operating
  system, which tests set to exercise the Windows path on Linux: the view
  then carries no terminal colors and its content is painted. The paint
  function itself is tested directly: every cell carries the background,
  lines are padded to the width and the grid to the height, a cell with
  its own background keeps it, a cell without a foreground gets the palette
  foreground, a cell with one keeps it.
- **Before the first window size message, paint nothing.** With a zero
  width or height the content is passed through unchanged; the first frame
  Bubble Tea draws after the size arrives is painted. Alternative: a
  guessed default size - wrong on every terminal but one.

## Risks / Trade-offs

- [The terminal keeps a column or row no program can draw into] → It shows
  the terminal's own background on Windows. Documented in the README's
  Themes section; invisible with a dark theme on a dark terminal.
- [Each frame on Windows is parsed into a grid, painted and serialized
  once more before Bubble Tea parses it again] → The work is bounded by
  the screen size, a few thousand cells, and runs only on a redraw.
- [Bubble Tea's renderer could drop trailing styled spaces when it writes
  a frame] → Verified with a pty capture on Linux: the renderer does not
  write the trailing cells one by one but fills the rest of each row with
  erase-to-end-of-line while the pen carries the painted background, which
  relies on background color erase (BCE). ultraviolet assumes BCE for every
  terminal and offers no way to force per-cell output. All xterm-compatible
  terminals and the Windows console host fill erased cells with the current
  background, so the result is the same; the Windows run is the check.
- [ultraviolet is a pre-release module with an unstable API] → Pin the
  pseudo-version Bubble Tea already requires; the four calls used (new
  buffer, draw a styled string, read and set cells, render) are its core.
- [Windows Terminal 1.23 and later would have honored the resets] → The
  per-cell result there differs only by the undrawable strip; accepted
  rather than detecting the terminal.

## Migration Plan

No data or config changes. Rollback is reverting the commit; the Linux and
macOS path is untouched.
