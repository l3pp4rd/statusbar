package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

const (
	KB       = 1024
	MB       = KB * KB
	DOWNLOAD = "rx"
	UPLOAD   = "tx"
)

var nw_colors = map[string]string{
	DOWNLOAD: "#859900",
	UPLOAD:   "#dc322f",
}

var nw_current *nw_stats

type nw_stats struct {
	device, typ, ssid string
	rx, tx            int64
}

func (s *nw_stats) signal_strength() (int, error) {
	lines, err := exec.Command("nmcli", "-t", "-f", "SSID,SIGNAL", "device", "wifi", "list").Output()
	if err != nil {
		return 0, err
	}
	for _, ln := range strings.Split(string(lines), "\n") {
		parts := strings.Split(strings.TrimSpace(ln), ":")
		if parts[0] != s.ssid {
			continue
		}

		return strconv.Atoi(parts[1])
	}
	return 0, nil
}

func network_stats() (string, error) {
	lines, err := exec.Command("nmcli", "-t", "-f", "DEVICE,TYPE,STATE,CONNECTION", "device", "status").Output()
	if err != nil {
		return "", fmt.Errorf("nmcli failed to get information on devices: %s", err)
	}

	var stats *nw_stats
	for _, ln := range strings.Split(string(lines), "\n") {
		parts := strings.Split(strings.TrimSpace(ln), ":")
		switch {
		case len(parts) != 4:
			continue
		case parts[1] == "bridge":
			continue
		case parts[1] == "loopback":
			continue
		case parts[2] != "connected":
			continue
		}

		stats = &nw_stats{
			device: parts[0],
			typ:    parts[1],
			ssid:   parts[3],
		}
		break
	}

	if nil == stats {
		return fmt.Sprintf("^i(%s)", xbm("net-wired")), nil
	}

	stats.rx, err = network_device_bytes(stats.device, DOWNLOAD)
	if err != nil {
		return "", err
	}
	stats.tx, err = network_device_bytes(stats.device, UPLOAD)
	if err != nil {
		return "", err
	}

	if nw_current == nil {
		nw_current = stats
	}

	var out string
	switch stats.typ {
	case "wifi":
		out = fmt.Sprintf("^i(%s)", xbm("net-wifi5"))
	case "ethernet":
	default:
		out = fmt.Sprintf("^i(%s)", xbm("net-wired2"))
	}

	out += " " + network_traffic(nw_current.rx, stats.rx, DOWNLOAD)
	out += " " + network_traffic(nw_current.tx, stats.tx, UPLOAD)

	nw_current = stats
	return out, nil
}

func network_device_bytes(dev, typ string) (int64, error) {
	data, err := ioutil.ReadFile("/sys/class/net/" + dev + "/statistics/" + typ + "_bytes")
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

func network_traffic(prev, next int64, typ string) string {
	nb := (next - prev) / KB / INTERVAL_SECS
	format := "%s ^i(" + xbm("arr_down") + ")^fg()"
	if typ == UPLOAD {
		format = "%s ^i(" + xbm("arr_up") + ")^fg()"
	}
	if nb > 0 {
		format = "^fg(" + nw_colors[typ] + ")" + format
	}

	var traffic string
	switch {
	case nb >= MB:
		traffic = fmt.Sprintf("%d MB", nb/MB)
	default:
		traffic = fmt.Sprintf("%d KB", nb)
	}

	return fmt.Sprintf(format, traffic)
}
