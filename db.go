package main

import (
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"time"
)

const (
	URL_TABLE  = "urls"
	VALID_CODE = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_-"
)

type URLData struct {
	ID      string
	URL     string
	Hashsum string
	Visited time.Time
}

func (d *URLData) computeHash() string {
	sum := md5.Sum([]byte(d.URL))
	return hex.EncodeToString(sum[:])
}

func createTable(db *sql.DB) error {
	qstr := "CREATE TABLE IF NOT EXISTS `" + URL_TABLE + "` (id CHAR(6) PRIMARY KEY,raw TEXT,visited TIMESTAMP DEFAULT CURRENT_TIMESTAMP, `hashsum` CHAR(32),INDEX (visited), UNIQUE KEY hashid (`hashsum`)) DEFAULT CHARACTER SET utf8 DEFAULT COLLATE utf8_general_ci"
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
	stmtIns  *sql.Stmt
	stmtSel  *sql.Stmt
	stmtHash *sql.Stmt
	stmtUpd  *sql.Stmt
)

func initStmt(db *sql.DB) {
	var err error
	stmtIns, err = db.Prepare(`INSERT INTO ` + URL_TABLE + " (id,raw,`hashsum`) VALUES (?,?,?)")
	if err != nil {
		log.Fatalf("Error preparing statement for insertion: %s", err)
	}

	stmtSel, err = db.Prepare(`SELECT raw,visited FROM ` + URL_TABLE + ` WHERE id=?`)
	if err != nil {
		log.Fatalf("Error preparing statement for loading: %s", err)
	}

	stmtHash, err = db.Prepare(`SELECT id,raw,visited FROM ` + URL_TABLE + ` WHERE hashsum=?`)
	if err != nil {
		log.Fatalf("Error preparing statement for hashed loading: %s", err)
	}

	stmtUpd, err = db.Prepare(`UPDATE ` + URL_TABLE + ` SET visited=CURRENT_TIMESTAMP WHERE id=?`)
	if err != nil {
		log.Fatalf("Error preparing statement for updating: %s", err)
	}
}

type DupErr struct{}

func (e DupErr) Error() string {
	return "duplicated"
}

func save(d *URLData) error {
	res, err := stmtIns.Exec(d.ID, d.URL, d.Hashsum)
	if err != nil {
		return err
	}

	if cnt, _ := res.RowsAffected(); cnt != 1 {
		return DupErr{}
	}

	return nil
}

func upd(id string) error {
	_, err := stmtUpd.Exec(id)
	return err
}

func loadbyhash(h string) *URLData {
	var raw string
	var id string
	var t time.Time
	row := stmtHash.QueryRow(h)

	if err := row.Scan(&id, &raw, &t); err != nil {
		return nil
	}

	return &URLData{ID: id, URL: raw, Visited: t}
}

func load(id string) *URLData {
	var raw string
	var t time.Time
	row := stmtSel.QueryRow(id)

	if err := row.Scan(&raw, &t); err != nil {
		return nil
	}

	return &URLData{ID: id, URL: raw, Visited: t}
}
