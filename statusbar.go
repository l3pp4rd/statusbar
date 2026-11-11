package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

type element interface {
	value() string
}

// output holds the latest rendered value of an element,
// it is written by the element poller and read by the statusbar loop
type output struct {
	p atomic.Pointer[string]
}

func (o *output) set(s string) {
	o.p.Store(&s)
}

func (o *output) value() string {
	if s := o.p.Load(); s != nil {
		return *s
	}
	return ""
}

// elements start their pollers, which run for the lifetime
// of the process, so it must be called only once
func elements() []element {
	return []element{
		keyboard(),
		network(),
		cpu_temp(),
		power(),
		cpu_load(),
		memory_usage(),
		date(),
	}
}

type statusbar struct {
	elements []element

	// configuration properties
	Dzen2 []string
}

func run(ctx context.Context, conf string, els []element) error {
	bar := statusbar{elements: els}
	file, err := os.ReadFile(conf)
	if err != nil {
		return fmt.Errorf("failed to read config file: %s - %s", conf, err)
	}

	if err := json.Unmarshal(file, &bar); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %s - %s", conf, err)
	}

	cmd := exec.CommandContext(ctx, "dzen2", bar.Dzen2...)
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
			select {
			case <-ctx.Done():
				return
			default:
			}
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
	res := make([]string, 0, len(bar.elements))
	for _, el := range bar.elements {
		res = append(res, el.value())
	}
	return res
}
