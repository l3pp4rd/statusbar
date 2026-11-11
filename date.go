package main

import (
	"log"
	"time"
)

type Date struct {
	output
	localTZ, extraTZ *time.Location
}

func date() element {
	e := &Date{}
	// time zones are loaded once, since every load reads and parses the zoneinfo file
	var err error
	if e.localTZ, err = time.LoadLocation("Europe/Vilnius"); err != nil {
		log.Printf("could not load local time zone: %v", err)
		return e
	}
	if e.extraTZ, err = time.LoadLocation("Europe/London"); err != nil {
		log.Printf("could not load extra time zone: %v", err)
		return e
	}
	go func() {
		for {
			e.set(e.read())
			time.Sleep(time.Second)
		}
	}()
	return e
}

func (k *Date) read() string {
	now := time.Now()
	local := now.In(k.localTZ).Format("Mon _2 Jan 15:04")
	extra := now.In(k.extraTZ).Format("15:04")

	return "^fg(white)" + local + " ^i(" + xbm("clock2") + ") ^fg()UK " + extra
}
