## 1. Front matter package

- [x] 1.1 Add `internal/frontmatter` with `Split` (first-line `---`, next `---` closes, CRLF normalized, unclosed error) and verify with table tests for present, absent, unclosed and CRLF input
- [x] 1.2 Add `Tags` parsing (flow and block sequences, scalar text, trim, drop empty, case-insensitive dedupe, non-sequence and invalid YAML errors) and verify the tags spec's Tags key scenarios as unit tests
- [x] 1.3 Add the flow-form tag serializer and verify a round-trip test that values needing quotes (`a, b`, `x: y`, `[z]`, `#t`) parse back unchanged
- [x] 1.4 Add `MergeTags` (no block, block without tags, block/flow tags entry, already present, invalid front matter, line delta) with the re-parse safety check, and verify the tags spec's merge scenarios plus delta values as unit tests
- [x] 1.5 Switch `templates.Parse` to `frontmatter.Split` and verify the existing `internal/templates` tests still pass unchanged

## 2. Templates

- [x] 2.1 Parse the `tags` key in template front matter (literal, invalid tags is a parse error) and verify the templates spec's Template tags scenarios as tests
- [x] 2.2 Change `Skeleton` to include `tags: []` and add `SkeletonWithTags`; verify the skeleton tests (parses, no tags, one text field, contains `tags: []`, tags filled when given)

## 3. Notes listing with tags

- [x] 3.1 Add `Tags` to `notes.Note` and read each file's head in `Store.List` (stop after the closing fence; broken front matter gives no tags, no error); verify with a `t.TempDir()` test covering tagged, untagged, unclosed and invalid-YAML notes, and that order and filtering of files are unchanged
- [x] 3.2 Verify with a `t.TempDir()` test that the templates store lists template tags the same way

## 4. Finder

- [x] 4.1 Add `ParseQuery` and tag-aware matching, and change `BuildRows` to use them (create row from the name query only, with `CreateTags`; no create row for tags alone); verify the tags spec's Tag query scenarios and the finder Create row scenarios as `rows_test.go` tests
- [x] 4.2 Render tags on rows dimmed after the name with `…` truncation by terminal width (track width from `WindowSizeMsg`), including the create row's tags; verify with view tests for tagged, untagged and narrow-width rows in both scopes
- [x] 4.3 Create notes with tags (`---\ntags: [..]\n---\n` written after `Create`) and templates with `SkeletonWithTags`; verify with `t.TempDir()` tests for the Create with tags, written-as-typed and template-with-tags scenarios
- [x] 4.4 Open edit for a note created with tags with the cursor on the line after the closing fence (root inserts a newline) and verify with an app test that the cursor row is after the fence and an untouched save leaves the file as created (temp dir)
- [x] 4.5 Add theme styles for tags (`theme.Tag`, dimmed row tags) in the theme package and verify `theme_test.go` covers them

## 5. Preview

- [x] 5.1 Render only the body through Glamour (whole file when unclosed) and verify the Front matter not rendered scenario with a temp-dir test
- [x] 5.2 Add the fixed tags header (wrapped, viewport height reduced by its lines, error line for invalid front matter, none for untagged) and verify the Tags header scenarios, including the header staying visible after PgDn

## 6. Template insert

- [x] 6.1 Filter the template picker with the finder's tag-aware matching and show tags on rows; verify the Filter by tag and Tags on rows scenarios in `tmpl_test.go` (temp dir)
- [x] 6.2 Carry the template's tags in `InsertMsg` for both the no-fields and the submit path and verify with `tmpl_test.go` tests
- [x] 6.3 Add `editor.InsertTemplate(body, tags)` (insert, merge, restore cursor by line delta, error line on refused merge) and route `tmpl.InsertMsg` to it in the root; verify the Tag merge on insert scenarios, including a block-list collapse above the cursor, with editor and app tests

## 7. Wrap-up

- [x] 7.1 Run `gofmt -l .`, `go vet ./...` and `go test ./...` and verify all are clean
- [x] 7.2 Manually run clipt on a scratch notes dir: tag notes, filter with `#tag` in both scopes and the picker, create with tags, insert a tagged template, and check the preview header
