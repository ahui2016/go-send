package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ahui2016/go-send/database"
	"github.com/ahui2016/go-send/util"
)

const (
	staticFolder = "static"
)

var (
	passwordTry = 0
	HTML        = make(map[string]string)
	db          = new(database.DB)
)

func init() {
	fillHTML()
}

// fillHTML 把读取 html 文件的内容，塞进 HTML (map[string]string)。
// 目的是方便以字符串的形式把 html 文件直接喂给 http.ResponseWriter.
func fillHTML() {
	filePaths, err := util.FilesInDir(staticFolder, ".html")
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
