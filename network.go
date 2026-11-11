package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	KB = 1024
	MB = KB * KB
	GB = MB * KB
)

type nw_direction struct {
	color, suffix string
}

var (
	nw_download = nw_direction{"^fg(#859900)", " ^i(" + xbm("arr_down") + ")^fg()"}
	nw_upload   = nw_direction{"^fg(#dc322f)", " ^i(" + xbm("arr_up") + ")^fg()"}

	nw_disconnected = "^fg(#dc322f)^i(" + xbm("disconnected") + ")^fg()"
	nw_wired        = "^i(" + xbm("net-wired") + ")"
	nw_wifi_full    = "^i(" + xbm("wifi-full") + ")"
	nw_wifi_mid     = "^i(" + xbm("wifi-mid") + ")"
	nw_wifi_low     = "^i(" + xbm("wifi-low") + ")"
)

type nw_device struct {
	name string
	wlan bool
	// stat file paths
	operstate, rx, tx string
}

type nw_stats struct {
	device *nw_device
	rx, tx int64
}

type Network struct {
	output
	devs    []*nw_device // ordered by priority
	current *nw_stats
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
	var dev *nw_device
	for _, d := range n.devs {
		dat, err := os.ReadFile(d.operstate)
		if err != nil {
			continue // unreadable status
		}
		if strings.TrimSpace(string(dat)) == "up" {
			dev = d
			break
		}
	}

	if dev == nil {
		n.current = nil
		n.set(nw_disconnected)
		return nil
	}

	stats := &nw_stats{device: dev}
	var err error
	if stats.rx, err = read_int(dev.rx); err != nil {
		return fmt.Errorf("stat downloaded bytes: %s", err)
	}
	if stats.tx, err = read_int(dev.tx); err != nil {
		return fmt.Errorf("stat uploaded bytes: %s", err)
	}

	// counters of different devices are not comparable
	if n.current == nil || n.current.device != dev {
		n.current = stats
	}

	out := nw_wired
	if dev.wlan {
		sig, err := signal_strength(dev.name)
		if err != nil {
			return err
		}

		// link quality is reported on a scale up to 70
		var color string
		switch {
		case sig >= 60:
			color, out = "#859900", nw_wifi_full
		case sig >= 30:
			color, out = "#b58900", nw_wifi_mid
		default:
			color, out = "#dc322f", nw_wifi_low
		}
		out = "^fg(" + color + ")" + out + " " + strconv.Itoa(min(100, sig*100/70)) + "%^fg()"
	}

	out += " " + network_traffic(n.current.rx, stats.rx, &nw_download)
	out += " " + network_traffic(n.current.tx, stats.tx, &nw_upload)

	n.set(out)
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

	var wireless []*nw_device
	for _, dev := range devs {
		p := filepath.Join("/sys/class/net", dev.Name())
		// filter out non devices
		if !file_exists(filepath.Join(p, "device")) {
			continue // not a physical device
		}

		d := &nw_device{
			name:      dev.Name(),
			wlan:      file_exists(filepath.Join(p, "wireless")),
			operstate: filepath.Join(p, "operstate"),
			rx:        filepath.Join(p, "statistics", "rx_bytes"),
			tx:        filepath.Join(p, "statistics", "tx_bytes"),
		}
		if d.wlan {
			wireless = append(wireless, d)
			continue
		}
		n.devs = append(n.devs, d)
	}
	// ethernet takes priority since usually it is faster and more reliable if wifi is also connected
	n.devs = append(n.devs, wireless...)

	return nil
}

func signal_strength(dev string) (int, error) {
	dat, err := os.ReadFile("/proc/net/wireless")
	if err != nil {
		return 0, fmt.Errorf("wifi signal strength: %v", err)
	}

	for ln := range strings.Lines(string(dat)) {
		if !strings.Contains(ln, dev) {
			continue
		}

		parts := strings.Fields(ln)
		if len(parts) < 3 {
			continue
		}

		sig, err := strconv.Atoi(strings.TrimRight(parts[2], "."))
		if err != nil {
			return 0, fmt.Errorf("wifi signal strength to int: %v", err)
		}

		return sig, nil
	}
	return 0, nil
}

func read_int(fp string) (int64, error) {
	data, err := os.ReadFile(fp)
	if err != nil {
		return 0, fmt.Errorf("could not read %s - %s", fp, err)
	}

	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

func network_traffic(prev, next int64, dir *nw_direction) string {
	nb := (next - prev) / INTERVAL_SECS

	var color string
	if nb > 0 {
		color = dir.color
	}

	n, unit := nb/KB, " KB"
	switch {
	case nb >= GB:
		n, unit = nb/GB, " GB"
	case nb >= MB:
		n, unit = nb/MB, " MB"
	}
	return color + strconv.FormatInt(n, 10) + unit + dir.suffix
}
