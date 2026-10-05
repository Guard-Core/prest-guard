package main

import (
	"encoding/json"
	"net/http"

	"github.com/urfave/negroni/v3"
)

// writeJSONError writes a JSON error body. The message is marshaled rather
// than interpolated, so quotes or control characters in it can never produce
// a body that is not valid JSON. Callers pass generic messages only: the
// specific cause goes to the server log, not to anonymous clients.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	encoded, err := json.Marshal(message)
	if err != nil {
		// Marshaling a string cannot fail; this is a fixed fallback anyway.
		encoded = []byte(`"request failed"`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":` + string(encoded) + `}`))
}

// misconfiguredHandler answers 500 with a generic body when the guard was
// asked to run but its configuration is unusable. A security layer the
// operator enabled must never silently degrade to absent; the specific
// config error is logged once at load time.
func misconfiguredHandler() negroni.Handler {
	return negroni.HandlerFunc(func(w http.ResponseWriter, _ *http.Request, _ http.HandlerFunc) {
		writeJSONError(w, http.StatusInternalServerError, "guard misconfigured")
	})
}

// passThroughHandler is the no-op middleware installed when the guard is
// disabled: it forwards the request untouched.
func passThroughHandler() negroni.Handler {
	return noopMiddleware()
}

// noopMiddleware forwards requests without touching them. It exists because
// negroni's HandlerFunc receives (rw, rq, next) directly.
func noopMiddleware() negroni.Handler {
	return negroni.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request, next http.HandlerFunc) {
		next(rw, rq)
	})
}
