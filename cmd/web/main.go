package main

import (
	"fmt"
	"log"
	"net/http"
)

type application struct {
}

func main() {

	port := 8080
	// set up an app config
	app := application{}

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
