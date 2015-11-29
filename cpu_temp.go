package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func cpu_temp() (string, error) {
	data, err := ioutil.ReadFile(CPU_TEMP_FILE)
	if err != nil {
		return "", err
	}

	temp, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return "", err
	}
	c := temp / 1000
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
