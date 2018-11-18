package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/davecgh/go-spew/spew"
	//"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

type noForwarder struct{}

var NoForwarder = &noForwarder{}

func (_ *noForwarder) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Del("X-Forwarded-For")
	return http.DefaultTransport.RoundTrip(r)
}

func main() {
	proxy := &httputil.ReverseProxy{
		Transport:      NoForwarder,
		Director:       handleReq,
		ModifyResponse: handleResp,
	}
	fmt.Println("Started")
	http.ListenAndServe(":8000", proxy)
}

func handleReq(r *http.Request) {
	r.URL.Scheme = "https"
	r.URL.Host, r.URL.Path = getRedirect(r)
	fmt.Println(r.URL.Path)
	newBody, body := copyReader(r.Body)
	r.Body = newBody
	ctx := context.WithValue(r.Context(), "req.body", body)
	nr := r.WithContext(ctx)
	*r = *nr
}

func copyReader(r io.ReadCloser) (io.ReadCloser, []byte) {
	if r == nil {
		return r, []byte{}
	}
	body, err := ioutil.ReadAll(r)
	fmt.Printf("%+v\n", string(body))
	if err != nil {
		panic(err)
	}
	return ioutil.NopCloser(bytes.NewBuffer(body)), body
}

func handleResp(r *http.Response) error {
	newBody, body := copyReader(r.Body)
	reqBody := r.Request.Context().Value("req.body").([]byte)
	go StoreReq(r, body, reqBody)
	r.Body = injection(r, newBody)
	return nil
}

func StoreReq(r *http.Response, respBody, reqBody []byte) {
	resp := decodeFull(respBody)
	req := decodeFull(reqBody)
	data := NewData(r, resp, req)
	f, err := os.Create("logs/" + r.Request.Method + "--" + strings.Join(strings.Split(r.Request.URL.Path, "/"), ".") + ".json")
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(data)
	if err != nil {
		panic(err)
	}
}

func decodeFull(data []byte) interface{} {
	dataBytes := decodeToBytes(data)
	if !json.Valid(dataBytes) {
		return string(dataBytes)
	}
	d := make(map[string]interface{})
	err := json.Unmarshal(dataBytes, &d)
	if err != nil {
		spew.Dump(err)
		return string(dataBytes)
	}
	return d
}

func decodeToBytes(d []byte) []byte {
	buf := bytes.NewBuffer(d)
	reader, err := gzip.NewReader(buf)
	if err != nil {
		return d
	}
	ret, err := ioutil.ReadAll(reader)
	if err != nil {
		return d
	}
	return ret
}
