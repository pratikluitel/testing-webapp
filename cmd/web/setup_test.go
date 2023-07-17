package main

import (
	"os"
	"testing"
	"webapp/pkg/repository/dbrepo"
)

var app application

// this function is always executed before any test runs
// this is useful for setting up databases, sessions, etc that will then
// not be needed to defined in every single test
func TestMain(m *testing.M) {
	pathToTemplates = "./../../templates"
	// get a session manager
	app.Session = getSession()

	app.DB = &dbrepo.TestDBRepo{}

	os.Exit(m.Run())
}
