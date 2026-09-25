package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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

	boot := bootTime()
	uptime := time.Since(boot)
	remaining := workday - uptime
	end := boot.Add(workday)

	fmt.Printf("Boot time:   %s\n", boot.Round(time.Minute).Format("15:04"))
	fmt.Printf("Uptime:      %s\n", formatDuration(uptime))
	fmt.Printf("Workday end: %s\n", end.Round(time.Minute).Format("15:04"))
	if remaining >= 0 {
		fmt.Printf("Remaining:   %s\n", formatDuration(remaining))
	} else {
		fmt.Printf("Remaining:   %s (overtime)\n", formatDuration(remaining))
	}
}
