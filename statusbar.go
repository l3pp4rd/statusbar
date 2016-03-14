package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type element func() (string, error)

type statusbar struct {
	sync.Mutex
	sync.WaitGroup

	order    []string
	elements map[string]element
	results  map[string]string

	// configuration properties
	Dzen2 []string
	Gmail []gmailAccount
}

func (bar *statusbar) reg(name string, el element) {
	bar.elements[name] = el
	bar.order = append(bar.order, name)
}

func run(conf string) error {
	bar := &statusbar{
		results:  make(map[string]string),
		elements: make(map[string]element),
	}

	file, err := ioutil.ReadFile(conf)
	if err != nil {
		return fmt.Errorf("failed to read config file: %s - %s", conf, err)
	}

	if err := json.Unmarshal(file, bar); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %s - %s", conf, err)
	}

	bar.reg("keyboard", keyboard_layout)
	if len(bar.Gmail) > 0 {
		bar.reg("emails", unread_emails(bar.Gmail))
	}
	bar.reg("network", network_stats)
	bar.reg("temp", cpu_temp)
	bar.reg("power", power_battery)
	bar.reg("load", cpu_load)
	bar.reg("memory", memory_usage)
	bar.reg("date", date)

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

func (bar *statusbar) output() {
	cmd := exec.Command("dzen2", bar.Dzen2...)
	in := strings.Join(bar.iterate(), " ")
	log.Println(in)
	cmd.Stdin = bytes.NewBufferString(in)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("dzen2 command failed with err: %s and output: %s\n", err, string(out))
	}
}

func (bar *statusbar) iterate() []string {
	bar.Add(len(bar.elements))
	for nm, el := range bar.elements {
		go func(name string, elem element) {
			part, err := elem()
			if err != nil {
				log.Printf("failed to load: %s element, reason: %s\n", name, err)
			}
			bar.Lock()
			bar.results[name] = part
			bar.Unlock()
			bar.Done()
		}(nm, el)
	}
	bar.Wait()

	var res []string
	for _, ordered := range bar.order {
		res = append(res, bar.results[ordered])
	}
	return res
}
