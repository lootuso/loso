package client

import (
	"net/http"
	"io/ioutil"
	"strings"
)

func Get(url string,params Param)([]byte,error ) {
	if params != nil{
		if strings.Index(url,"?") == -1 {
			url = url + "?" + params.ToString()
		}else{
			url = url + "&" + params.ToString()
		}
	}
	res,err := http.Get(url)
	if err != nil{
		return nil,err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}