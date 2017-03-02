package client

import "net/url"

type Param map[string]string

func (p Param) ToString() string {
	var s string
	for k, v := range p {
		s = s + k + "=" + v + "&"
	}
	return url.QueryEscape(s)
}
