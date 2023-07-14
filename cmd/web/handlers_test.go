package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	var tests = []struct {
		name               string
		url                string
		expectedStatusCode int
	}{
		{"home", "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound},
	}

	var app application
	routes := app.routes()

	// create a test server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	pathToTemplates = "./../../templates"

	//range through test data
	for _, e := range tests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s: expected status %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func Test_application_ipFromContext(t *testing.T) {

	var app application
	var tests = []struct {
		name       string
		ip         string
		expectedIp string
	}{
		{"none", "", ""},
		{"IP", "1.1.1.1", "1.1.1.1"},
	}

	for _, e := range tests {
		var ctx = context.Background()
		ctx = context.WithValue(ctx, contextUserKey, e.ip)
		ip := app.ipFromContext(ctx)
		if ip != e.expectedIp {
			t.Errorf("expected context to have %s, but found %s", e.expectedIp, ip)
		}
	}
}
