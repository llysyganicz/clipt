## Why

On Windows the terminal's own colors do not come back when clipt exits: the
screen stays on the theme's `background` and `foreground`. clipt asks the
terminal to change its default colors while it runs and to reset them on exit,
but on Windows the console host that sits between clipt and the terminal
applies the change and discards the reset (the built-in console only learned
the reset sequences in the open-source console in 2025, and terminals such as
alacritty use the built-in one). The same bytes restore the colors correctly
on Linux. Windows is the platform clipt is used on most, so it must not leave
the terminal recolored.

## What Changes

- On Windows, clipt stops changing the terminal's default colors. The root
  view instead paints every cell of its own screen in the palette
  `background` and gives any cell without a text color the palette
  `foreground`, so the screen looks the same as today while clipt runs and
  the terminal is untouched when it exits.
- On Linux and macOS nothing changes: the root view keeps setting the
  terminal's background and foreground, and Bubble Tea keeps resetting them
  on exit.
- On Windows, any strip of the terminal that no program can draw into (a
  spare column or row left over by the terminal) shows the terminal's own
  background rather than the theme's. The README's Themes section documents
  the difference.

Non-goals: detecting which console host or terminal is present (environment
variables, conpty.dll, Windows Terminal), querying the terminal for its colors,
restoring colors by other means on exit, any change to the theme file format,
or any change to how the other platforms paint the screen.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `theme`: the "Palette roles in the interface" requirement changes from
  "the whole screen on `background`" by way of the terminal's default colors
  to the same visible result on every platform, with the terminal's own
  colors never changed on Windows; the "terminal colors restored on exit"
  scenario gains the Windows case, where the colors were never changed, and
  a scenario covers the terminal's undrawable strip on Windows.

## Impact

- `internal/ui/app`: the root `View` chooses, by operating system, between
  setting the view's terminal colors and painting the content per cell.
- `internal/ui/theme`: the per-cell paint over a screen buffer, applied to
  the composed content.
- `go.mod`: `github.com/charmbracelet/ultraviolet` becomes a direct
  dependency (it is already pulled in through Bubble Tea and Lip Gloss) for
  the screen buffer the paint runs over.
- Tests in `internal/ui/app` and the new helper cover both paths without a
  real terminal; the Windows path is exercised on Linux by calling the paint
  directly.
- `README.md`: the Themes section describes the Windows behavior.
