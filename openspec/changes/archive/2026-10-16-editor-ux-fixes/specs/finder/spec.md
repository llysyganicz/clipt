## MODIFIED Requirements

### Requirement: Leaving the finder
In notes scope, Esc SHALL return to preview of the current note when there is
one, and SHALL quit clipt when there is none. Esc SHALL NOT clear the query.
In template scope, Esc SHALL return to notes scope with an empty query.
`ctrl+q` SHALL quit clipt from either scope and SHALL be listed in the key
hints of both scopes. `ctrl+c` SHALL always quit. `ctrl+o` SHALL open the
notes directory picker from either scope, and accepting or cancelling it
lands in notes scope.

#### Scenario: Esc with a current note
- **WHEN** `todo` is the current note, the query is `ide`, and the user
  presses Esc
- **THEN** preview opens for `todo`

#### Scenario: Esc without a current note
- **WHEN** no note is current and the user presses Esc
- **THEN** clipt exits

#### Scenario: ctrl+q with a current note
- **WHEN** `todo` is the current note and the user presses `ctrl+q`
- **THEN** clipt exits

#### Scenario: ctrl+q in template scope
- **WHEN** the user is in template scope and presses `ctrl+q`
- **THEN** clipt exits

#### Scenario: Quit hint
- **WHEN** the finder is open in either scope
- **THEN** the key hints include `ctrl+q`

#### Scenario: Current note disappeared
- **WHEN** `todo` is the current note, its file has been removed from disk,
  and the user presses Esc
- **THEN** the finder stays open, shows an error that `todo` no longer
  exists, and clears the current note

#### Scenario: Esc in template scope
- **WHEN** no note is current, the user is in template scope and presses Esc
- **THEN** the finder shows notes scope with an empty query and clipt keeps
  running

#### Scenario: Change directory from template scope
- **WHEN** the user presses `ctrl+o` in template scope and cancels the picker
- **THEN** the finder is in notes scope
