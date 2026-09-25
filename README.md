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
Workday end: 14:27
Remaining:   06:52
```

If the workday is over, `Remaining` shows a negative value with `(overtime)`.

### Flags

| Flag        | Default | Description                          |
|-------------|---------|---------------------------------------|
| `-workday`  | `7h45m` | Workday duration (e.g. `8h`, `7h30m`) |

```powershell
.\time2go.exe -workday 8h
```
