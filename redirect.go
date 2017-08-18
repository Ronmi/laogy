package main

import (
	"log"
	"net/http"
	"regexp"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) == 7 {
		// path /xxxxxx costs 7 bytes
		redirectHandler(w, r)
		return
	}

	http.NotFound(w, r)
}

var validator *regexp.Regexp

func init() {
	var err error
	regstr := `^[` + VALID_CODE + `]{6}$`
	validator, err = regexp.Compile(regstr)
	if err != nil {
		log.Fatalf("Error compiling regexp %s: %s", regstr, err)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	msg := []byte(code + " not found")
	if !validator.MatchString(code) {
		w.WriteHeader(404)
		w.Write(msg)
		return
	}

	data := load(code)
	if data == nil {
		w.WriteHeader(404)
		w.Write(msg)
		return
	}

	http.Redirect(w, r, data.URL, 301)
}
