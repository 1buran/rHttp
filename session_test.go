package main

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

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

	var buf strings.Builder
	t.Run("save", func(t *testing.T) {
		s, _ := NewSession(rq, rs, 10, "10.57µs", rqf, rsb)

		err := s.Save(&buf)
		if err != nil {
			t.Errorf("there is an error during the Session.Save(): %s", err)
		}

		if !strings.Contains(
			buf.String(), `"url":"/api/endpoint","headers":{"Auth-Token":["token"]`) {
			t.Errorf("not found expected data, saved: %s", buf.String())
		}
	})

	t.Run("load", func(t *testing.T) {
		var rq *http.Request
		var rs *http.Response

		rqf := map[string][]string{}
		rsb := []string{"{", `"msg": "hello!",`, `"id": 455`, "}"}

		s, _ := NewSession(rq, rs, 10, "10.57µs", rqf, []string{})
		r := strings.NewReader(buf.String())

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
