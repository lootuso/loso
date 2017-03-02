package client

import (
	"net/url"
	"strings"
)

type Param map[string]string

func (p Param) ToString() string {
	var s string
	for k, v := range p {
		s = s + k + "=" + v + "&"
	}
	s = strings.TrimRight(s, "&")
	return url.QueryEscape(s)
}
