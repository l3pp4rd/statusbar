package main

import (
	"fmt"
	"log"
	"time"
)

type Date struct {
	val string
}

func (k *Date) value() string {
	return k.val
}

func date() element {
	e := &Date{}
	go func() {
		for {
			if val, err := e.read(); err == nil {
				e.val = val
			} else {
				log.Printf("could not read date: %v", err)
			}
			time.Sleep(time.Second)
		}
	}()
	return e
}

func (k *Date) read() (string, error) {
	localTZ, err := time.LoadLocation("Europe/Vilnius")
	if err != nil {
		return "", err
	}
	extraTZ, err := time.LoadLocation("Europe/London")
	if err != nil {
		return "", err
	}

	local := time.Now().In(localTZ).Format("Mon _2 Jan 15:04")
	extra := "UK " + time.Now().In(extraTZ).Format("15:04")

	return fmt.Sprintf("^fg(white)%s ^i(%s) ^fg()%s", local, xbm("clock2"), extra), nil
}
