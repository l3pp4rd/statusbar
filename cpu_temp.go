package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var match_temp = regexp.MustCompile(`Core\s+\d+:\s+\+([\d]+)`)

type CpuTemp struct {
	val string
}

func (k *CpuTemp) value() string {
	return k.val
}

func cpu_temp() element {
	e := &CpuTemp{}
	go func() {
		for {
			if val, err := e.read(); err == nil {
				e.val = val
			} else {
				log.Printf("could not read cpu temp: %v", err)
			}
			time.Sleep(time.Second * 2)
		}
	}()
	return e
}

func (k *CpuTemp) read() (string, error) {
	data, err := exec.Command("sensors").Output()
	if err != nil {
		return "", fmt.Errorf("'sensors' command: %s", err)
	}

	var cores []int
	var total int
	for _, match := range match_temp.FindAllStringSubmatch(string(data), -1) {
		core, err := strconv.Atoi(match[1])
		if err != nil {
			return "", fmt.Errorf("failed to parse cpu temp from: %s - %s", match[1], err)
		}
		cores = append(cores, core)
		total += core
	}

	if len(cores) == 0 {
		return "", nil
	}

	c := total / len(cores)
	var color string
	switch {
	case c >= 80:
		color = "#dc322f"
	case c >= 60:
		color = "#b58900"
	default:
		color = "#859900"
	}
	return fmt.Sprintf("^fg(%s)%d °C^fg()", color, c), nil
}
