package utils

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

var webcl *http.Client

func TryConnect(host string, port int) bool {
	dest := host + `:` + strconv.Itoa(port)
	Debug(`Attempting connection to %s`, dest)
	if c, err := net.Dial("tcp", dest); err == nil {
		defer func(c net.Conn) {
			_ = c.Close()
		}(c)
		Debug(`  Success.`)
		return true
	} else {
		Debug(`  Failure.`)
		return false
	}
}

func WaitUntilPortAttached(port int, message string) {
	if message == "" {
		fmt.Printf(`Waiting for port %d`, port)
	} else {
		fmt.Print(message)
	}

	for !TryConnect(`localhost`, port) {
		fmt.Print(`.`)
		time.Sleep(time.Second)
	}
}

func WaitUntilResponseSuccess(url string, message string) {
	if message == "" {
		fmt.Printf(`Waiting until %s responds successfully`, url)
	} else {
		fmt.Print(message)
	}

	for !RespondsWithSuccess(url) {
		fmt.Print(`.`)
		time.Sleep(time.Second)
	}
}

func RespondsWithSuccess(url string) bool {
	if res, err := client().Get(url); err == nil {
		Debug(`status %d`+"\n", res.StatusCode)
		if res.StatusCode < 400 {
			return true
		}
	} else {
		Debug(`error %v`, err)
	}
	return false
}

func client() *http.Client {
	if webcl == nil {
		return &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	} else {
		return webcl
	}
}
