package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"webapp/pkg/data"
	"webapp/pkg/db"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	Session *scs.SessionManager
	DSN     string
	DB      db.PostgresConn
}

func main() {

	gob.Register(data.User{})

	port := 9000
	// set up an app config
	app := application{}

	// read DSN as flag from commandline when starting
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection")
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app.DB = db.PostgresConn{DB: conn}

	// get a session manager
	app.Session = getSession()

	// get application route
	mux := app.routes()

	// print out a message
	log.Printf("Starting server on port %d", port)

	// start the server
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		log.Fatal(err)
	}
}
