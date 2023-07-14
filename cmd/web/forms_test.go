package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Has(t *testing.T) {
	form := NewForm(nil)
	has := form.Has("hello") // a val that doesn't exist

	if has {
		t.Error("form shows has field when it should not")
	}

	postedData := url.Values{}
	postedData.Add("ex", "ex") // a form with data
	form = NewForm(postedData)

	has = form.Has("ex") // a val that doesn't exist

	if !has {
		t.Error("form shows does not have field when it does")
	}

}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil) // arbitary request
	form := NewForm(r.PostForm)

	form.Required("a", "b", "c")

	if form.Valid() { // a, b, c should not exist
		t.Error("form shows valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a") // a form with data
	postedData.Add("b", "b") // a form with data
	postedData.Add("c", "c") // a form with data

	r, _ = http.NewRequest("POST", "/", nil)
	r.PostForm = postedData

	form = NewForm(r.PostForm)

	form.Required("a", "b", "c")

	if !form.Valid() { // a, b, c should not exist
		t.Error("form shows not valid when required fields are present")
	}
}

func TestForm_Check(t *testing.T) {
	form := NewForm(nil)
	form.Check(false, "password", "password is invalid")

	if form.Valid() {
		t.Error("Valid() returns true, it should return false")
	}

	form = NewForm(nil)
	form.Check(true, "password", "password is invalid")

	if !form.Valid() {
		t.Error("Valid() returns false, it should return true")
	}
}

func TestForm_ErrorGet(t *testing.T) {
	form := NewForm(nil)
	form.Check(false, "password", "password is invalid")
	s := form.Errors.Get("password")

	if len(s) == 0 {
		t.Error("Should have an error returned from Get, but did not")
	}

	s = form.Errors.Get("whatever")
	if len(s) != 0 {
		t.Error("Should not have an error returned from Get, but did")
	}
}
