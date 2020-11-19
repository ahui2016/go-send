package database // import "github.com/ahui2016/go-send/database"

import (
	"errors"
	"sync"
	"time"

	"github.com/ahui2016/go-send/model"
	"github.com/ahui2016/goutil"
	"github.com/ahui2016/goutil/session"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
)

const (

	// 数据库条目数上限
	//countLimit = 100

	// 文件的最长保存时间
	keepAlive = time.Hour * 24 * 30 // 30 days

	// 文件变灰时间，应小于 keepAlive, 预警该文件即将被自动删除。
	turnGrey = time.Hour * 24 * 15 // 15 days
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

	// 只在 package database 外部使用锁，不在 package database 内部使用锁。
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
	err1 := db.createIndexes()
	err2 := db.initFirstID()
	err3 := db.initTotalSize()
	err4 := db.DB.ReIndex(&Message{}) // 后续要删除
	return goutil.WrapErrors(err1, err2, err3, err4)
}

// Close 只是 db.DB.Close(), 不清空 db 里的其它部分。
func (db *DB) Close() error {
	return db.DB.Close()
}

// 创建 bucket 和索引
func (db *DB) createIndexes() error {
	return db.DB.Init(&Message{})
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
	_, err = db.GetTotalSize()
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

// GetTotalSize .
func (db *DB) GetTotalSize() (size int64, err error) {
	err = db.DB.Get(metadataBucket, totalSizeKey, &size)
	return
}

func (db *DB) setTotalSize(size int64) error {
	return db.DB.Set(metadataBucket, totalSizeKey, size)
}

func (db *DB) checkTotalSize(addition int64) error {
	totalSize, err := db.GetTotalSize()
	if err != nil {
		return err
	}
	if totalSize+addition > db.capacity {
		return errors.New("超过数据库总容量上限")
	}
	return nil
}

// addTotalSize 用于向数据库添加或删除单项内容时更新总体积。
// 添加时，应先使用 db.checkTotalSize, 再使用 db.Save, 最后使才使用 db.addTotalSize
// 删除时，应先获取即将删除项目的体积，再删除，最后使用 db.addTotalSize, 此时 addition 应为负数。
func (db *DB) addTotalSize(addition int64) error {
	totalSize, err := db.GetTotalSize()
	if err != nil {
		return err
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

// NewTextMsg .
func (db *DB) NewTextMsg(textMsg string) (*Message, error) {
	message, err := db.newMessage(model.TextMsg)
	if err != nil {
		return nil, err
	}
	if err := message.SetTextMsg(textMsg); err != nil {
		return nil, err
	}
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
	if err := message.SetFileNameType(filename); err != nil {
		return nil, err
	}
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
	if err := db.checkTotalSize(message.FileSize); err != nil {
		return err
	}
	_, err := db.getByID(message.ID)
	if err == nil {
		return errors.New("id: " + message.ID + " already exists")
	}
	if err := db.DB.Save(message); err != nil {
		return err
	}
	return db.addTotalSize(message.FileSize)
}

// Delete by id
func (db *DB) Delete(id string) error {
	message, err1 := db.getByID(id)
	err2 := db.DB.DeleteStruct(message)
	if err := goutil.WrapErrors(err1, err2); err != nil {
		return err
	}
	return db.addTotalSize(-message.FileSize)
}

func (db *DB) getByID(id string) (*Message, error) {
	var message Message
	err := db.DB.One("ID", id, &message)
	return &message, err
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

// OldItems 找出最老的 (更新日期最早的) n 条记录，返回 []Message.
func (db *DB) OldItems(n int) (items []Message, err error) {
	err = db.DB.AllByIndex("UpdatedAt", &items, storm.Limit(n))
	return
}

// GreyItems 找出变灰的条目
func (db *DB) GreyItems() (items []Message, err error) {
	// 如果 UpdatedAt 在 turnGreyTime 之前，说明它已变灰。
	turnGreyTime := time.Now().Add(-turnGrey).Format(model.ISO8601)
	err = db.DB.Select(q.Lt("UpdatedAt", turnGreyTime)).Find(&items)
	return
}

// ExpiredItems 找出过期的条目
func (db *DB) ExpiredItems() (items []Message, err error) {
	// 如果 UpdatedAt 在 expiredTime 之前，说明它已过期。
	expiredTime := time.Now().Add(-keepAlive).Format(model.ISO8601)
	err = db.DB.Select(q.Lt("UpdatedAt", expiredTime)).Find(&items)
	return
}

// OldFiles 找出最老的 (更新日期最早的) n 个文件 (Type = FileMsg)
// 返回 []Message.
func (db *DB) OldFiles(n int) (files []Message, err error) {
	query := db.queryOldFiles(n)
	err = query.Find(&files)
	return
}

func (db *DB) queryOldFiles(n int) storm.Query {
	return db.DB.Select(q.Eq("Type", model.FileMsg)).
		OrderBy("UpdatedAt").Limit(n)
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
	return db.DB.UpdateField(&Message{ID: id}, "UpdatedAt", goutil.TimeNow(model.ISO8601))
}

// LastTextMsg .
func (db *DB) LastTextMsg() (string, error) {
	var message Message
	err := db.DB.Select(q.Eq("Type", model.TextMsg)).
		OrderBy("UpdatedAt").Reverse().First(&message)
	if err != nil {
		return "", err
	}
	return message.TextMsg, nil
}

// InsertTextMsg .
func (db *DB) InsertTextMsg(textMsg string) (message *Message, err error) {
	if message, err = db.NewTextMsg(textMsg); err != nil {
		return
	}
	return message, db.Insert(message)
}
