package main

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type gmailAccount struct {
	Username, Password string
}

type gmail struct {
	sync.WaitGroup
	sync.Mutex

	accounts []gmailAccount

	client *http.Client

	iteration int
	counts    []int
	results   map[string]int
}

func new_gmail_client(accounts []gmailAccount) *gmail {
	gm := &gmail{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		accounts:  accounts,
		iteration: EMAIL_PER_ITERATIONS - 1,
	}

	gm.client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 60 * time.Second,
		}).Dial,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout:   3 * time.Second,
		MaxIdleConnsPerHost:   len(gm.accounts),
		DisableCompression:    false,
		DisableKeepAlives:     false,
		ResponseHeaderTimeout: 3 * time.Second,
	}

	for _ = range gm.accounts {
		gm.counts = append(gm.counts, 0)
	}

	gm.results = make(map[string]int, len(gm.accounts))
	return gm
}

func (gm *gmail) fetch(usr, psw string) (c int, err error) {
	req, err := http.NewRequest("GET", EMAIL_FEED, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(usr, psw)
	res, err := gm.client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		res.Body.Close()
		return c, fmt.Errorf(res.Status)
	}

	data := struct {
		Count int `xml:"fullcount"`
	}{}
	err = xml.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		res.Body.Close()
		return
	}

	return data.Count, res.Body.Close()
}

func unread_emails(accounts []gmailAccount) element {
	gm := new_gmail_client(accounts)

	return func() (string, error) {
		if gm.iteration < EMAIL_PER_ITERATIONS {
			gm.iteration++
			return gm.result(), nil
		}

		gm.Add(len(gm.accounts))
		for _, acc := range gm.accounts {
			go func(u, p string) {
				c, err := gm.fetch(u, p)
				if err != nil {
					log.Printf("failed to fetch email count from: %s - %s\n", u, err)
					c = 0
				}
				gm.Lock()
				gm.results[u] = c
				gm.Unlock()
				gm.Done()
			}(acc.Username, acc.Password)
		}
		gm.Wait()

		var counts []int
		for _, acc := range gm.accounts {
			counts = append(counts, gm.results[acc.Username])
		}
		gm.counts = counts
		gm.iteration = 1
		return gm.result(), nil
	}
}

func (gm *gmail) result() string {
	var out string
	if len(gm.counts) > 0 {
		out = "^i(" + xbm("mail") + ")"
		for _, c := range gm.counts {
			if c > 0 {
				out += fmt.Sprintf(" ^fg(#dc322f)%d^fg()", c)
			} else {
				out += fmt.Sprintf(" %d", c)
			}
		}
	}
	return out
}
