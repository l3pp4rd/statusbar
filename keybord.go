package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var match_layout = regexp.MustCompile(`xkb_symbols[^"]+"([^"]+)`)

type Keyboard struct {
	output
}

func keyboard() element {
	e := &Keyboard{}
	go func() {
		for {
			if val, err := e.layout(); err == nil {
				e.set(val)
			} else {
				log.Printf("could not read keyboard layout: %v", err)
			}
			time.Sleep(time.Second)
		}
	}()
	return e
}

func (k *Keyboard) layout() (string, error) {
	data, err := exec.Command("setxkbmap", "-print").Output()
	if err != nil {
		return "", fmt.Errorf("'setxkbmap -print' command: %s", err)
	}

	m := match_layout.FindStringSubmatch(string(data))
	if len(m) != 2 {
		return "", fmt.Errorf("could not extract keybord layout from %s", string(data))
	}

	parts := strings.Split(m[1], "+")
	if len(parts) < 2 {
		return "", fmt.Errorf("expected at least two elements in keyboard details: %s", m[1])
	}

	return "^fg(white) " + parts[1] + "^fg()", nil
}
