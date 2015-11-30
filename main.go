package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	INTERVAL_SECS = 1

	CPU_LOAD_FILE = "/proc/loadavg"

	EMAIL_FEED           = "https://mail.google.com/a/gmail.com/feed/atom"
	EMAIL_PER_ITERATIONS = 30 // every 30 seconds if interval is 1s

	LOG_FILE = "/tmp/statusbar.log"
	XBM_DIR  = "/tmp/statusbar_xbm"
)

func xbm(name string) string {
	return XBM_DIR + "/" + name + ".xbm"
}

func main() {
	var emailConfPath string
	if len(os.Args) > 1 {
		emailConfPath = os.Args[1]
	}
	f, err := os.OpenFile(LOG_FILE, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %s - %v", LOG_FILE, err)
	}
	defer f.Close()
	log.SetOutput(f)

	if err := init_assets(); err != nil {
		log.Fatalf("asset initialization failed: %s", err)
	}

	bar := &statusbar{
		results:  make(map[string]string),
		elements: make(map[string]element),
	}

	bar.reg("keyboard", keyboard_layout)
	if len(emailConfPath) > 0 {
		bar.reg("emails", unread_emails(emailConfPath))
	}
	bar.reg("network", network_stats)
	bar.reg("temp", cpu_temp)
	bar.reg("power", power_battery)
	bar.reg("load", cpu_load)
	bar.reg("memory", memory_usage)
	bar.reg("date", date)

	for {
		fmt.Println(strings.Join(bar.run(), " "))
		time.Sleep(time.Second * INTERVAL_SECS)
	}
}

func init_assets() error {
	_, err := exec.Command("mkdir", "-p", XBM_DIR).Output()
	if err != nil {
		return err
	}
	for p, f := range _bindata {
		loc := XBM_DIR + strings.Replace(p, "xbm/", "/", 1)
		if _, err := os.Stat(loc); err == nil {
			continue
		}

		asset, err := f()
		if err != nil {
			return fmt.Errorf("asset %s can't read by error: %v", p, err)
		}

		file, err := os.Create(loc)
		if err != nil {
			return fmt.Errorf("failed to open asset: %s file for writing: %s", loc, err)
		}

		if _, err = file.Write(asset.bytes); err != nil {
			file.Close()
			return fmt.Errorf("failed to write asset to file: %s, because: %s", loc, err)
		}
		file.Close()
	}
	return nil
}
