## MODIFIED Requirements

### Requirement: Release artifacts
A release SHALL contain exactly these two archives, each holding the clipt
binary built from the tagged commit for the x86-64 architecture without cgo,
plus `README.md` and a `themes/` folder with the built-in theme files
`gruvbox_dark.toml` and `gruvbox_light.toml`:

| Archive                           | Binary      |
|-----------------------------------|-------------|
| `clipt_X.Y.Z_linux_amd64.tar.gz`  | `clipt`     |
| `clipt_X.Y.Z_windows_amd64.zip`   | `clipt.exe` |

`X.Y.Z` is the tag name without the leading `v`. The release SHALL also
contain `checksums.txt` listing the SHA-256 sum of every archive.

#### Scenario: Assets of v0.1.0
- **WHEN** the release run for `v0.1.0` succeeds
- **THEN** the GitHub Release has the assets `clipt_0.1.0_linux_amd64.tar.gz`,
  `clipt_0.1.0_windows_amd64.zip` and `checksums.txt`, and no others

#### Scenario: Linux archive content
- **WHEN** `clipt_0.1.0_linux_amd64.tar.gz` is extracted on an x86-64 Linux
  machine
- **THEN** it yields an executable `clipt` that runs without any shared
  library beyond the system's C library, a `README.md`, and
  `themes/gruvbox_dark.toml` and `themes/gruvbox_light.toml`

#### Scenario: Windows archive content
- **WHEN** `clipt_0.1.0_windows_amd64.zip` is extracted
- **THEN** it yields `clipt.exe`, `README.md`, `themes\gruvbox_dark.toml`
  and `themes\gruvbox_light.toml`

#### Scenario: Shipped themes match the binary
- **WHEN** a theme file from the archive's `themes/` folder is compared with
  the same-named theme built into the archive's binary
- **THEN** they are identical

#### Scenario: Checksums match
- **WHEN** the SHA-256 sum of a downloaded archive is compared with its line
  in `checksums.txt`
- **THEN** they are equal
