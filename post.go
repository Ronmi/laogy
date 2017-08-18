package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/dgryski/dgoogauth"
)

const tmpl = `{"data":{"code":"%s"}}`

type postHandler struct {
	sharedSecret string
	otpConfig    *dgoogauth.OTPConfig
}

func (h postHandler) validateTOTP(secret string) bool {
	if h.otpConfig == nil {
		return true
	}

	if res, err := h.otpConfig.Authenticate(secret); err == nil && res {
		return true
	}

	return false
}

func (h postHandler) validateShared(secret string) bool {
	if h.otpConfig != nil || h.sharedSecret == "" {
		return true
	}

	return h.sharedSecret == secret
}

func (h postHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	secret := r.Form.Get("secret")

	if !h.validateTOTP(secret) {
		http.Error(w, "Not permitted.", 403)
		return
	}

	if !h.validateShared(secret) {
		http.Error(w, "Not permitted.", 403)
		return
	}

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
		} else {
			log.Printf("error saving ID %s: %s", data.ID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(strings.Replace(tmpl, "%s", data.ID, 1)))
}
