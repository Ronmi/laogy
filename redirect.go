package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"regexp"
)

var validator *regexp.Regexp

func init() {
	var err error
	regstr := `^[` + VALID_CODE + `]{6}$`
	validator, err = regexp.Compile(regstr)
	if err != nil {
		log.Fatalf("Error compiling regexp %s: %s", regstr, err)
	}
}

type Redir struct {
	GAID string
}

func (h *Redir) rootHandler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) == 7 {
		// path /xxxxxx costs 7 bytes
		h.redirectHandler(w, r)
		return
	}

	http.NotFound(w, r)
}

const htmltmpl = `<!DOCTYPE html>
<html lang="zh-Hant">
  <head>
    <script>
      window.ga = window.ga || function(){
        (ga.q = ga.q || []).push(arguments)
      };
      ga.l=+new Date;
      ga('create', '%s', 'auto');
      ga('send', 'pageview');
      ga(function() {
        location.href='%s';
      });
    </script>
    <script async src="https://www.google-analytics.com/analytics.js"></script>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width,initial-scale=1.0">
    <meta name="description" content="老雞歪縮網址">
    <link rel="icon" href="/favicon.ico">
    <title>LAO.GY</title>
  </head>
  <body>
    Redirecting to <a href="%s">%s</a>
    <script defer>
      setTimeout(function() {location.href='%s'}, 5000);
    </script>
  </body>
</html>`

func (h *Redir) redirectHandler(w http.ResponseWriter, r *http.Request) {
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

	if h.GAID == "" {
		http.Redirect(w, r, data.URL, 301)
		return
	}

	s := html.EscapeString(data.URL)
	fmt.Fprintf(
		w,
		htmltmpl,
		h.GAID,
		data.URL,
		s,
		s,
		data.URL,
	)

	id := data.ID
	go upd(id)
}
