package database

import (
	"log"

	"github.com/ahui2016/go-send/model"
	"github.com/ahui2016/go-send/session"
	"github.com/asdine/storm/v3"
)

type (
	Message = model.Message
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
	if err := db.DB.Init(&Message{}); err != nil {
		return err
	}
	db.path = dbPath
	db.Sess = session.NewManager(maxAge)
	log.Print(db.path)
	return nil
}

// Close .
func (db *DB) Close() error {
	return db.DB.Close()
}

// Insert .
func (db *DB) Insert(message *Message) error {
	return db.DB.Save(message)
}

// AllByUpdatedAt .
func (db *DB) AllByUpdatedAt() (all []Message, err error) {
	err = db.DB.AllByIndex("UpdatedAt", &all)
	return
}
