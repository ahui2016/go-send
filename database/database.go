package database

import (
	"log"

	"github.com/ahui2016/go-send/session"
	"github.com/asdine/storm/v3"
)

// DB .
type DB struct {
	path string
	DB   *storm.DB
	Sess *session.Manager
}

// Open .
func (db *DB) Open(maxAge int, dbPath string) (err error) {
	if db.DB, err = storm.Open(dbPath); err != nil {
		return err
	}
	db.path = dbPath
	db.Sess = session.NewManager(maxAge)
	log.Print(db.path)
	return nil
}
