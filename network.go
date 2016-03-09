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
		return 0, fmt.Errorf("wifi signal strength nmcli: %s", err)
	}
	for _, ln := range strings.Split(string(lines), "\n") {
		parts := strings.Split(strings.TrimSpace(ln), ":")
		if strings.Index(s.ssid, parts[0]) == -1 {
			continue
		}

		sig, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("wifi signal strength to int: %s", err)
		}

		return sig, nil
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
			continue // making sure has number of sections required
		case parts[1] == "bridge":
			continue // filter bridge interfaces
		case parts[1] == "loopback":
			continue // filter loopback interfaces
		case parts[2] != "connected":
			continue // filter out divices which are not connected
		case strings.Index(parts[0], "dock") == 0:
			continue // filter out docker connections
		case strings.Index(parts[0], "veth") == 0:
			continue // filter out docker connections
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
		return "", fmt.Errorf("stat downloaded bytes: %s", err)
	}
	stats.tx, err = network_device_bytes(stats.device, UPLOAD)
	if err != nil {
		return "", fmt.Errorf("stat uploaded bytes: %s", err)
	}

	if nw_current == nil {
		nw_current = stats
	}

	var out string
	switch stats.typ {
	case "wifi":
		sig, err := stats.signal_strength()
		if err != nil {
			return out, err
		}

		switch {
		case sig >= 70:
			out = fmt.Sprintf("^i(%s)", xbm("wifi-full"))
		case sig >= 40:
			out = fmt.Sprintf("^i(%s)", xbm("wifi-mid"))
		default:
			out = fmt.Sprintf("^i(%s)", xbm("wifi-low"))
		}
	case "ethernet":
		out = fmt.Sprintf("^i(%s)", xbm("net-wired"))
	default:
		out = fmt.Sprintf("^i(%s)", xbm("net-wired"))
	}

	out += " " + network_traffic(nw_current.rx, stats.rx, DOWNLOAD)
	out += " " + network_traffic(nw_current.tx, stats.tx, UPLOAD)

	nw_current = stats
	return out, nil
}

func network_device_bytes(dev, typ string) (int64, error) {
	fp := "/sys/class/net/" + dev + "/statistics/" + typ + "_bytes"
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return 0, fmt.Errorf("could not read %s - %s", fp, err)
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

	return fmt.Sprintf(format, fmt.Sprintf("%d KB", nb))
}
