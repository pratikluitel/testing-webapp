package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

func Test_application_addIPToContext(t *testing.T) {
	tests := []struct {
		headerName  string
		headerValue string
		addr        string
		emptyAddr   bool
	}{
		{"", "", "", false},
		{"", "", "", true},
		{"X-Forwarded-For", "192.3.2.1", "", false},
		{"", "", "hello:world", false},
	}
	// create a dummy handler for testing the context

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// make sure that the value exists in the context
		val := r.Context().Value(contextUserKey)
		if val == nil {
			t.Error(contextUserKey, "not present")
		}

		// make sure what was returned was a string
		ip, ok := val.(string)
		if !ok {
			t.Error("not string")
		}
		t.Log(ip)
	})

	for _, e := range tests {
		// create the handler to test
		handlerToTest := app.addIPToContext(nextHandler)

		req := httptest.NewRequest("GET", "http://testing", nil)

		if e.emptyAddr {
			req.RemoteAddr = ""
		}
		if len(e.headerName) > 0 {
			req.Header.Add(e.headerName, e.headerValue)
		}
		if len(e.addr) > 0 {
			req.RemoteAddr = e.addr
		}

		handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
	}
}

func Test_application_auth(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	var tests = []struct {
		name   string
		isAuth bool
	}{
		{"logged in", true},
		{"not logged in", false},
	}

	for _, e := range tests {
		handlerToTest := app.auth(nextHandler)
		req := httptest.NewRequest("GET", "http://testing", nil)
		req = addContextAndSessionToRequest(req, app)
		if e.isAuth {
			app.Session.Put(req.Context(), "user", data.User{ID: 1}) // dummy user
		}
		rr := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rr, req)

		if e.isAuth && rr.Code != http.StatusOK {
			t.Errorf("%s: expected status code %d, but got %d", e.name, http.StatusOK, rr.Code)

		}

		if !e.isAuth && rr.Code != http.StatusTemporaryRedirect {
			t.Errorf("%s: expected status code %d, but got %d", e.name, http.StatusTemporaryRedirect, rr.Code)
		}
	}
}
