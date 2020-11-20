package model // import "github.com/ahui2016/go-send/model"

import (
	"errors"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ahui2016/goutil"
)

// ISO8601 需要根据服务器的具体时区来设定正确的时区
const ISO8601 = "2006-01-02T15:04:05.999+00:00"

const (
	// FileNameMinLength 规定包括后缀名在内文件名长度不可小于 5.
	FileNameMinLength = 3
)

// MsgType 是一个枚举类型，用来区分 Message 的类型。
type MsgType string

const (
	// TextMsg 是文本信息，不带文件。
	TextMsg MsgType = "TextMsg"

	// FileMsg 是文件信息，带有文件。
	FileMsg MsgType = "FileMsg"
)

const (
	// GosendZip 是自定义的文件类型，表示自动打包的文件，避免重复打包。
	GosendZip = "gosend/zip"
)

// Message 表示一个数据表。
// 本来想过用 Note 来命名，但考虑到不管是熟人间互传还是个人设备间互传，
// 也不管互传文件还是互传文本信息，都更适合用 “消息、信息” 而不是 “笔记”。
type Message struct {
	ID        string // primary key
	Type      MsgType
	TextMsg   string
	FileName  string `storm:"index"`
	FileSize  int64
	FileType  string // MIME
	Checksum  string `storm:"unique"` // hex(sha256)
	CreatedAt string `storm:"index"`  // ISO8601
	UpdatedAt string `storm:"index"`
	DeletedAt string `storm:"index"`
}

// NewMessage .
func NewMessage(id string, msgType MsgType) *Message {
	now := goutil.TimeNow(ISO8601)
	return &Message{
		ID:        id,
		Type:      msgType,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetTextMsg 同时更新 TextMsg 和 FileSize,
// 注意不可直接设置 TextMsg, 每次都应该使用 SetTextMsg 以确保同时设置 FileSize.
func (message *Message) SetTextMsg(textMsg string) error {
	textMsg = strings.TrimSpace(textMsg)
	if textMsg == "" {
		return errors.New("the message is empty")
	}
	message.TextMsg = textMsg
	message.FileSize = int64(len(textMsg))
	return nil
}

// SetFileNameType 同时设置 FileName 和 FileType.
// 注意不可直接设置 FileName, 每次都应该使用 SetFileNameType 以确保同时设置 FileType.
func (message *Message) SetFileNameType(filename string) error {
	filename = strings.TrimSpace(filename)
	if len(filename) < FileNameMinLength {
		return errors.New("filename is too short")
	}
	message.FileName = filename
	message.FileType = typeByFilename(filename)
	return nil
}

// IsImage .
func (message *Message) IsImage() bool {
	return strings.HasPrefix(message.FileType, "image")
}

// ClipText 表示剪贴板文本消息，创建新的类型只是为了方便在数据库里创建一个独立的 bucket,
// 结构与 Message 一样。
type ClipText struct {
	ID        string // primary key
	Type      MsgType
	TextMsg   string
	FileName  string `storm:"index"`
	FileSize  int64
	FileType  string // MIME
	Checksum  string `storm:"unique"` // hex(sha256)
	CreatedAt string `storm:"index"`  // ISO8601
	UpdatedAt string `storm:"index"`
	DeletedAt string `storm:"index"`
}

// NewClipText .
func NewClipText(id string, msgType MsgType) *ClipText {
	now := goutil.TimeNow(ISO8601)
	return &ClipText{
		ID:        id,
		Type:      msgType,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetTextMsg .
func (clip *ClipText) SetTextMsg(textMsg string) error {
	textMsg = strings.TrimSpace(textMsg)
	if textMsg == "" {
		return errors.New("the clip is empty")
	}
	clip.TextMsg = textMsg
	clip.FileSize = int64(len(textMsg))
	return nil
}

func typeByFilename(filename string) (filetype string) {
	ext := strings.ToLower(filepath.Ext(filename))
	filetype = mime.TypeByExtension(ext)
	ext = strings.TrimPrefix(ext, ".")
	switch ext {
	case "zip", "rar", "7z", "gz", "tar", "bz", "bz2", "xz":
		filetype = "compressed/" + ext
	case "md", "xml", "html", "xhtml", "htm":
		filetype = "text/" + ext
	case "doc", "docx", "ppt", "pptx", "rtf", "xls", "xlsx":
		filetype = "office/" + ext
	case "epub", "pdf", "mobi", "azw", "azw3", "djvu":
		filetype = "ebook/" + ext
	}
	return filetype
}

// IncreaseID 用来记录自动生成 ID 的状态，便于生成特有的自增 ID.
// 该 ID 由年份与自增数两部分组成，分别取两个部分的 36 进制, 转字符串后拼接而成。
type IncreaseID struct {
	Year  int
	Count int
}

// FirstID 生成初始 id, 当且仅当程序每一次使用时（数据库为空时）使用该函数，
// 之后应使用 Increase 函数来获得新 id.
func FirstID() IncreaseID {
	nowYear := time.Now().Year()
	return IncreaseID{nowYear, 0}
}

// ParseID 把字符串形式的 id 转换为 IncreaseID.
func ParseID(strID string) (id IncreaseID, err error) {
	strYear := strID[:3] // 可以认为年份总是占前三个字符
	strCount := strID[3:]
	year, err := strconv.ParseInt(strYear, 36, 0)
	if err != nil {
		return id, err
	}
	count, err := strconv.ParseInt(strCount, 36, 0)
	if err != nil {
		return id, err
	}
	id.Year = int(year)
	id.Count = int(count)
	return
}

// Increase 使 id 自增一次，输出自增后的新 id.
// 如果当前年份大于 id 中的年份，则年份进位，Count 重新计数。
// 否则，年份不变，Count 加一。
func (id IncreaseID) Increase() IncreaseID {
	nowYear := time.Now().Year()
	if nowYear > id.Year {
		return IncreaseID{nowYear, 1}
	}
	return IncreaseID{id.Year, id.Count + 1}
}

// String 返回 id 的字符串形式。
func (id IncreaseID) String() string {
	year := strconv.FormatInt(int64(id.Year), 36)
	count := strconv.FormatInt(int64(id.Count), 36)
	return year + count
}

// CompareTo 让 id 与 another 对比，如果两者相等就返回 0,
// 如果 id 更大则返回正数，如果 id 更小则返回负数。
/*
func (id IncreaseID) CompareTo(another IncreaseID) int {
	if id.Year > another.Year {
		return 1
	}
	if id.Year < another.Year {
		return -1
	}
	if id.Year == another.Year {
		if id.Count > another.Count {
			return 1
		}
		if id.Count < another.Count {
			return -1
		}
	}
	return 0
}
*/
