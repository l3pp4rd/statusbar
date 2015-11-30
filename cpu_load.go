package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func cpu_load() (string, error) {
	data, err := ioutil.ReadFile(CPU_LOAD_FILE)
	if err != nil {
		return "", fmt.Errorf("read cpu load from %s - %s", CPU_LOAD_FILE, err)
	}

	parts := strings.Split(strings.TrimSpace(string(data)), " ")
	load, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return "", fmt.Errorf("parse cpu average load: %s", err)
	}
	var color string
	switch {
	case load >= 10:
		color = "#dc322f"
	case load >= 4:
		color = "#b58900"
	default:
		color = "#6c71c4"
	}
	return fmt.Sprintf("^fg(%s)%.02f ^i(%s)^fg()", color, load, xbm("load")), nil
}
