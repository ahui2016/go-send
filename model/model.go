package model

// SMS 意思是短消息，取名灵感来自传统手机短信。
type SMS struct {
	ID        string // primary key
	Msg       string
	CreatedAt string `storm:"index"` // ISO8601
	UpdatedAt string `storm:"index"`
	DeletedAt string `storm:"index"`
}

// File .
type File struct {
	ID        string // primary key
	Name      string `storm:"unique"`
	Size      int64
	Type      string // MIME
	Checksum  string `storm:"unique"` // hex(sha256)
	CreatedAt string `storm:"index"`  // ISO8601
	UpdatedAt string `storm:"index"`
	DeletedAt string `storm:"index"`
}
