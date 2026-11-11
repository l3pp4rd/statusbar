package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Memory struct {
	output
}

func memory_usage() element {
	e := &Memory{}
	go func() {
		for {
			if val, err := e.read(); err == nil {
				e.set(val)
			} else {
				log.Printf("could not read memory usage: %v", err)
			}
			time.Sleep(time.Second)
		}
	}()
	return e
}

func (k *Memory) read() (string, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return "", fmt.Errorf("read /proc/meminfo: %s", err)
	}

	// values are in kB, only the ratio matters
	var total, avail int64
	for ln := range strings.Lines(string(data)) {
		f := strings.Fields(ln)
		if len(f) < 2 {
			continue
		}
		switch f[0] {
		case "MemTotal:":
			if total, err = strconv.ParseInt(f[1], 10, 64); err != nil {
				return "", fmt.Errorf("read total memory: %s", err)
			}
		case "MemAvailable:":
			if avail, err = strconv.ParseInt(f[1], 10, 64); err != nil {
				return "", fmt.Errorf("read available memory: %s", err)
			}
		}
		if total > 0 && avail > 0 {
			break
		}
	}
	if total == 0 {
		return "", fmt.Errorf("could not find MemTotal in /proc/meminfo")
	}
	perc := 100 * (total - avail) / total

	var color string
	switch {
	case perc >= 90:
		color = "#dc322f"
	case perc >= 70:
		color = "#b58900"
	default:
		color = "#859900"
	}

	return fmt.Sprintf("^fg(%s)%d%% ^i(%s)^fg()", color, perc, xbm("mem")), nil
}
