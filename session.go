package main

import (
	"encoding/json"
	"io"
	"net/http"
)

// Request reflects the [http.Request] params.
type Request struct {
	Host       string              `json:"host"`
	Scheme     string              `json:"scheme"`
	Method     string              `json:"method"`
	UrlPath    string              `json:"url"`
	Headers    map[string][]string `json:"headers"`
	FormValues map[string][]string `json:"form"`
	RawQuery   string              `json:"qs"`
}

// Response reflects the [http.Response] data.
type Response struct {
	Status    string              `json:"status"`
	Proto     string              `json:"proto"`
	Headers   map[string][]string `json:"headers"`
	BodyLines []string            `json:"body"`
}

// Session reflect current state: some stats, request settings and last response (with data).
type Session struct {
	ReqCount int      `json:"reqCount"`
	ResTime  string   `json:"resTime"`
	Request  Request  `json:"req"`
	Response Response `json:"res"`
}

// Create a new session.
func NewSession(rq *http.Request, rs *http.Response, rqc int, rqt string, rqf map[string][]string, rsb []string) (*Session, error) {
	req := Request{Headers: make(map[string][]string), FormValues: make(map[string][]string)}
	res := Response{Headers: make(map[string][]string)}

	// populate Request with [http.Request] data
	if rq != nil {
		req.Method = rq.Method
		req.Scheme = rq.URL.Scheme
		req.Host = rq.URL.Host
		req.UrlPath = rq.URL.Path
		for k, v := range rq.Header {
			req.Headers[k] = v
		}
		req.RawQuery = rq.URL.RawQuery
		req.FormValues = rqf
	}

	// populate Response with [http.Response] data

	if rs != nil {
		res.Status = rs.Status
		res.Proto = rs.Proto
		for k, v := range rs.Header {
			res.Headers[k] = v
		}
		res.BodyLines = rsb
	}
	return &Session{ReqCount: rqc, ResTime: rqt, Request: req, Response: res}, nil
}

// Save the session.
func (s *Session) Save(o io.Writer) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	_, err = o.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// Load the session.
func (s *Session) Load(r io.Reader) error {

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, s)
	if err != nil {
		return err
	}

	return nil
}
