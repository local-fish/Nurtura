package app

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// serverErr logs the real error and returns a generic 500. Never logs request bodies.
func serverErr(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	writeErr(w, http.StatusInternalServerError, "internal error")
}

// decode reads a JSON body (max 1MB). If allowEmpty, an empty body is not an error.
func decode(w http.ResponseWriter, r *http.Request, v any, allowEmpty bool) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	err := dec.Decode(v)
	if allowEmpty && errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// pathUUID returns the {name} path value if it is a valid UUID.
func pathUUID(r *http.Request, name string) (string, bool) {
	v := r.PathValue(name)
	return strings.ToLower(v), uuidRe.MatchString(v)
}

// ageMonths returns completed months between dob and at.
func ageMonths(dob, at time.Time) int {
	m := (at.Year()-dob.Year())*12 + int(at.Month()-dob.Month())
	if at.Day() < dob.Day() {
		m--
	}
	if m < 0 {
		m = 0
	}
	return m
}

