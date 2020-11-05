package database

import (
	"log"
	"sync"

	"github.com/ahui2016/go-send/model"
	"github.com/ahui2016/go-send/session"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
)

// 用来保存当前最新 id.
const (
	currentIDBucket = "current-id-bucket"
	currentIDKey    = "current-id-key"
)

type (
	Message    = model.Message
	IncreaseID = model.IncreaseID
)

// DB .
type DB struct {
	path string
	DB   *storm.DB
	Sess *session.Manager
	sync.Mutex
}

// Open .
func (db *DB) Open(maxAge int, dbPath string) (err error) {
	if db.DB, err = storm.Open(dbPath); err != nil {
		return err
	}
	db.path = dbPath
	db.Sess = session.NewManager(maxAge)
	if err := db.createIndexes(); err != nil {
		return err
	}
	log.Print(db.path)
	return nil
}

// 创建 bucket 和索引，并生成初始 id.
func (db *DB) createIndexes() error {
	if err := db.DB.Init(&Message{}); err != nil {
		return err
	}
	_, err := db.getCurrentID()

	// 如果 current-id 不存在，则生成 first-id.
	if err == storm.ErrNotFound {
		if err := db.createFirstID(); err != nil {
			return err
		}
	}

	// 如果有其他错误，则返回错误。
	if err != nil && err != storm.ErrNotFound {
		return err
	}
	return nil
}

func (db *DB) getCurrentID() (id IncreaseID, err error) {
	err = db.DB.Get(currentIDBucket, currentIDKey, &id)
	return
}

func (db *DB) createFirstID() error {
	id := model.FirstID()
	return db.DB.Set(currentIDBucket, currentIDKey, id)
}

// Close .
func (db *DB) Close() error {
	return db.DB.Close()
}

// NewTextMsg .
func (db *DB) NewTextMsg(textMsg string) (*Message, error) {
	message, err := db.newMessage(model.TextMsg)
	if err != nil {
		return nil, err
	}
	message.TextMsg = textMsg
	return message, nil
}

// NewZipMsg 用于自动打包，具有特殊的文件类型，避免重复打包。
func (db *DB) NewZipMsg(filename string) (*Message, error) {
	message, err := db.NewFileMsg(filename)
	if err != nil {
		return nil, err
	}
	message.FileType = model.GosendZip
	return message, nil
}

// NewFileMsg .
func (db *DB) NewFileMsg(filename string) (*Message, error) {
	message, err := db.newMessage(model.FileMsg)
	if err != nil {
		return nil, err
	}
	message.SetFileNameType(filename)
	return message, nil
}

func (db *DB) newMessage(msgType model.MsgType) (*Message, error) {
	id, err := db.getNextID()
	if err != nil {
		return nil, err
	}
	message := model.NewMessage(id.String(), msgType)
	return message, nil
}

func (db *DB) getNextID() (nextID IncreaseID, err error) {
	db.Lock()
	defer db.Unlock()

	currentID, err := db.getCurrentID()
	if err != nil {
		return nextID, err
	}
	nextID = currentID.Increase()
	if err := db.DB.Set(currentIDBucket, currentIDKey, &nextID); err != nil {
		return nextID, err
	}
	return
}

// Insert .
func (db *DB) Insert(message *Message) error {
	return db.Save(message)
}

// Save wraps storm.DB.Save with a lock.
func (db *DB) Save(data interface{}) error {
	db.Lock()
	defer db.Unlock()
	return db.DB.Save(data)
}

// AllByUpdatedAt .
func (db *DB) AllByUpdatedAt() (all []Message, err error) {
	err = db.DB.AllByIndex("UpdatedAt", &all)
	return
}

// AllFiles finds all files(Type = FileMsg).
func (db *DB) AllFiles() (files []Message, err error) {
	err = db.DB.Find("Type", model.FileMsg, &files)
	return
}

// DeleteAllFiles .
func (db *DB) DeleteAllFiles() error {
	return db.DB.Select(q.Eq("Type", model.FileMsg)).Delete(new(Message))
}
