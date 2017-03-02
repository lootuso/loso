package client

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func Request(urlStr string, data Param, method string, h http.Header) ([]byte, error) {
	method = strings.ToUpper(method)
	var body io.Reader = nil
	if method == "" {
		method = "GET"
	}
	if method == "POST" {
		body = strings.NewReader(data.ToString())
	} else if method == "GET" {
		if strings.Index(urlStr, "?") == -1 {
			urlStr = urlStr + "?" + data.ToString()
		} else {
			urlStr = urlStr + "&" + data.ToString()
		}
	}
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if h != nil {
		req.Header = h
	}else{
		req.Header.Set("Accept","text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Referer",urlStr)
		req.Header.Set("Upgrade-Insecure-Requests","1")
		req.Header.Set("User-Agent","Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36")
	}
	if method == "POST"{
		req.Header.Set("Content-Type","application/x-www-form-urlencoded")
	}
	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}
