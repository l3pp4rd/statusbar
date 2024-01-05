package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type element interface {
	value() string
}

type statusbar struct {
	elements []element

	// configuration properties
	Dzen2 []string
	Gmail []gmailAccount
}

func run(conf string) error {
	var bar statusbar
	file, err := os.ReadFile(conf)
	if err != nil {
		return fmt.Errorf("failed to read config file: %s - %s", conf, err)
	}

	if err := json.Unmarshal(file, &bar); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %s - %s", conf, err)
	}

	bar.elements = append(bar.elements, keyboard())
	if len(bar.Gmail) > 0 {
		bar.elements = append(bar.elements, emails(bar.Gmail))
	}
	bar.elements = append(bar.elements, network())
	bar.elements = append(bar.elements, cpu_temp())
	bar.elements = append(bar.elements, power())
	bar.elements = append(bar.elements, cpu_load())
	bar.elements = append(bar.elements, memory_usage())
	bar.elements = append(bar.elements, date())

	cmd := exec.Command("dzen2", bar.Dzen2...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %s", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start dzen2 command: %s", err)
	}

	// run the iteration loop
	go func() {
		for {
			if _, e := stdin.Write([]byte(strings.Join(bar.iterate(), " ") + "\n")); e != nil {
				log.Printf("probably the pipe closed: %s", e)
				break
			}
			time.Sleep(time.Second * INTERVAL_SECS)
		}
	}()

	return cmd.Wait()
}

func (bar *statusbar) iterate() []string {
	var res []string
	for _, el := range bar.elements {
		res = append(res, el.value())
	}
	return res
}
