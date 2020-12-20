package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/ahui2016/go-send/database"
	"github.com/ahui2016/goutil"
	"golang.org/x/net/webdav"
)

const (
	dataFolderName   = "gosend_data_folder"
	filesFolderName  = "files"
	databaseFileName = "gosend.db"
	configFileName   = "config"
	gosendFileExt    = ".send"
	thumbFileExt     = ".small"
	passwordMaxTry   = 5
	defaultPassword  = "abc"
	defaultAddress   = "127.0.0.1:80"
	webdavFolderName = "webdav"

	// 剪贴板文本消息上限
	defaultClipsLimit = 100

	// 99 days, for session
	maxAge = 99 * time.Hour * 24

	// databaseCapacity 控制数据库总容量，
	// maxBodySize 控制单个文件的体积。
	databaseCapacity = 1 << 30 // 1GB

	// 100 MB, for http.MaxBytesReader
	// 注意在 Nginx 的设置里进行相应的设置，例如 client_max_body_size 100m
	maxBodySize = 1024 * 1024 * 100

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
	webdavDir   = filepath.Join(dataDir, webdavFolderName)
	passwordTry = 0
	db          = new(database.DB)
	dav         = newDav(webdavDir)
)

// Config .
type Config struct {
	Password   string
	Address    string
	ClipsLimit int
}

func init() {
	goutil.MustMkdir(dataDir)
	goutil.MustMkdir(filesDir)
	goutil.MustMkdir(webdavDir)

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
		config = Config{
			defaultPassword,
			defaultAddress,
			defaultClipsLimit,
		}
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

func newDav(dirPath string) *webdav.Handler {
	return &webdav.Handler{
		Prefix:     "/" + webdavFolderName,
		FileSystem: webdav.Dir(dirPath),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WEBDAV [%s]: %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Printf("WEBDAV [%s]: %s \n", r.Method, r.URL)
			}
		},
	}
}
