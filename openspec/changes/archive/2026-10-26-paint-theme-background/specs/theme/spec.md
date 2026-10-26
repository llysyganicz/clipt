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
SHALL set a background of its own. When clipt exits, the terminal SHALL be
left with the colors it had before.

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
- **THEN** no row, selected or not, sets a background color; only the
  screen background shows behind it

#### Scenario: Typed text uses the palette
- **WHEN** the user types into the finder query or the directory path input
  under `gruvbox_light`
- **THEN** the typed text is in the light theme's `foreground`

#### Scenario: Terminal colors restored on exit
- **WHEN** clipt quits after running with any theme
- **THEN** the terminal shows its own background and foreground colors again

#### Scenario: No color outside the palette
- **WHEN** any screen is rendered under any theme
- **THEN** every color in the output is one of that theme's palette values
