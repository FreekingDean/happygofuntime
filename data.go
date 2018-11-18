package main

import (
	"net/http"
	"net/url"
)

type Request struct {
	Headers http.Header
	Method  string
	URL     *url.URL
	Body    interface{}
}
type Response struct {
	Headers http.Header
	Status  string
	Body    interface{}
}
type Data struct {
	Request  Request
	Response Response
}

func NewData(r *http.Response, respBody interface{}, reqBody interface{}) *Data {
	return &Data{
		Request: Request{
			Headers: r.Request.Header,
			Method:  r.Request.Method,
			URL:     r.Request.URL,
			Body:    reqBody,
		},
		Response: Response{
			Headers: r.Header,
			Status:  r.Status,
			Body:    respBody,
		},
	}
}
