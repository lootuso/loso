package client

import (
	"net/http"
	"net/url"
	"io/ioutil"
)

func Post(url string,data url.Values)([]byte,error){
	res,err := http.PostForm(url,data)
	if err != nil{
		return nil,err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}
