package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/ahui2016/go-send/database"
	"github.com/ahui2016/goutil"
)

const (
	dataFolderName   = "gosend_data_folder"
	filesFolderName  = "files"
	databaseFileName = "gosend.db"
	configFileName   = "config"
	gosendFileExt    = ".send"
	thumbFileExt     = ".small"
	staticFolder     = "static"
	passwordMaxTry   = 5
	defaultPassword  = "abc"
	defaultAddress   = "127.0.0.1:80"

	// 99 days, for session
	maxAge = 60 * 60 * 24 * 99

	// databaseCapacity 控制数据库总容量，
	// maxBodySize 控制单个文件的体积。
	databaseCapacity = 1 << 30 // 1GB

	// 100 MB, for http.MaxBytesReader
	// 注意在 Nginx 的设置里进行相应的设置，例如 client_max_body_size 100m
	maxBodySize int64 = 1024 * 1024 * 100

	// 512 KB, 普通请求（没有文件的请求）需要限制更严格。
	defaultBodySize int64 = 1 << 19
)

var (
	config Config
)

var (
	dataDir     = filepath.Join(goutil.UserHomeDir(), dataFolderName)
	filesDir    = filepath.Join(dataDir, filesFolderName)
	dbPath      = filepath.Join(dataDir, databaseFileName)
	configPath  = filepath.Join(dataDir, configFileName)
	passwordTry = 0
	HTML        = make(map[string]string)
	db          = new(database.DB)
)

// Config .
type Config struct {
	Password string
	Address  string
}

func init() {
	goutil.MustMkdir(dataDir)
	goutil.MustMkdir(filesDir)

	fillHTML()
	setConfig()

	// open the db here, close the db in main().
	err := db.Open(maxAge, databaseCapacity, dbPath)
	goutil.CheckErrorPanic(err)
	log.Print(dbPath)
}

func setConfig() {
	configJSON, err := ioutil.ReadFile(configPath)

	// configPath 没有文件或内容为空
	if err != nil || len(configJSON) == 0 {
		config.Password = defaultPassword
		config.Address = defaultAddress
		configJSON, err := json.MarshalIndent(config, "", "    ")
		goutil.CheckErrorFatal(err)
		goutil.CheckErrorFatal(
			ioutil.WriteFile(configPath, configJSON, 0600))
		return
	}

	// configPath 有内容
	goutil.CheckErrorFatal(json.Unmarshal(configJSON, &config))
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
	goutil.CheckErrorFatal(err)

	for _, path := range filePaths {
		base := filepath.Base(path)
		name := strings.TrimSuffix(base, ".html")
		html, err := ioutil.ReadFile(path)
		goutil.CheckErrorFatal(err)
		HTML[name] = string(html)
	}
}
