# Agent guide

Notes for AI agents and contributors working on mtpx. `CLAUDE.md` is a symlink to
this file.

## What this is

A CLI and terminal UI for managing files on MTP devices (built around Garmin
watches) over USB via [libmtp](https://libmtp.sourceforge.net/). See the
[README](README.md) for user-facing docs.

## Layout

- `cmd/mtpx/` — `main`: wires the version string, the real libmtp backend, and the
  TUI launcher, then calls `cli.Execute`. This is the only package without tests.
- `internal/backend/` — device access behind the `Backend` interface
  (`Detect`/`List`/`Get`/`Put`/`Delete`). `LibmtpBackend` shells out to the `mtp-*`
  tools; `FakeBackend` is an in-memory implementation for tests. Parsers for the
  `mtp-detect`/`mtp-folders`/`mtp-files` output are pure functions with fixtures in
  `testdata/`.
- `internal/vfs/` — read-only navigation over the flat object list (children of a
  directory, path lookup, files under a folder).
- `internal/cli/` — cobra commands: `devices`, `list`, `pull`, `delete`, `purge`,
  plus the device gate (`--wait`) and the root command that launches the TUI.
- `internal/ui/` — the Bubble Tea column browser.

## Conventions

- Go 1.26.5, cobra, Bubble Tea v1 (`github.com/charmbracelet/bubbletea`).
- Tests use the standard library with table-driven cases. No testify.
- 100% statement coverage on `internal/...`, enforced by `make cover-check` and CI.
  The exceptions live in `main` (untested wiring) and the `execRunner`/TUI program
  loop, which are covered by injecting fakes (`Runner`, `Backend`) and by driving
  `ui.Run` headless with `tea.WithInput`/`tea.WithOutput`.
- `.golangci.yml` excludes `fmt.Fprint*` from errcheck; check other write errors.
- Conventional commits. Commits go straight to `master`.

## How device access stays testable

Everything that touches hardware goes through `backend.Backend`. Commands take a
`Backend` and are tested with `FakeBackend`, so `go test` needs no device. When
adding a command, inject the backend the same way and keep the run logic in a
plain function that takes `io.Writer`/`io.Reader` for output and input.

## Fixtures and privacy

`internal/backend/testdata/*` are trimmed, sanitized samples of real `mtp-*`
output: device serials are masked and filenames are synthetic. Never commit a real
device serial or personal filenames.

## Constraints found on real hardware

- Writing to the device (`push`) is not supported: `mtp-sendfile` cannot target a
  folder and Garmin rejects writes to the storage root. `backend.Put` remains for a
  future re-add.
- Only one program can use an MTP device at a time, and macOS occasionally fails to
  claim the USB interface (rerun the command).

## Commands

```
make build         # build ./bin/mtpx
make test          # go test -race with coverage
make cover-check   # enforce 100% on internal packages
make lint          # golangci-lint
```

Running against a real device needs libmtp installed (`brew install libmtp` or
`apt install libmtp-runtime`).
