package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	INTERVAL_SECS = 1

	EMAIL_FEED = "https://mail.google.com/a/gmail.com/feed/atom"

	XBM_DIR = "/tmp/statusbar_xbm"
)

func xbm(name string) string {
	return XBM_DIR + "/" + name + ".xbm"
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("expected configuration file path as first the argument\n")
	}

	if err := init_assets(); err != nil {
		log.Fatalf("asset initialization failed: %s\n", err)
	}

	conf := os.Args[1]
	for {
		ctx, cancel := context.WithCancel(context.Background())

		// watch for xrandr changes and cancel context when detected
		go watchXrandr(ctx, cancel)

		err := run(ctx, conf)
		cancel()
		if err != nil && ctx.Err() == nil {
			log.Fatalf("statusbar failed: %s\n", err)
		}
		log.Println("screen configuration changed, restarting...")
	}
}

func init_assets() error {
	_, err := exec.Command("mkdir", "-p", XBM_DIR).Output()
	if err != nil {

		return fmt.Errorf("could not create dir: %s - %s", XBM_DIR, err)
	}
	files, err := iconFiles.ReadDir("xbm")
	if err != nil {

		return fmt.Errorf("could not read embedded icons: %s", err)
	}
	for _, f := range files {
		loc := XBM_DIR + "/" + f.Name()
		_, err := os.Stat(loc)
		switch {
		case err == nil:
			continue
		case !os.IsNotExist(err):
			return fmt.Errorf("stat asset %s - %s", loc, err)
		}

		asset, err := iconFiles.ReadFile("xbm/" + f.Name())
		if err != nil {
			return fmt.Errorf("asset can't read: %s", err)
		}

		file, err := os.Create(loc)
		if err != nil {
			return fmt.Errorf("failed to open asset: %s file for writing: %s", loc, err)
		}

		if _, err = file.Write(asset); err != nil {
			errClose := file.Close()
			if errClose != nil {
				return errClose
			}
			return fmt.Errorf("failed to write asset to file: %s, because: %s", loc, err)
		}
		err = file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func xrandrOutput() string {
	out, err := exec.Command("xrandr").Output()
	if err != nil {
		log.Printf("xrandr failed: %s", err)
		return ""
	}
	return string(out)
}

func watchXrandr(ctx context.Context, cancel context.CancelFunc) {
	prev := xrandrOutput()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cur := xrandrOutput()
			if cur != "" && cur != prev {
				log.Println("xrandr change detected")
				cancel()
				return
			}
		}
	}
}

func file_exists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}
