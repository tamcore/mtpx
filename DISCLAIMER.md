# Disclaimer

`mtpx` deletes files on connected devices. The `delete` and `purge` commands, and
the delete action in the TUI, remove files permanently — there is no trash or undo
on an MTP device. `purge` with the default folders removes every activity, workout,
course and pace-band file on the watch.

Back up before deleting. `mtpx purge --backup <dir>` copies files off the device
first; verify the copies before you rely on them.

The software is provided "as is", without warranty of any kind. You are responsible
for any data loss or device issues that result from using it. See [LICENSE](LICENSE)
for the full terms.

mtpx is not affiliated with, endorsed by, or supported by Garmin.
