package model

import "github.com/ahui2016/goutil"

// MsgType 是一个枚举类型，用来区分 Message 的类型。
type MsgType string

const (
	TextMsg MsgType = "TextMsg"
	FileMsg MsgType = "FileMsg"
)

// Message 表示一个数据表。
// 本来想过用 Note 来命名，但考虑到不管是熟人间互传还是个人设备间互传，
// 也不管互传文件还是互传文本信息，都更适合用 “消息、信息” 而不是 “笔记”。
type Message struct {
	ID        string // primary key
	Type      MsgType
	TextMsg   string
	FileName  string `storm:"unique"`
	FileSize  int64
	FileType  string // MIME
	Checksum  string `storm:"unique"` // hex(sha256)
	CreatedAt string `storm:"index"`  // ISO8601
	UpdatedAt string `storm:"index"`
	DeletedAt string `storm:"index"`
}

func NewMessage(msgType MsgType) *Message {
	now := goutil.TimeNow()
	return &Message{
		ID:        goutil.NewID(),
		Type:      msgType,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewTextMsg(msg string) *Message {
	message := NewMessage(TextMsg)
	message.TextMsg = msg
	return message
}
