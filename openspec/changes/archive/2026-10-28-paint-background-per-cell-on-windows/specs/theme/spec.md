## MODIFIED Requirements

### Requirement: Palette roles in the interface
The interface SHALL draw every element from the theme palette and nothing
else, with these roles: the whole screen on `background` for as long as
clipt runs, in every mode; ordinary text, including text typed into the
finder query and the directory path input, in `foreground`; the selected
row, prompts, confirmation questions and the cursor's line number in
`bright_yellow`; the create row and tags in `bright_cyan`; errors in
`bright_red`; hints, dimmed row tags and the other line numbers in
`bright_black`; the editor selection on the `selection` background; the
text cursor in `foreground`. No element other than the editor selection
SHALL set a background of its own; the screen background comes from clipt's
root screen, not from the elements. On Windows, clipt SHALL paint the
screen itself, cell by cell, and SHALL NOT change the terminal's own default
colors; on Linux and macOS clipt MAY set the terminal's default colors for
as long as it runs. When clipt exits, the terminal SHALL be left with the
colors it had before.

#### Scenario: Light theme colors the finder
- **WHEN** clipt runs with `gruvbox_light` and the finder shows a selected
  row, a create row and a hint line
- **THEN** the selected row uses the light theme's `bright_yellow`, the
  create row its `bright_cyan` and the hint its `bright_black`

#### Scenario: Screen painted in every mode
- **WHEN** clipt runs with `gruvbox_light` in the directory picker, the
  finder, preview, edit or template insert mode
- **THEN** the screen background is the light theme's `background` and the
  default text color its `foreground`

#### Scenario: Rows carry no background of their own
- **WHEN** the finder lists notes under any theme
- **THEN** no row, selected or not, sets a background color of its own; the
  screen background shows behind it

#### Scenario: Typed text uses the palette
- **WHEN** the user types into the finder query or the directory path input
  under `gruvbox_light`
- **THEN** the typed text is in the light theme's `foreground`

#### Scenario: Every cell painted on Windows
- **WHEN** clipt runs on Windows with any theme and the screen holds text
  shorter than the terminal width and fewer lines than the terminal height
- **THEN** every cell of the screen, including the empty ones after each
  line and below the last line, is on the theme's `background`, any text
  without a color of its own is in `foreground`, and the terminal's own
  default colors are unchanged

#### Scenario: Editor selection keeps its background on Windows
- **WHEN** clipt runs on Windows and the editor shows selected text
- **THEN** the selected cells are on `selection` and every other cell on
  `background`

#### Scenario: Terminal colors restored on exit
- **WHEN** clipt quits after running with any theme on Linux or macOS
- **THEN** the terminal shows its own background and foreground colors again

#### Scenario: Terminal colors untouched on Windows
- **WHEN** clipt quits after running with any theme on Windows
- **THEN** the terminal shows its own background and foreground colors,
  which clipt never changed

#### Scenario: Undrawable strip on Windows
- **WHEN** clipt runs on Windows in a terminal that leaves a column or row
  no program can draw into
- **THEN** that strip shows the terminal's own background and every cell
  clipt can draw into shows the theme's `background`

#### Scenario: No color outside the palette
- **WHEN** any screen is rendered under any theme
- **THEN** every color in the output is one of that theme's palette values
