package main

import (
	"database/sql"
	"encoding/base32"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/dgryski/dgoogauth"
	_ "github.com/go-sql-driver/mysql"
)

func getenv(key, defaultValue string) string {
	ret := os.Getenv(key)
	if ret == "" {
		ret = defaultValue
	}

	return ret
}

func genqr(data string) string {
	return "https://chart.googleapis.com/chart?cht=qr&chs=256x256&chl=" + url.QueryEscape(data)
}

func getConfig(hexstr string) *dgoogauth.OTPConfig {
	re, err := regexp.Compile(`^[0-9a-fA-F]{20}$`)
	if err != nil {
		log.Fatalf("Failed to compile regexp: %s", err)
	}
	if !re.MatchString(hexstr) {
		return nil
	}

	// ignore err since we validated the string with regexp
	bin, _ := hex.DecodeString(hexstr)
	secret := base32.StdEncoding.EncodeToString(bin)
	ret := &dgoogauth.OTPConfig{
		Secret:     secret,
		WindowSize: 1,
	}

	issuer := getenv("TOTP_ISSUER", "my URL shorter")
	user := getenv("TOTP_USER", "admin")

	secretURL := ret.ProvisionURIWithIssuer(user, issuer)

	log.Printf("Add this secret code to your TOTP application (like Google Authenticator): %s", secretURL)
	log.Printf("Or browse this URL for a qrcode: %s", genqr(secretURL))
	return ret
}

func main() {
	bind := getenv("BIND_ADDRESS", ":80")
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("You must set envvar MYSQL_DSN before running this program.")
	}
	sharedSecret := os.Getenv("SHARED_SECRET")
	totpSecret := os.Getenv("TOTP_SECRET")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Cannot connect to MySQL server with %s: %s", dsn, err)
	}

	if err = createTable(db); err != nil {
		log.Fatalf("Cannot prepare table: %s", err)
	}

	initStmt(db)

	http.Handle("/s", postHandler{
		sharedSecret: sharedSecret,
		otpConfig:    getConfig(totpSecret),
	})
	http.HandleFunc("/", rootHandler)

	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Print(err)
	}
}
