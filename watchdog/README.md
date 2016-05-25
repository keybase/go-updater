## Watchdog

The watchdog is in charge of ensuring that a list of programs are always running.
The watchdog will start these programs (it will be the parent of those processes).
If the watchdog is terminated, the monitored programs will also be terminated.

When the watchdog starts up:
- It will kill any other watchdogs (processes at the same executable path)
- It will kill any programs that it will be monitoring

After startup:
- It will monitor a list of programs and will restart them if they exit
