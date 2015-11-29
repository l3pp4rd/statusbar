package main

import (
	"fmt"
	"time"
)

func date() (string, error) {
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
