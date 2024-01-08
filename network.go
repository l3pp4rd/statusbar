package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	KB       = 1024
	MB       = KB * KB
	GB       = MB * KB
	DOWNLOAD = "rx"
	UPLOAD   = "tx"
)

var nw_colors = map[string]string{
	DOWNLOAD: "#859900",
	UPLOAD:   "#dc322f",
}

type nw_stats struct {
	device string
	wlan   bool
	rx, tx int64
}

type Network struct {
	val      string
	ethernet []string
	wireless []string
	current  *nw_stats
}

func (n *Network) value() string {
	return n.val
}

func network() element {
	n := &Network{}
	if err := n.devices(); err != nil {
		log.Printf("failed to read network devices: %v\n", err)
		return n
	}
	go func() {
		for {
			if err := n.stats(); err != nil {
				log.Printf("could not read network stats: %v\n", err)
			}
			time.Sleep(time.Second)
		}
	}()
	return n
}

func (n *Network) stats() error {
	var stats *nw_stats
	// ethernet takes priority since usually it is faster and more reliable if wifi is also connected
	for _, dev := range append(n.ethernet, n.wireless...) {
		dat, err := os.ReadFile(filepath.Join("/sys/class/net", dev, "operstate"))
		if err != nil {
			continue // unreadable status
		}

		if strings.TrimSpace(string(dat)) != "up" {
			continue
		}

		var wlan bool
		for _, d := range n.wireless {
			if d == dev {
				wlan = true
				break
			}
		}

		stats = &nw_stats{
			device: dev,
			wlan:   wlan,
		}
		break
	}

	if stats == nil {
		n.current = nil
		n.val = fmt.Sprintf("^fg(#dc322f)^i(%s)^fg()", xbm("disconnected"))
		return nil
	}

	var err error
	stats.rx, err = network_device_bytes(stats.device, DOWNLOAD)
	if err != nil {
		return fmt.Errorf("stat downloaded bytes: %s", err)
	}
	stats.tx, err = network_device_bytes(stats.device, UPLOAD)
	if err != nil {
		return fmt.Errorf("stat uploaded bytes: %s", err)
	}

	if n.current == nil {
		n.current = stats
	}

	var out string
	if stats.wlan {
		sig, err := stats.signal_strength()
		if err != nil {
			return err
		}

		switch {
		case sig >= 60:
			out = fmt.Sprintf("^i(%s)", xbm("wifi-full"))
		case sig >= 30:
			out = fmt.Sprintf("^i(%s)", xbm("wifi-mid"))
		default:
			out = fmt.Sprintf("^i(%s)", xbm("wifi-low"))
		}
	} else {
		out = fmt.Sprintf("^i(%s)", xbm("net-wired"))
	}

	out += " " + network_traffic(n.current.rx, stats.rx, DOWNLOAD)
	out += " " + network_traffic(n.current.tx, stats.tx, UPLOAD)

	n.val = out
	n.current = stats
	return nil
}

// see https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-class-net
// and http://unix.stackexchange.com/questions/40560/how-to-know-if-a-network-interface-is-tap-tun-bridge-or-physical
func (n *Network) devices() error {
	devs, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return err
	}

	for _, dev := range devs {
		d := filepath.Base(dev.Name())
		// filter out non devices
		if !file_exists(filepath.Join("/sys/class/net", d, "device")) {
			continue // not a physical device
		}

		// maybe wireless
		if file_exists(filepath.Join("/sys/class/net", d, "wireless")) {
			n.wireless = append(n.wireless, d)
			continue
		}

		// otherwise ethernet
		n.ethernet = append(n.ethernet, d)
	}

	return nil
}

var trimSpaces = regexp.MustCompile("\\s+")

func (s *nw_stats) signal_strength() (int, error) {
	dat, err := os.ReadFile("/proc/net/wireless")
	if err != nil {
		return 0, fmt.Errorf("wifi signal strength: %v", err)
	}

	for _, ln := range strings.Split(string(dat), "\n") {
		if strings.Index(ln, s.device) == -1 {
			continue
		}

		ln = trimSpaces.ReplaceAllString(ln, " ")
		parts := strings.Split(strings.TrimSpace(ln), " ")

		sig, err := strconv.Atoi(strings.TrimRight(parts[2], "."))
		if err != nil {
			return 0, fmt.Errorf("wifi signal strength to int: %v", err)
		}

		return sig, nil
	}
	return 0, nil
}

func network_device_bytes(dev, typ string) (int64, error) {
	fp := "/sys/class/net/" + dev + "/statistics/" + typ + "_bytes"
	data, err := os.ReadFile(fp)
	if err != nil {
		return 0, fmt.Errorf("could not read %s - %s", fp, err)
	}

	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

func network_traffic(prev, next int64, typ string) string {
	nb := (next - prev) / INTERVAL_SECS
	format := "%s ^i(" + xbm("arr_down") + ")^fg()"
	if typ == UPLOAD {
		format = "%s ^i(" + xbm("arr_up") + ")^fg()"
	}
	if nb > 0 {
		format = "^fg(" + nw_colors[typ] + ")" + format
	}

	gbs := nb / GB
	if gbs > 0 {
		return fmt.Sprintf(format, fmt.Sprintf("%d GB", gbs))
	}
	mbs := nb / MB
	if mbs > 0 {
		return fmt.Sprintf(format, fmt.Sprintf("%d MB", mbs))
	}
	return fmt.Sprintf(format, fmt.Sprintf("%d KB", nb/KB))
}
