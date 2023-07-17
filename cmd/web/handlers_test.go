package main

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	var tests = []struct {
		name                    string
		url                     string
		expectedStatusCode      int
		expectedURL             string // the final url we are going to be at as a result
		expectedFirstStatusCode int
	}{
		{"home", "/", http.StatusOK, "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound, "/fish", http.StatusNotFound},
		{"profile", "/user/profile", http.StatusOK, "/", http.StatusTemporaryRedirect},
	}

	routes := app.routes()

	// create a test server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// a client that returns the response code for the 1st redirect, not the last one like as is done by default in above
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

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

		if resp.Request.URL.Path != e.expectedURL {
			t.Errorf("for %s: expected final url %s, but got %s", e.name, e.expectedURL, resp.Request.URL.Path)
		}

		respcl, _ := client.Get(ts.URL + e.url)
		if respcl.StatusCode != e.expectedFirstStatusCode {
			t.Errorf("%s: expected first returned status code to be %d but got %d", e.name, e.expectedFirstStatusCode, respcl.StatusCode)
		}
	}
}

func Test_application_ipFromContext(t *testing.T) {

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

func Test_application_Home(t *testing.T) {
	var tests = []struct {
		name         string
		putInSession string // what we need to put in session to run the test
		expectedHTML string
	}{
		{"first visit", "", "From Session:"}, // no session data in first visit
		{"second visit", "test", "From Session: test"},
	}

	for _, e := range tests {
		req, _ := http.NewRequest("GET", "/", nil)
		req = addContextAndSessionToRequest(req, app)
		_ = app.Session.Destroy(req.Context()) // clear session if it already exists

		if e.putInSession != "" {
			app.Session.Put(req.Context(), "test", e.putInSession)
		}
		// a dummy response writer
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.Home)

		handler.ServeHTTP(rr, req)

		// check status code
		if rr.Code != http.StatusOK {
			t.Errorf("Test_application_Home expected http.StatusOK, but got %d", rr.Code)
		}

		// check the session stored info html
		body, _ := io.ReadAll(rr.Body)
		if !strings.Contains(string(body), e.expectedHTML) {
			t.Errorf("%s: Did not find %s in html", e.name, e.expectedHTML)
		}

	}
}

func Test_application_renderWithBadTemplate(t *testing.T) {
	// set template path to a location with a bad template
	pathToTemplates = "./testdata/"

	req, _ := http.NewRequest("GET", "/", nil)
	req = addContextAndSessionToRequest(req, app)
	rr := httptest.NewRecorder()

	err := app.render(rr, req, "bad.page.gohtml", &TemplateData{})

	if err == nil {
		t.Error("Expected an error from bad template, but did not get error")
	}

	pathToTemplates = "./../../templates/"
}

func getCtx(req *http.Request) context.Context {
	ctx := context.WithValue(req.Context(), contextUserKey, "unknown")

	return ctx
}

func addContextAndSessionToRequest(req *http.Request, app application) *http.Request {
	req = req.WithContext(getCtx(req))

	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session"))

	return req.WithContext(ctx)
}

func Test_application_Login(t *testing.T) {
	var tests = []struct {
		name               string
		postedData         url.Values
		expectedStatusCode int
		expectedLoc        string
	}{
		{
			name: "valid login credentials",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/user/profile",
		},
		{
			name: "missing form data",
			postedData: url.Values{
				"email":    {""},
				"password": {""},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
		{
			name: "user not found",
			postedData: url.Values{
				"email":    {"ss@example.com"},
				"password": {"seet"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
		{
			name: "bad credentials",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"wrong"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
	}
	for _, e := range tests {
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(e.postedData.Encode()))
		req = addContextAndSessionToRequest(req, app)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.Login)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: returned wrong status code, expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		actualLoc, err := rr.Result().Location()
		if err == nil {
			if actualLoc.String() != e.expectedLoc {
				t.Errorf("%s: returned wrong location, expected %s but got %s", e.name, e.expectedLoc, actualLoc.String())
			}
		} else {
			t.Errorf("%s: No location header ser", e.name)
		}
	}
}
