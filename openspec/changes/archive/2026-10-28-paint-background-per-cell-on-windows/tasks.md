## 1. Per-cell paint in the theme package

- [x] 1.1 Add `github.com/charmbracelet/ultraviolet` to `go.mod` as a direct dependency at the pseudo-version Bubble Tea already requires; verify `go mod tidy` leaves `go.mod` and `go.sum` unchanged apart from moving the line out of the indirect block and `go build ./...` passes
- [x] 1.2 Add `theme.PaintScreen(content string, width, height int) string` that draws `content` as a styled string into a screen buffer of the given size, gives every cell without a background the palette `background` and every cell without a foreground the palette `foreground`, leaves cells that already have either alone, and renders the buffer with ultraviolet's renderer (not the Lip Gloss canvas); with a zero width or height it returns `content` unchanged; verify theme tests under `gruvbox_light` assert that every cell of the output, including the padding after a short line and the rows below the last line, carries the light `background`, that a cell styled with the `selection` background keeps it, that text styled with `bright_yellow` keeps that foreground, that unstyled text gets the light `foreground`, and that `PaintScreen(s, 0, 0) == s`
- [x] 1.3 Add a `TestNoColorLiteralsOutsideTheme`-style check that the painted output under each built-in theme uses only that palette's colors; verify the test passes for `gruvbox_dark` and `gruvbox_light`

## 2. Root view chooses the mechanism

- [x] 2.1 In `internal/ui/app`, add a package-level `paintPerCell` value initialised to `runtime.GOOS == "windows"`; in `Model.View`, when it is false keep today's behavior (terminal `BackgroundColor` and `ForegroundColor` from the theme), and when it is true set neither color and pass the content through `theme.PaintScreen` with the model's width and height; verify an app test sets `paintPerCell = true` for its duration, applies `gruvbox_light`, drives the model through picker, finder, preview, edit and template insert, and asserts in every mode that `View()` has nil `BackgroundColor` and `ForegroundColor` and that its content's first line is padded to the model's width with cells on the light `background`
- [x] 2.2 Keep the existing `TestViewPaintsThemeColors` passing with `paintPerCell = false`, and add a check that before any window size message the content is returned unpadded on the per-cell path; verify both tests pass

## 3. Docs and context

- [x] 3.1 Update the README's "Themes" section: on Linux and macOS clipt sets the terminal's own background and foreground and restores them on exit; on Windows it paints every cell of its screen instead and never changes the terminal's colors, so a column or row the terminal keeps for itself shows the terminal's own background; verify by reading the rendered section
- [x] 3.2 Update the themes paragraph in `openspec/config.yaml` project context to say the root view sets the terminal colors on Linux and macOS and paints per cell on Windows; verify `openspec validate paint-background-per-cell-on-windows` passes

## 4. Verification

- [x] 4.1 Run `gofmt -l .`, `go vet ./...` and `go test ./...` and confirm all clean
- [x] 4.2 On Linux, run the built binary under `script` with `paintPerCell` forced on through a temporary test build, quit it, and confirm in the capture that no OSC 10/11 sequence is written and that every row's tail is painted on the background color: Bubble Tea's renderer fills the rest of a row with erase-to-end-of-line (`ESC[K`) while the pen carries the background (`48;2;`), relying on background color erase, rather than writing each trailing cell; the Windows run in 4.3 is the check that the console host fills those cells too
- [x] 4.3 On Windows in alacritty, run the built binary with `gruvbox_light` and `gruvbox_dark`, check every mode is drawn on the theme background with text in the palette foreground, quit, and confirm the terminal shows its own colors immediately after exit
