package main

import (
	"bytes"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type matchFunc func(*http.Response) bool
type bodyFunc func(*http.Response, io.Reader) []byte

type injector struct {
	injector bodyFunc
	matcher  matchFunc
}

var injectors = []injector{
//	injector{
//		matcher:  matchTitleServers,
//		injector: injectTitleServers,
//	},
}

func injection(r *http.Response, body io.ReadCloser) io.ReadCloser {
	for _, injector := range injectors {
		if injector.matcher(r) {
			newBody := injector.injector(r, body)
			buf := bytes.NewBuffer(newBody)
			return ioutil.NopCloser(buf)
		}
	}
	return body
}

func matchTitleServers(r *http.Response) bool {
	if r.Request.URL.Host == "titlestorage.bethesda.net" &&
		r.Request.URL.Path == "/public/957a9cef-bd8d-480d-8595-2305b6d19555/pc/1/prodpc01/bf7036c0fccd6d358869366d03d3e75c" {
		return true
	}
	return false
}

func injectTitleServers(r *http.Response, body io.Reader) []byte {
	q := &struct {
		Global *struct {
			Services []*struct {
				Name   string `json:"name"`
				Pubkey string `json:"pubkey"`
				URL    string `json:"url"`
			} `json:"services"`
		} `json:"global"`
		Regions []*struct {
			ID       int    `json:"id"`
			PingURL  string `json:"ping_url"`
			Services []*struct {
				Name   string `json:"name"`
				Pubkey string `json:"pubkey"`
				URL    string `json:"url"`
			} `json:"services"`
		} `json:"regions"`
	}{}
	data, err := ioutil.ReadAll(body)
	if err != nil {
		spew.Dump(err)
		return []byte{}
	}
	err = json.Unmarshal(data, &q)
	if err != nil {
		spew.Dump(err)
		return data
	}
	for i, service := range q.Global.Services {
		path := service.Name + "-global"
		addRedirect("/"+path, service.URL)
		q.Global.Services[i].URL = "https://api.bethesda.net/" + path
	}
	for i, region := range q.Regions {
		for j, service := range region.Services {
			path := service.Name + "-" + strconv.Itoa(region.ID)
			addRedirect("/"+path, service.URL)
			q.Regions[i].Services[j].URL = "https://api.bethesda.net/" + path
		}
	}

	spew.Dump(q)
	newData, err := json.Marshal(q)
	if err != nil {
		return data
	}
	return newData
}
