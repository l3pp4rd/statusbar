package main

import (
	"log"
	"sync"
)

type element func() (string, error)

type statusbar struct {
	sync.Mutex
	sync.WaitGroup

	order    []string
	elements map[string]element
	results  map[string]string
}

func (bar *statusbar) reg(name string, el element) {
	bar.elements[name] = el
	bar.order = append(bar.order, name)
}

func (bar *statusbar) run() []string {
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
