package main

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

// Test WriteCloser which is mimic the io.WriteCloser, it saves all written data to local var.
type testWriteCloser struct {
	data     []byte
	isClosed bool
}

func (wc *testWriteCloser) Write(p []byte) (n int, err error) {
	wc.data = p
	return len(p), nil
}
func (wc *testWriteCloser) Close() error {
	wc.isClosed = true
	return nil
}
func (wc *testWriteCloser) Contains(b []byte) bool {
	return bytes.Contains(wc.data, b)
}

func TestSession(t *testing.T) {
	var rq *http.Request
	var rs *http.Response

	rqf := map[string][]string{}
	rsb := []string{}

	t.Run("create new session with empty request and response", func(t *testing.T) {
		_, err := NewSession(rq, rs, 10, "10.57µs", rqf, rsb)
		if err != nil {
			t.Fatalf("cannot create new session, error: %s", err)
		}
	})

	rq = &http.Request{
		Method: "GET",
		URL: &url.URL{
			Host:     "localhost",
			Path:     "/api/endpoint",
			RawQuery: "order_by=name&limit=10",
		},
		Header: make(http.Header),
	}
	rq.Header.Set("content-type", "application/json")
	rq.Header.Set("auth-token", "token")

	rs = &http.Response{
		Status: "200 OK",
		Proto:  "HTTP/1.1",
		Header: make(http.Header),
	}
	rs.Header.Set("connection", "keep-alive")
	rs.Header.Set("content-length", "34")
	rs.Header.Set("content-type", "application/json; charset=utf-8")
	rs.Header.Set("date", "Date:  Wed, 20 Mar 2024 09:11:23 GMT")

	rsb = []string{"{", `"msg": "hello!",`, `"id": 455`, "}"}

	t.Run("create new session filled request and response", func(t *testing.T) {
		s, err := NewSession(rq, rs, 10, "10.57µs", rqf, rsb)
		if err != nil {
			t.Fatalf("cannot create new session, error: %s", err)
		}
		if s.ReqCount != 10 {
			t.Errorf("expected req count: 10, got: %d", s.ReqCount)
		}
		if s.Response.Status != "200 OK" {
			t.Errorf("expected res status: 200 OK, got: %s", s.Response.Status)
		}
		v, ok := s.Response.Headers["Content-Type"]
		if !ok || slices.Compare(v, []string{"application/json; charset=utf-8"}) != 0 {
			t.Errorf(
				`expected: res header Content-Type="application/json; charset=utf-8", got: %#v`, v,
			)
		}
		if slices.Compare(s.Response.BodyLines, rsb) != 0 {
			t.Errorf("expected res body lines %s != %s", rsb, s.Response.BodyLines)
		}

	})

	buf := testWriteCloser{}

	t.Run("save", func(t *testing.T) {
		s, _ := NewSession(rq, rs, 10, "10.57µs", rqf, rsb)

		err := s.Save(&buf)
		if err != nil {
			t.Errorf("there is an error during the Session.Save(): %s", err)
		}

		if !buf.isClosed {
			t.Errorf("the Close() was not invoked")
		}

		if !buf.Contains([]byte(`"url":"/api/endpoint","headers":{"Auth-Token":["token"]`)) {
			t.Errorf("not found expected data, saved: %s", buf.data)
		}
	})

	t.Run("load", func(t *testing.T) {
		var rq *http.Request
		var rs *http.Response

		rqf := map[string][]string{}
		rsb := []string{"{", `"msg": "hello!",`, `"id": 455`, "}"}

		s, _ := NewSession(rq, rs, 10, "10.57µs", rqf, []string{})
		r := io.NopCloser(strings.NewReader(string(buf.data)))
		err := s.Load(r)
		if err != nil {
			t.Errorf("there is an error during the Session.Load(): %s", err)
		}

		if s.ReqCount != 10 {
			t.Errorf("expected req count: 10, got: %d", s.ReqCount)
		}
		if s.Response.Status != "200 OK" {
			t.Errorf("expected res status: 200 OK, got: %s", s.Response.Status)
		}
		v, ok := s.Response.Headers["Content-Type"]
		if !ok || slices.Compare(v, []string{"application/json; charset=utf-8"}) != 0 {
			t.Errorf(
				`expected: res header Content-Type="application/json; charset=utf-8", got: %#v`, v,
			)
		}
		if slices.Compare(s.Response.BodyLines, rsb) != 0 {
			t.Errorf("expected res body lines %s != %s", rsb, s.Response.BodyLines)
		}
	})
}
