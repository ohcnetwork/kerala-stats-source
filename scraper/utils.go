package scraper

import (
	"io"
	"net/http"
	"strings"
	"sync"
)

var sessid = ""
var mutex = sync.Mutex{}

func makeRequest(source string, referer string) (io.ReadCloser, error) {
	client := &http.Client{}
	var req *http.Request
	req, _ = http.NewRequest("GET", source, nil)
	req.Host = "dashboard.kerala.gov.in"
	req.Header.Set("Origin", "https://dashboard.kerala.gov.in")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:75.0) Gecko/20100101 Firefox/75.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Referer", referer)
	req.Header.Set("Connection", "keep-alive")
	if sessid != "" {
		req.Header.Set("Cookie", sessid)
	}
	res, err := client.Do(req)
	if err != nil {
		return &io.PipeReader{}, err
	}
	if res.Header.Get("Set-Cookie") != "" {
		mutex.Lock()
		sessid = strings.Split(res.Header.Get("Set-Cookie"), ";")[0]
		mutex.Unlock()
	}
	return res.Body, nil
}
