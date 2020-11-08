package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ahui2016/go-send/database"
	"github.com/ahui2016/goutil"
)

const (
	dataFolderName   = "gosend_data_folder"
	filesFolderName  = "files"
	databaseFileName = "gosend.db"
	gosendFileExt    = ".send"
	thumbFileExt     = ".small"
	staticFolder     = "static"
	defaultPassword  = "abc"
	passwordMaxTry   = 5

	// 99 days, for session
	maxAge = 60 * 60 * 24 * 99

	// 100 MB, for http.MaxBytesReader
	// 注意在 Nginx 的设置里进行相应的设置，例如 client_max_body_size 100m
	maxBodySize int64 = 1024 * 1024 * 100

	// 3 MB, 普通请求（没有文件的请求）需要限制更严格。
	defaultBodySize int64 = 3 << 20
)

var (
	dataDir  string
	filesDir string
	dbPath   string
)

var (
	passwordTry = 0
	HTML        = make(map[string]string)
	db          = new(database.DB)
)

func init() {
	dataDir = filepath.Join(goutil.UserHomeDir(), dataFolderName)
	filesDir = filepath.Join(dataDir, filesFolderName)
	dbPath = filepath.Join(dataDir, databaseFileName)
	fillHTML()
	goutil.MustMkdir(dataDir)
	goutil.MustMkdir(filesDir)

	// open the db here, close the db in main().
	if err := db.Open(maxAge, dbPath); err != nil {
		panic(err)
	}
}

func localFilePath(id string) string {
	return filepath.Join(filesDir, id+gosendFileExt)
}

func thumbFilePath(id string) string {
	return filepath.Join(filesDir, id+thumbFileExt)
}

func getFileAndThumb(id string) (originFile, thumb string) {
	return localFilePath(id), thumbFilePath(id)
}

// fillHTML 把读取 html 文件的内容，塞进 HTML (map[string]string)。
// 目的是方便以字符串的形式把 html 文件直接喂给 http.ResponseWriter.
func fillHTML() {
	filePaths, err := goutil.GetFilesByExt(staticFolder, ".html")
	if err != nil {
		panic(err)
	}

	for _, path := range filePaths {
		base := filepath.Base(path)
		name := strings.TrimSuffix(base, ".html")
		html, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		HTML[name] = string(html)
	}
}
