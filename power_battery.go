package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var ac_check = regexp.MustCompile(`online:\s+(.+)`)
var bat_check = regexp.MustCompile(`percentage:\s+([^%]+)`)

type pw_sources struct {
	battery string
	ac      string
}

func (src *pw_sources) is_on_ac() (bool, error) {
	if len(src.ac) == 0 {
		return false, nil
	}

	data, err := exec.Command("upower", "-i", src.ac).Output()
	if err != nil {
		return false, err
	}
	m := ac_check.FindStringSubmatch(string(data))
	if len(m) != 2 {
		return false, fmt.Errorf("number of matches was not expected for AC power check")
	}

	return m[1] == "yes", nil
}

func (src *pw_sources) battery_percent() (int, error) {
	if len(src.battery) == 0 {
		return 0, nil
	}

	data, err := exec.Command("upower", "-i", src.battery).Output()
	if err != nil {
		return 0, err
	}
	m := bat_check.FindStringSubmatch(string(data))
	if len(m) != 2 {
		return 0, fmt.Errorf("number of matches was not expected for BATTERY power check")
	}

	return strconv.Atoi(m[1])
}

var pw_found *pw_sources

func power_battery() (string, error) {
	if pw_found == nil {
		src, err := power_detect_sources()
		if err != nil {
			return "", err
		}
		pw_found = src
	}

	on, err := pw_found.is_on_ac()
	if err != nil {
		return "", err
	}

	if on || len(pw_found.battery) == 0 {
		return "^i(" + xbm("power-ac") + ")", nil
	}

	perc, err := pw_found.battery_percent()
	if err != nil {
		return "", err
	}

	var color string
	switch {
	case perc <= 20:
		color = "#dc322f"
	case perc <= 50:
		color = "#b58900"
	default:
		color = "#859900"
	}

	return fmt.Sprintf("^fg(%s)%d%%^i(%s)^fg()", color, perc, xbm("power-bat")), nil
}

func power_detect_sources() (*pw_sources, error) {
	_, err := exec.Command("upower", "--version").Output()
	if err != nil {
		return nil, fmt.Errorf("needs 'upower' installed on system")
	}

	data, err := exec.Command("upower", "-e").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	src := &pw_sources{}
	for _, line := range lines {
		if i := strings.Index(line, "_AC"); i != -1 {
			src.ac = strings.TrimSpace(line)
		} else if i := strings.Index(line, "_BAT"); i != -1 {
			src.battery = strings.TrimSpace(line)
		}
	}
	return src, nil
}
