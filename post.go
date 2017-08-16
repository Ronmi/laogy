package main

import (
	"net/http"
	"strings"
)

const tmpl = `{"data":{"code":"%s"}}`

func postHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	raw := r.Form.Get("url")
	if raw == "" {
		http.Redirect(w, r, "/", 302)
		return
	}

	data := URLData{
		URL: raw,
	}
	for {
		data.ID = GenCode()
		if err := save(data); err == nil {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(strings.Replace(tmpl, "%s", data.ID, 1)))
}
