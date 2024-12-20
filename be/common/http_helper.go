package common

import (
	"io"
	"net"
	"net/http"
	"time"
)

var DefaultHttpClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	},
}

func QuickGetHttp(method, url string, body io.Reader) (*http.Response, func(), error) {
	client, err := http.NewRequest(method, url, body)
	cleanup := func() {}
	if err != nil {
		return nil, cleanup, err
	}
	resp, err := DefaultHttpClient.Do(client)
	if err != nil {
		return nil, cleanup, err
	}
	cleanup = func() {
		if resp.Body == nil {
			return
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	return resp, cleanup, err
}
