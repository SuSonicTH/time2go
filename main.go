package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const defaultWorkday = 8 * time.Hour

func workdayFromConfig() (time.Duration, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, false
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "time2go"))
	if err != nil {
		return 0, false
	}
	text := strings.TrimPrefix(string(data), string([]byte{0xEF, 0xBB, 0xBF}))
	d, err := time.ParseDuration(strings.TrimSpace(text))
	if err != nil {
		return 0, false
	}
	return d, true
}

func workdayFromEnv() (time.Duration, bool) {
	val, ok := os.LookupEnv("time2go_workday")
	if !ok {
		return 0, false
	}
	d, err := time.ParseDuration(strings.TrimSpace(val))
	if err != nil {
		return 0, false
	}
	return d, true
}

func resolveWorkday(flagVal time.Duration, flagSet bool) time.Duration {
	if flagSet {
		return flagVal
	}
	if d, ok := workdayFromEnv(); ok {
		return d
	}
	if d, ok := workdayFromConfig(); ok {
		return d
	}
	return defaultWorkday
}

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetTickCount = kernel32.NewProc("GetTickCount64")
)

func bootTime() time.Time {
	ret, _, _ := procGetTickCount.Call()
	uptime := time.Duration(ret) * time.Millisecond
	return time.Now().Add(-uptime)
}

// parseStart parses an hh:mm time as today's local time. It must not be in
// the future relative to now.
func parseStart(s string, now time.Time) (time.Time, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is not a valid hh:mm time", s)
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
	if start.After(now) {
		return time.Time{}, fmt.Errorf("start time %s is in the future", s)
	}
	return start, nil
}

const (
	barInterval = 15 * time.Minute
	ansiRed     = "\x1b[31m"
	ansiReset   = "\x1b[0m"
)

// enableColor turns on ANSI escape processing for the console and reports
// whether colors can be used.
func enableColor() bool {
	const enableVirtualTerminalProcessing = 0x0004
	getMode := kernel32.NewProc("GetConsoleMode")
	setMode := kernel32.NewProc("SetConsoleMode")
	h := uintptr(syscall.Stdout)
	var mode uint32
	if r, _, _ := getMode.Call(h, uintptr(unsafe.Pointer(&mode))); r == 0 {
		return false
	}
	if r, _, _ := setMode.Call(h, uintptr(mode|enableVirtualTerminalProcessing)); r == 0 {
		return false
	}
	return true
}

// progressBar renders worked time in 15 minute blocks. Blocks beyond the end
// of the workday are overtime, shown in red.
func progressBar(uptime, workday time.Duration, color bool) string {
	total := int((workday + barInterval - 1) / barInterval)
	worked := int(uptime / barInterval)

	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < total; i++ {
		if i < worked {
			sb.WriteString("█")
		} else {
			sb.WriteString("░")
		}
	}
	if over := worked - total; over > 0 {
		if color {
			sb.WriteString(ansiRed)
		}
		sb.WriteString(strings.Repeat("█", over))
		if color {
			sb.WriteString(ansiReset)
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func formatDuration(d time.Duration) string {
	sign := ""
	if d < 0 {
		sign = "-"
		d = -d
	}
	totalMin := int(d.Round(time.Minute).Minutes())
	h := totalMin / 60
	m := totalMin % 60
	return fmt.Sprintf("%s%02d:%02d", sign, h, m)
}

func main() {
	flagVal := flag.Duration("workday", defaultWorkday, "workday duration (e.g. 7h45m)")
	startVal := flag.String("start", "", "start time of the workday as hh:mm (replaces boot time)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), `
Workday duration precedence (highest wins):
  1. -workday flag
  2. time2go_workday environment variable
  3. ~/.config/time2go file (plain text duration, e.g. "7h45m")
  4. default: %s
`, defaultWorkday)
	}
	flag.Parse()

	flagSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "workday" {
			flagSet = true
		}
	})
	workday := resolveWorkday(*flagVal, flagSet)

	start := bootTime()
	startLabel, uptimeLabel := "Boot time:", "Uptime:"
	if *startVal != "" {
		t, err := parseStart(*startVal, time.Now())
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid -start value: %v\n", err)
			os.Exit(2)
		}
		start = t
		startLabel, uptimeLabel = "Start time:", "Worked:"
	}
	uptime := time.Since(start)
	remaining := workday - uptime
	end := start.Add(workday)

	fmt.Printf("%-12s %s\n", startLabel, start.Round(time.Minute).Format("15:04"))
	fmt.Printf("%-12s %s\n", uptimeLabel, formatDuration(uptime))
	fmt.Printf("Workday:     %s\n", formatDuration(workday))
	fmt.Printf("Workday end: %s\n", end.Round(time.Minute).Format("15:04"))
	if remaining >= 0 {
		fmt.Printf("Remaining:   %s\n", formatDuration(remaining))
	} else {
		fmt.Printf("Remaining:   %s (overtime)\n", formatDuration(remaining))
	}
	fmt.Println(progressBar(uptime, workday, enableColor()))
}
