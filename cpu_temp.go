package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

var match_temp = regexp.MustCompile(`Core\s+\d+:\s+\+([\d]+)`)

func cpu_temp() (string, error) {
	data, err := exec.Command("sensors").Output()
	if err != nil {
		return "", err
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
