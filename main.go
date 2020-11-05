package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ahui2016/go-send/model"
	"github.com/ahui2016/goutil"
	"github.com/ahui2016/goutil/zipper"
)

type (
	Message = model.Message
)

func main() {
	defer db.Close()

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	filesFS := http.FileServer(http.Dir(filesDir))
	http.Handle("/files/", http.StripPrefix("/files/", filesFS))

	http.HandleFunc("/", homePage)

	http.HandleFunc("/send-file", addFilePage)
	http.HandleFunc("/api/checksum", checksumHandler)
	http.HandleFunc("/api/upload-file", setMaxBytes(uploadHandler))

	http.HandleFunc("/messages", messagesPage)
	http.HandleFunc("/api/add-text-msg", setMaxBytes(addTextMsg))
	http.HandleFunc("/api/all", getAllHandler)
	http.HandleFunc("/api/delete", deleteHandler)

	http.HandleFunc("/api/zip-all-files", zipAllHandler)
	http.HandleFunc("/api/delete-all-files", deleteAllHandler)

	addr := "127.0.0.1:80"
	log.Print(addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		fallthrough
	case "/home":
		http.Redirect(w, r, "/messages", 302)
	default:
		http.NotFound(w, r)
	}
}

func addFilePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, HTML["send-file"])
}

func messagesPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, HTML["messages"])
}

func addTextMsg(w http.ResponseWriter, r *http.Request) {
	textMsg := strings.TrimSpace(r.FormValue("text-msg"))
	if textMsg == "" {
		goutil.JsonMessage(w, "the message is empty", 400)
	}
	message, err := db.NewTextMsg(textMsg)
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.CheckErr(w, db.Insert(message), 500)
}

func getAllHandler(w http.ResponseWriter, r *http.Request) {
	all, err := db.AllByUpdatedAt()
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.JsonResponse(w, all, 200)
}

func checksumHandler(w http.ResponseWriter, r *http.Request) {
	hashHex := r.FormValue("hashHex")
	var message Message
	err := db.DB.One("Checksum", hashHex, &message)

	if err != nil && err.Error() != "not found" {
		goutil.JsonMessage(w, err.Error(), 500)
		return
	}

	// 找不到，表示没有冲突。
	if err != nil && err.Error() == "not found" {
		goutil.JsonMsgOK(w)
		return
	}

	// err == nil, 正常找到已存在 hashHex, 表示发生文件冲突。
	goutil.JsonMessage(w, "Checksum Already Exists", 400)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fileContents, err := goutil.GetFileContents(r)
	if goutil.CheckErr(w, err, 400) {
		return
	}

	filename := r.FormValue("filename")
	message, err := db.NewFileMsg(filename)
	if goutil.CheckErr(w, err, 500) {
		return
	}
	message.Checksum = r.FormValue("checksum")
	message.FileSize = int64(len(fileContents))

	// 至此，message 的全部内容都已经填充完毕，可以写入数据库。
	if goutil.CheckErr(w, db.Insert(message), 500) {
		return
	}

	// 数据库操作成功，保存文件（如果是图片，则顺便生成缩略图）。
	// 不可在数据库操作结束之前保存文件，因为数据库操作发生错误时不应保存文件。
	goutil.CheckErr(w, writeFile(message, fileContents), 500)
}

func writeFile(message *Message, fileContents []byte) error {
	file, thumb := getFileAndThumb(message.ID)
	err := ioutil.WriteFile(file, fileContents, 0600)
	if err != nil {
		return err
	}

	// 如果是图片, 一律生成缩略图
	if message.IsImage() {
		err := goutil.BytesToThumb(fileContents, thumb)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := goutil.GetID(w, r)
	if !ok {
		return
	}
	err := goutil.DeleteFiles(getFileAndThumb(id))
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.CheckErr(w, db.DB.DeleteStruct(&Message{ID: id}), 500)
}

func zipAllHandler(w http.ResponseWriter, r *http.Request) {
	goutil.CheckErr(w, zipAllFiles(), 500)
}

// zipAllFiles 把全部文件打包，打包后的文件将会在列表中显示，因此用户可以下载和删除。
// zipAllFiles 会自动剔除使用 zipAllFiles 等函数打包的文件，避免重复打包。
func zipAllFiles() error {
	message, err := db.NewZipMsg("gosend_all_files.zip")
	if err != nil {
		return err
	}
	allFiles, err := db.AllFiles()
	if err != nil {
		return err
	}
	zipFilePath := localFilePath(message.ID)
	err = zipper.Create(zipFilePath, zipperFiles(allFiles))
	if err != nil {
		return err
	}
	stat, err := os.Lstat(zipFilePath)
	if err != nil {
		return err
	}
	message.FileSize = stat.Size()
	return db.Save(message)
}

// zipperFiles 会自动剔除使用 GosendZip, 避免重复打包。
func zipperFiles(fileMessages []Message) (files []zipper.File) {
	for i := range fileMessages {
		message := fileMessages[i]
		if message.FileType == model.GosendZip {
			continue
		}
		file := zipper.File{
			Name: message.FileName,
			Path: localFilePath(message.ID),
		}
		files = append(files, file)
	}
	return
}

func deleteAllHandler(w http.ResponseWriter, r *http.Request) {
	allFiles, err := db.AllFiles()
	if goutil.CheckErr(w, err, 500) {
		return
	}
	var filePaths []string
	for _, file := range allFiles {
		originFile, thumb := getFileAndThumb(file.ID)
		filePaths = append(filePaths, originFile, thumb)
	}
	if goutil.CheckErr(w, goutil.DeleteFiles(filePaths...), 500) {
		return
	}
	goutil.CheckErr(w, db.DeleteAllFiles(), 500)
}
