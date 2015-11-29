package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func keyboard_layout() (string, error) {
	data, err := exec.Command("setxkbmap", "-print").Output()
	if err != nil {
		return "", err
	}

	r := regexp.MustCompile(`xkb_symbols[^"]+"([^"]+)`)

	m := r.FindStringSubmatch(string(data))
	if len(m) != 2 {
		return "", fmt.Errorf("could not extract keybord layout from %s", string(data))
	}

	parts := strings.Split(m[1], "+")
	if len(parts) < 2 {
		return "", fmt.Errorf("expected at least two elements in: %s", m[1])
	}

	return strings.ToUpper(parts[1]), nil
}
