# time2go

Windows CLI tool that shows your system boot time, current uptime, and the
remaining time left in your workday.

## Build

```powershell
go build -o time2go.exe .
```

## Usage

```powershell
.\time2go.exe
```

Example output:

```
Boot time:   06:42
Uptime:      00:53
Workday:     08:00
Workday end: 14:27
Remaining:   06:52
[███░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]
```

The progress bar shows worked time in 15 minute blocks. Blocks beyond the end
of the workday are overtime and shown in red.

If the workday is over, `Remaining` shows a negative value with `(overtime)`.

### Custom start time

By default the workday starts at boot time. Use `-start hh:mm` (or `--start`)
to use a different start time today instead; boot time and uptime are then
ignored (the output shows `Start time` and `Worked`).

```powershell
.\time2go.exe -start 07:30
```

### Configuring the workday duration

The workday duration defaults to `8h` and can be set, in order of precedence
(highest wins):

1. **Command-line flag:** `-workday`
   ```powershell
   .\time2go.exe -workday 7h45m
   ```
2. **Environment variable:** `time2go_workday`
   ```powershell
   $env:time2go_workday = "7h45m"
   ```
3. **Config file:** `~/.config/time2go` — a plain text file containing a Go
   duration string (e.g. `7h45m`)
4. **Default:** `8h`
