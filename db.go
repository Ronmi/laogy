package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"log"
	"time"
)

const (
	URL_TABLE  = "urls"
	VALID_CODE = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_-"
)

type URLData struct {
	ID        string
	URL       string
	CreatedAt time.Time
}

func createTable(db *sql.DB) error {
	qstr := `CREATE TABLE IF NOT EXISTS ` + URL_TABLE + ` (id CHAR(6) PRIMARY KEY,raw TEXT,created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,INDEX (created)) DEFAULT CHARACTER SET utf8 DEFAULT COLLATE utf8_general_ci`
	_, err := db.Exec(qstr)
	return err
}

func GenCode() string {
	ret := []byte("      ")
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	for idx, b := range buf {
		pos := int(b % 64)
		ret[idx] = VALID_CODE[pos]
	}
	return string(ret)
}

var (
	stmtIns *sql.Stmt
	stmtSel *sql.Stmt
)

func initStmt(db *sql.DB) {
	var err error
	stmtIns, err = db.Prepare(`INSERT INTO ` + URL_TABLE + ` (id,raw) VALUES (?,?)`)
	if err != nil {
		log.Fatalf("Error preparing statement for insertion: %s", err)
	}

	stmtSel, err = db.Prepare(`SELECT raw,created FROM ` + URL_TABLE + ` WHERE id=?`)
	if err != nil {
		log.Fatalf("Error preparing statement for loading: %s", err)
	}
}

func save(d URLData) error {
	res, err := stmtIns.Exec(d.ID, d.URL)
	if err != nil {
		return err
	}

	if cnt, _ := res.RowsAffected(); cnt != 1 {
		return errors.New("duplicated")
	}

	return nil
}

func load(id string) *URLData {
	var raw string
	var t time.Time
	row := stmtSel.QueryRow(id)

	if err := row.Scan(&raw, &t); err != nil {
		return nil
	}

	return &URLData{id, raw, t}
}
