package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	bind := os.Getenv("BIND_ADDRESS")
	if bind == "" {
		bind = ":80"
	}
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("You must set envvar MYSQL_DSN before running this program.")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Cannot connect to MySQL server with %s: %s", dsn, err)
	}

	if err = createTable(db); err != nil {
		log.Fatalf("Cannot prepare table: %s", err)
	}

	initStmt(db)

	http.HandleFunc("/s", postHandler)
	http.HandleFunc("/", rootHandler)

	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Print(err)
	}
}
