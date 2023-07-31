package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"webapp/pkg/data"

	"github.com/go-chi/chi/v5"
)

func Test_application_authenticate(t *testing.T) {
	var theTests = []struct {
		name               string
		requestBody        string // json as string
		expectedStatusCode int
	}{
		{"valid user", `{"email":"admin@example.com","password":"secret"}`, http.StatusOK},
		{"not json", `not json`, http.StatusUnauthorized},
		{"empty json", `{}`, http.StatusUnauthorized},
		{"empty email", `{"email":""}`, http.StatusUnauthorized},
		{"empty password", `{"email":"admin@example.com","password":""}`, http.StatusUnauthorized},
		{"invalid user", `{"email":"admin@someotherdomain.com","password":"secret"}`, http.StatusUnauthorized},
	}

	for _, e := range theTests {
		reader := strings.NewReader(e.requestBody)
		req, _ := http.NewRequest("POST", "/auth", reader)
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.authenticate)

		handler.ServeHTTP(rr, req)

		if e.expectedStatusCode != rr.Code {
			t.Errorf("%s: returned wrong status code; expected %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}

func Test_application_refresh(t *testing.T) {
	var tests = []struct {
		name               string
		token              string
		expectedStatusCode int
		resetRefreshTime   bool
	}{
		{"valid", "", http.StatusOK, true},
		{"valid but not ready to expire", "", http.StatusTooEarly, false},
		{"expired token", expiredToken, http.StatusBadRequest, false},
	}

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	oldRefreshTime := refreshTokenExpiry

	for _, e := range tests {
		var tkn string
		if e.token == "" {
			if e.resetRefreshTime {
				refreshTokenExpiry = time.Second * 1
			}
			tokens, _ := app.generateTokenPair(&testUser)
			tkn = tokens.RefreshToken
		} else {
			tkn = e.token
		}

		postedData := url.Values{
			"refresh_token": {tkn},
		}

		req, _ := http.NewRequest("POST", "/refresh-token", strings.NewReader(postedData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.refresh)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: expected status of %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		refreshTokenExpiry = oldRefreshTime
	}

}

// one test to test multiple handlers
func Test_application_userHandlers(t *testing.T) {
	var tests = []struct {
		name               string
		method             string
		json               string
		paramID            string
		handler            http.HandlerFunc
		expectedStatusCode int
	}{
		{"allUsers", "GET", "", "", app.allUsers, http.StatusOK},
		{"getUser valid", "GET", "", "1", app.getUser, http.StatusOK},
		{"getUser invalid", "GET", "", "0", app.getUser, http.StatusBadRequest},
		{"getUser bad url param", "GET", "", "y", app.getUser, http.StatusBadRequest},

		{"deleteUser", "DELETE", "", "1", app.deleteUser, http.StatusNoContent},
		{"deleteUser bad url param", "DELETE", "", "y", app.deleteUser, http.StatusBadRequest},

		{
			"updateUser valid",
			"PATCH",
			`{"id":1,"first_name":"administrator","last_name":"user", "email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusNoContent,
		},
		{
			"updateUser invalid",
			"PATCH",
			`{"id":2,"first_name":"administrator","last_name":"user", "email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"updateUser invalid json",
			"PATCH",
			`{"id":1,first_name:"administrator","last_name":"user", "email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"insertUser valid",
			"PUT",
			`{"first_name":"me","last_name":"who", "email":"me@example.com"}`,
			"",
			app.insertUser,
			http.StatusNoContent,
		},
		{
			"insertUser invalid",
			"PUT",
			`{"first_name":"me","foo":"bar","last_name":"who", "email":"me@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},
		{
			"insertUser bad json",
			"PUT",
			`{first_name:"me","last_name":"who", "email":"me@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},
	}

	for _, e := range tests {
		var req *http.Request
		if e.json == "" {
			req, _ = http.NewRequest(e.method, "/", nil)
		} else {
			req, _ = http.NewRequest(e.method, "/", strings.NewReader(e.json))
		}

		if e.paramID != "" {
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", e.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		}

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(e.handler)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: wrong status returned, expected %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}
