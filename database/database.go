package database

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/ahui2016/go-send/model"
	"github.com/ahui2016/go-send/session"
	"github.com/ahui2016/goutil"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
)

// 用来保存数据库的当前状态.
const (
	metadataBucket = "metadata-bucket"
	currentIDKey   = "current-id-key"
	totalSizeKey   = "total-size-key"
)

type (
	Message    = model.Message
	IncreaseID = model.IncreaseID
)

// DB .
type DB struct {
	path     string
	capacity int64
	DB       *storm.DB
	Sess     *session.Manager
	sync.Mutex
}

// Open .
func (db *DB) Open(maxAge int, cap int64, dbPath string) (err error) {
	if db.DB, err = storm.Open(dbPath); err != nil {
		return err
	}
	db.path = dbPath
	db.capacity = cap
	db.Sess = session.NewManager(maxAge)
	if err := db.createIndexes(); err != nil {
		return err
	}
	log.Print(db.path)
	return nil
}

// 创建 bucket 和索引，并初始化数据库状态.
func (db *DB) createIndexes() error {
	err1 := db.DB.Init(&Message{})
	err2 := db.initFirstID()
	err3 := db.initTotalSize()
	return WrapErrors(err1, err2, err3)
}

// WrapErrors 把多个错误合并为一个错误.
func WrapErrors(allErrors ...error) (wrapped error) {
	for _, err := range allErrors {
		if err != nil {
			if wrapped == nil {
				wrapped = err
			} else {
				wrapped = fmt.Errorf("%v | %v", err, wrapped)
			}
		}
	}
	return
}

func (db *DB) initFirstID() (err error) {
	_, err = db.getCurrentID()
	if err != nil && err != storm.ErrNotFound {
		return
	}
	if err == storm.ErrNotFound {
		id := model.FirstID()
		return db.DB.Set(metadataBucket, currentIDKey, id)
	}
	return
}

func (db *DB) initTotalSize() (err error) {
	_, err = db.getTotalSize()
	if err != nil && err != storm.ErrNotFound {
		return
	}
	if err == storm.ErrNotFound {
		return db.setTotalSize(0)
	}
	return
}

func (db *DB) getCurrentID() (id IncreaseID, err error) {
	err = db.DB.Get(metadataBucket, currentIDKey, &id)
	return
}

func (db *DB) getTotalSize() (size int64, err error) {
	err = db.DB.Get(metadataBucket, totalSizeKey, &size)
	return
}

func (db *DB) setTotalSize(size int64) error {
	return db.DB.Set(metadataBucket, totalSizeKey, size)
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
// 注意在该函数里对文件名进行了特殊处理。
func (db *DB) NewZipMsg(filename string) (*Message, error) {
	message, err := db.NewFileMsg(filename)
	if err != nil {
		return nil, err
	}
	message.FileName = filename + "_" + message.ID + ".zip"
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
	if err := db.DB.Set(metadataBucket, currentIDKey, &nextID); err != nil {
		return nextID, err
	}
	return
}

// Insert .
func (db *DB) Insert(message *Message) error {
	return db.Save(message)
}

// Save wraps storm.DB.Save with a lock.
func (db *DB) Save(message *Message) error {
	db.Lock()
	defer db.Unlock()
	if err := db.increaseTotalSize(message.FileSize); err != nil {
		return err
	}
	return db.DB.Save(message)
}

// increaseTotalSize 用于向数据库添加内容时更新总体积。
func (db *DB) increaseTotalSize(addition int64) error {
	totalSize, err := db.getTotalSize()
	if err != nil {
		return err
	}
	if totalSize+addition > db.capacity {
		return errors.New("超过数据库总容量上限")
	}
	return db.setTotalSize(totalSize + addition)
}

// recountTotalSize 用于一次性删除多个项目时重新计算数据库总体积。
func (db *DB) recountTotalSize() error {
	var totalSize int64 = 0
	err := db.DB.Select(q.True()).Each(
		new(Message), func(record interface{}) error {
			message := record.(*Message)
			totalSize += message.FileSize
			return nil
		})
	if err != nil {
		return err
	}
	return db.setTotalSize(totalSize)
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
	err := db.DB.Select(q.Eq("Type", model.FileMsg)).Delete(new(Message))
	if err != nil {
		return err
	}
	return db.recountTotalSize()
}

// OldFiles 找出最老的 (更新日期最早的) n 个文件 (Type = FileMsg)
// 返回 []Message.
func (db *DB) OldFiles(n int) (files []Message, err error) {
	query := db.queryOldFiles(n)
	err = query.Find(&files)
	return
}

// queryOldFiles 找出最老的 (更新日期最早的) n 个文件 (Type = FileMsg),
// 返回 storm.Query.
func (db *DB) queryOldFiles(n int) storm.Query {
	return db.DB.Select(q.Eq("Type", model.FileMsg)).
		OrderBy("UpdatedAt").Limit(n)
}

// DeleteOldFiles .
func (db *DB) DeleteOldFiles(n int) error {
	query := db.queryOldFiles(n)
	if err := query.Delete(new(Message)); err != nil {
		return err
	}
	return db.recountTotalSize()
}

// DeleteMessages deletes messages by IDs.
func (db *DB) DeleteMessages(messages []Message) error {
	var IDs []string
	for i := range messages {
		IDs = append(IDs, messages[i].ID)
	}
	err := db.DB.Select(q.In("ID", IDs)).Delete(new(Message))
	if err != nil {
		return err
	}
	return db.recountTotalSize()
}

// UpdateDatetime ...
func (db *DB) UpdateDatetime(id string) error {
	db.Lock()
	defer db.Unlock()
	return db.DB.UpdateField(&Message{ID: id}, "UpdatedAt", goutil.TimeNow())
}
