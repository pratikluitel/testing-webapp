package main

import (
	"log"
	"os"
	"testing"
	"webapp/pkg/db"
)

var app application

// this function is always executed before any test runs
// this is useful for setting up databases, sessions, etc that will then
// not be needed to defined in every single test
func TestMain(m *testing.M) {
	pathToTemplates = "./../../templates"
	// get a session manager
	app.Session = getSession()
	app.DSN = "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5"

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app.DB = db.PostgresConn{DB: conn}

	os.Exit(m.Run())
}
