package main

import (
	"flag"
	"fmt"
	"syscall"
	"time"
)

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
	workday := flag.Duration("workday", 7*time.Hour+45*time.Minute, "workday duration (e.g. 7h45m)")
	flag.Parse()

	boot := bootTime()
	uptime := time.Since(boot)
	remaining := *workday - uptime
	end := boot.Add(*workday)

	fmt.Printf("Boot time:   %s\n", boot.Round(time.Minute).Format("15:04"))
	fmt.Printf("Uptime:      %s\n", formatDuration(uptime))
	fmt.Printf("Workday end: %s\n", end.Round(time.Minute).Format("15:04"))
	if remaining >= 0 {
		fmt.Printf("Remaining:   %s\n", formatDuration(remaining))
	} else {
		fmt.Printf("Remaining:   %s (overtime)\n", formatDuration(remaining))
	}
}
