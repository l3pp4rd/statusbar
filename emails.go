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

	counts []int
}

func emails(accounts []gmailAccount) element {
	e := &gmail{
		accounts: accounts,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	e.client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 3 * time.Minute,
		}).Dial,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout:   3 * time.Second,
		MaxIdleConnsPerHost:   len(e.accounts),
		DisableCompression:    false,
		DisableKeepAlives:     false,
		ResponseHeaderTimeout: 3 * time.Second,
	}

	e.counts = make([]int, len(e.accounts))

	go func() {
		for {
			e.read()
			time.Sleep(time.Second * 15)
		}
	}()
	return e
}

func (g *gmail) fetch(usr, psw string) (c int, err error) {
	req, err := http.NewRequest("GET", EMAIL_FEED, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(usr, psw)
	res, err := g.client.Do(req)
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

func (g *gmail) read() {
	g.Add(len(g.accounts))
	for i, acc := range g.accounts {
		go func(u, p string, n int) {
			if c, err := g.fetch(u, p); err != nil {
				log.Printf("failed to fetch email count from: %s - %s\n", u, err)
			} else {
				g.Lock()
				g.counts[n] = c
				g.Unlock()
			}
			g.Done()
		}(acc.Username, acc.Password, i)
	}
	g.Wait()
}

func (g *gmail) value() string {
	var out string
	if len(g.counts) > 0 {
		out = "^i(" + xbm("mail") + ")"
		for _, c := range g.counts {
			if c > 0 {
				out += fmt.Sprintf(" ^fg(#dc322f)%d^fg()", c)
			} else {
				out += fmt.Sprintf(" %d", c)
			}
		}
	}
	return out
}
