package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// it is recommended not to store primitive types in context, so creating a custom type.
type contextKey string

const contextUserKey contextKey = "user_ip"

func (app *application) ipFromContext(ctx context.Context) string {
	return ctx.Value(contextUserKey).(string)
}

func (app *application) addIPToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.Background()
		// get the ip as accurately as possible
		ip, err := getIP(r)
		if err != nil {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
			if len(ip) == 0 {
				ip = "1.1.1.1"
			}
		}
		ctx = context.WithValue(r.Context(), contextUserKey, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return "unknown", err
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return "", fmt.Errorf("UserIP: %q is not in IP:port format", r.RemoteAddr)
	}

	forward := r.Header.Get("X-Forwarded-For")
	// if this header exists, the request came through a proxy
	if len(forward) > 0 {
		ip = forward
	}

	if len(ip) == 0 {
		ip = forward
	}

	return ip, nil
}
