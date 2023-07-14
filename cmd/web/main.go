package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	Session *scs.SessionManager
}

func main() {

	port := 8080
	// set up an app config
	app := application{}

	// get a session manager
	app.Session = getSession()

	// get application route
	mux := app.routes()

	// print out a message
	log.Printf("Starting server on port %d", port)

	// start the server
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		log.Fatal(err)
	}
}
