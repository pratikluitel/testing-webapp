package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

func Test_application_enableCORS(t *testing.T) {
	//dummy
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	var tests = []struct {
		name         string
		method       string
		expectHeader bool
	}{
		{"preflight- OPTIONS request", "OPTIONS", true},
		{"get", "GET", false},
	}

	for _, e := range tests {
		handlerToTest := app.enableCORS(nextHandler)
		req := httptest.NewRequest(e.method, "http://testing", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)
		if e.expectHeader && rr.Header().Get("Access-Control-Allow-Credentials") == "" {
			t.Errorf("%s: expected header, but did not find it", e.name)
		}

		if !e.expectHeader && rr.Header().Get("Access-Control-Allow-Credentials") != "" {
			t.Errorf("%s: expected nil header, but got one", e.name)
		}
	}
}

func Test_application_authRequired(t *testing.T) {
	//dummy
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	var tests = []struct {
		name       string
		token      string
		expectAuth bool
		setHeader  bool
	}{
		{"valid token", fmt.Sprintf("Bearer %s", tokens.Token), true, true},
		{"no token", "", false, false},
		{"invalid token", fmt.Sprintf("Bearer %s", expiredToken), false, true},
	}

	for _, e := range tests {
		req, _ := http.NewRequest("GET", "/", nil)
		if e.setHeader {
			req.Header.Set("Authorization", e.token)
		}
		rr := httptest.NewRecorder()

		handlerToTest := app.authRequired(nextHandler)

		handlerToTest.ServeHTTP(rr, req)

		if e.expectAuth && rr.Code == http.StatusUnauthorized {
			t.Errorf("%s: got code 401, and should not have", e.name)
		}

		if !e.expectAuth && rr.Code != http.StatusUnauthorized {
			t.Errorf("%s: did not get code 401, and should have", e.name)
		}
	}
}
