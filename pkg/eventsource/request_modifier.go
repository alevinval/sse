package eventsource

import "net/http"

// RequestModifier function for modifying the HTTP connection request.
type RequestModifier func(r *http.Request)

// WithBasicAuth adds basic authentication to the HTTP request
func WithBasicAuth(username, password string) RequestModifier {
	return func(r *http.Request) {
		r.SetBasicAuth(username, password)
	}
}

// WithBearerTokenAuth adds bearer token header to the HTTP request
func WithBearerTokenAuth(token string) RequestModifier {
	return func(r *http.Request) {
		r.Header.Add("Authorization", "Bearer "+token)
	}
}
