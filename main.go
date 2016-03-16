package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	INTERVAL_SECS = 1

	CPU_LOAD_FILE = "/proc/loadavg"

	EMAIL_FEED           = "https://mail.google.com/a/gmail.com/feed/atom"
	EMAIL_PER_ITERATIONS = 30 // every 30 seconds if interval is 1s

	XBM_DIR = "/tmp/statusbar_xbm"
)

func xbm(name string) string {
	return XBM_DIR + "/" + name + ".xbm"
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("expected configuration file path as first the argument")
	}

	if err := init_assets(); err != nil {
		log.Fatalf("asset initialization failed: %s", err)
	}

	if err := run(os.Args[1]); err != nil {
		log.Fatalf("statusbar failed: %s", err)
	}
}

func init_assets() error {
	_, err := exec.Command("mkdir", "-p", XBM_DIR).Output()
	if err != nil {
		return fmt.Errorf("could not create dir: %s - %s", XBM_DIR, err)
	}
	for p, f := range _bindata {
		loc := XBM_DIR + strings.Replace(p, "xbm/", "/", 1)
		_, err := os.Stat(loc)
		switch {
		case err == nil:
			continue
		case !os.IsNotExist(err):
			return fmt.Errorf("stat asset %s - %s", loc, err)
		}

		asset, err := f()
		if err != nil {
			return fmt.Errorf("asset %s can't read: %s", p, err)
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
