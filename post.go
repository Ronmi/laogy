package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/dgryski/dgoogauth"
)

var reCAPTCHAValidator *regexp.Regexp

func init() {
	var err error
	reCAPTCHAValidator, err = regexp.Compile(`"success":\s*true`)
	if err != nil {
		log.Fatalf("Cannot compile regexp for validating reCAPTCHA: %s", err)
	}
}

const tmpl = `{"data":{"code":"%s"}}`

type postHandler struct {
	sharedSecret    string
	otpConfig       *dgoogauth.OTPConfig
	reCAPTCHASecret string
}

func (h postHandler) validateRECAPTCHA(res string) bool {
	if h.reCAPTCHASecret == "" {
		return true
	}

	resp, err := http.PostForm(
		`https://www.google.com/recaptcha/api/siteverify`,
		url.Values{
			"secret":   {h.reCAPTCHASecret},
			"response": {res},
		},
	)
	defer resp.Body.Close()

	if err != nil {
		log.Printf("failed to verify reCAPTCHA: %s", err)
		return false
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to get reCAPTCHA result from Google: %s", err)
		return false
	}

	return reCAPTCHAValidator.Match(buf)
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

	if !h.validateRECAPTCHA(r.Form.Get("g-recaptcha-response")) {
		http.Error(w, "Not human.", 403)
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
