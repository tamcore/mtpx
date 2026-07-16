# mtpx

A command-line file manager for MTP devices, with a terminal UI. It was built to
pull files off, clear out, and back up Garmin watches over USB, and works with
other MTP devices too.

## Requirements

`mtpx` talks to devices through [libmtp](https://libmtp.sourceforge.net/) and its
`mtp-*` command-line tools:

- macOS: `brew install libmtp`
- Debian/Ubuntu: `apt install libmtp-runtime`

The device must be in MTP/USB mode. On macOS, quit Garmin Express and unmount the
watch in Finder first — only one program can use an MTP device at a time.

Commands fail with a clear error when no device is found rather than doing nothing.
Pass `--wait <duration>` (for example `--wait 30s`) to poll until a device appears.

## Install

```
go install github.com/tamcore/mtpx/cmd/mtpx@latest
```

Or download a binary from the [releases](https://github.com/tamcore/mtpx/releases)
page.

## Usage

Run `mtpx` with no arguments to open the browser:

```
mtpx
```

The browser is a Finder-style column view: each folder you open adds a pane to the
right, and the highlighted folder is previewed in the next pane. Keys: `↑`/`↓` move,
`→`/`enter` open a folder, `←`/`⌫` go back, `space` select, `c` copy the selection to
disk, `d` delete the selection, `r` refresh, `q` quit. Pulled files go to the current
directory unless you pass `--dest`.

The same actions are available as subcommands for scripts:

```
mtpx list [path]                 # list a folder (add --json for machine output)
mtpx pull <remote> <local>       # copy a file off the device
mtpx delete <path>...            # delete files (--dry-run, --yes)
mtpx purge                       # delete files in the Garmin activity folders
```

### purge

`purge` deletes every file, but no folders, inside the target folders. With no
flags it clears `Activity`, `Workouts`, `Courses` and `PaceBands`:

```
mtpx purge                                   # asks for confirmation first
mtpx purge --dry-run                         # print what would be deleted
mtpx purge --backup ~/garmin-backup          # copy files out, then delete
mtpx purge --folders Activity,Courses --yes  # pick folders, skip the prompt
```

`--backup` copies each file to the given directory (keeping its device path) before
deleting it. If any copy fails, nothing is deleted.

Deletion asks for a `y/N` confirmation unless you pass `--yes`. In a non-interactive
shell without `--yes`, it reads no confirmation and does nothing.

## Limitations

- **No writing to the device.** libmtp's `mtp-sendfile` cannot target a folder, and
  Garmin devices reject writes to the storage root, so copying files onto the watch
  is not supported. This was confirmed against a Garmin EPIX 2.
- On macOS, libmtp occasionally fails to claim the USB interface. Rerun the command.

## Development

```
make build         # build ./bin/mtpx
make test          # go test -race with coverage
make cover-check   # fail if internal packages drop below 100% coverage
make lint          # golangci-lint
```

Device access sits behind a `backend.Backend` interface with an in-memory fake, so
the tests run without hardware. CI enforces 100% coverage on the `internal`
packages.

## License

MIT — see [LICENSE](LICENSE). Read the [DISCLAIMER](DISCLAIMER.md) before deleting
anything.
