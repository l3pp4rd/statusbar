package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ac_check = regexp.MustCompile(`online:\s+(.+)`)
var bat_check = regexp.MustCompile(`percentage:\s+([^%]+)`)

type pw_sources struct {
	battery string
	ac      string
	val     string
}

func (s *pw_sources) value() string {
	return s.val
}

func power() element {
	// @TODO do not depend on upower
	data, err := exec.Command("upower", "-e").Output()
	if err != nil {
		log.Fatalf("upower enumerate command: %s", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	e := &pw_sources{}
	for _, line := range lines {
		if i := strings.Index(line, "_AC"); i != -1 {
			e.ac = strings.TrimSpace(line)
		} else if i := strings.Index(line, "_BAT"); i != -1 {
			e.battery = strings.TrimSpace(line)
		}
	}
	go func() {
		for {
			if val, err := e.read(); err == nil {
				e.val = val
			} else {
				log.Printf("could not read power: %v", err)
			}
			time.Sleep(time.Second * 3)
		}
	}()
	return e
}

func (s *pw_sources) is_on_ac() (bool, error) {
	if len(s.ac) == 0 {
		return false, nil
	}

	data, err := exec.Command("upower", "-i", s.ac).Output()
	if err != nil {
		return false, fmt.Errorf("upower cmd is on AC: %s", err)
	}
	m := ac_check.FindStringSubmatch(string(data))
	if len(m) != 2 {
		return false, fmt.Errorf("number of matches was not expected for AC power check")
	}

	return m[1] == "yes", nil
}

func (s *pw_sources) battery_percent() (int, error) {
	if len(s.battery) == 0 {
		return 0, nil
	}

	data, err := exec.Command("upower", "-i", s.battery).Output()
	if err != nil {
		return 0, fmt.Errorf("upower battery percent check: %s", err)
	}
	m := bat_check.FindStringSubmatch(string(data))
	if len(m) != 2 {
		return 0, fmt.Errorf("number of matches was not expected for BATTERY power check")
	}

	return strconv.Atoi(m[1])
}

func (s *pw_sources) read() (string, error) {
	on, err := s.is_on_ac()
	if err != nil {
		return "", err
	}

	if on || len(s.battery) == 0 {
		return "^i(" + xbm("power-ac") + ")", nil
	}

	perc, err := s.battery_percent()
	if err != nil {
		return "", err
	}

	var color, icon string
	switch {
	case perc <= 20:
		icon = xbm("bat-low")
		color = "#dc322f"
	case perc <= 50:
		icon = xbm("bat-mid")
		color = "#b58900"
	default:
		icon = xbm("bat-full")
		color = "#859900"
	}

	return fmt.Sprintf("^fg(%s)%d%%^i(%s)^fg()", color, perc, icon), nil
}
