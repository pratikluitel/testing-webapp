package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type application struct {
	JWTSecret string
	Action    string
}

// used to generate a token so we can test our api
// Usage:
// go run ./cmd/cli -action=valid for valid token
// go run ./cmd/cli -action=expired for expired token

func main() {
	var app application
	flag.StringVar(&app.JWTSecret, "jwt-secret", "sss", "secret")
	flag.StringVar(&app.Action, "action", "valid", "action: valid|expired")
	flag.Parse()

	// generate a token
	token := jwt.New(jwt.SigningMethodHS256)

	// set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "John Doe"
	claims["sub"] = "1"
	claims["admin"] = true
	claims["aud"] = "example.com"
	claims["iss"] = "example.com"

	if app.Action == "valid" {
		expires := time.Now().UTC().Add(time.Hour * 72)
		claims["exp"] = expires.Unix()
	} else {
		expires := time.Now().UTC().Add(time.Hour * 100 * -1) // already expired token
		claims["exp"] = expires.Unix()
	}

	// create token
	if app.Action == "valid" {
		fmt.Println("Valid token:")
	} else {
		fmt.Println("Expired token:")
	}

	signedAccessToken, err := token.SignedString([]byte(app.JWTSecret))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(signedAccessToken))
}
