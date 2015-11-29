package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

var client = &http.Client{}
var last_iteration = EMAIL_PER_ITERATIONS
var last_counts []int

func get_unread(usr, psw string) (c int, err error) {
	req, err := http.NewRequest("GET", EMAIL_FEED, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(usr, psw)
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return c, fmt.Errorf(res.Status)
	}

	data := struct {
		Count int `xml:"fullcount"`
	}{}
	err = xml.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return
	}
	return data.Count, nil
}

func unread_emails(confPath string) element {
	return func() (string, error) {
		if last_iteration < EMAIL_PER_ITERATIONS {
			last_iteration++
			return unread_email_representation(last_counts), nil
		}

		file, err := ioutil.ReadFile(confPath)
		if err != nil {
			return "", err
		}
		var accounts []struct {
			Username, Password string
		}

		if err = json.Unmarshal(file, &accounts); err != nil {
			return "", err
		}

		var counts []int
		for _, acc := range accounts {
			cnt, err := get_unread(acc.Username, acc.Password)
			if err != nil {
				return "", err
			}
			counts = append(counts, cnt)
		}

		last_iteration = 1
		last_counts = counts
		return unread_email_representation(counts), nil
	}
}

func unread_email_representation(counts []int) string {
	var out string
	if len(counts) > 0 {
		out = "^i(" + xbm("mail") + ")"
		for _, c := range counts {
			if c > 0 {
				out += fmt.Sprintf(" ^fg(#dc322f)%d^fg()")
			} else {
				out += fmt.Sprintf(" %d", c)
			}
		}
	}
	return out
}
