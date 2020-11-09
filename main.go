package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ahui2016/go-send/common"
	"github.com/ahui2016/go-send/model"
	"github.com/ahui2016/goutil"
	"github.com/ahui2016/goutil/zipper"
)

type (
	// Message .
	Message = model.Message
)

func main() {
	defer db.Close()

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	filesFS := http.FileServer(http.Dir(filesDir))
	filesFS = http.StripPrefix("/files/", filesFS)
	http.Handle("/files/", checkLoginForFileServer(filesFS))

	http.HandleFunc("/", homePage)
	http.HandleFunc("/api/login", bodyLimit(loginHandler))

	http.HandleFunc("/send-file", checkLogin(addFilePage))
	http.HandleFunc("/api/checksum",
		withDB(bodyLimit(checkLogin(checksumHandler))))
	http.HandleFunc("/api/upload-file",
		withDB(maxBodyLimit(checkLogin(uploadHandler))))

	http.HandleFunc("/messages", checkLogin(messagesPage))
	http.HandleFunc("/api/add-text-msg",
		withDB(bodyLimit(checkLogin(addTextMsg))))
	http.HandleFunc("/api/all", withDB(checkLogin(getAllHandler)))
	http.HandleFunc("/api/delete",
		withDB(bodyLimit(checkLogin(deleteHandler))))

	http.HandleFunc("/api/update-datetime",
		withDB(bodyLimit(checkLogin(updateDatetime))))

	http.HandleFunc("/api/execute-command",
		bodyLimit(checkLogin(executeCommand)))

	http.HandleFunc("/totalSize", totalSize)

	addr := "127.0.0.1:80"
	log.Print(addr)
	log.Print(dbPath)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func totalSize(w http.ResponseWriter, r *http.Request) {
	size, _ := db.GetTotalSize()
	resp := make(map[string]int64)
	resp["totalSize"] = size
	resp["capacity"] = common.DatabaseCapacity
	goutil.JsonResponse(w, resp, 200)
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
	message, err := db.InsertTextMsg(textMsg)
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.JsonResponse(w, message, 200)
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
	if goutil.CheckErr(w, writeFile(message, fileContents), 500) {
		return
	}

	// 如果前端传来缩略图，就保存下来。如果没有，则忽略不管。
	thumbFile, err := getThumbnail(r)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(common.ThumbFilePath(filesDir, message.ID), thumbFile, 0600)
	goutil.CheckErr(w, err, 500)
}

func getThumbnail(r *http.Request) ([]byte, error) {
	file, _, err := r.FormFile("thumbnail")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func writeFile(message *Message, fileContents []byte) error {
	file, thumb := common.GetFileAndThumb(filesDir, message.ID)
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
	err := goutil.DeleteFiles(common.GetFileAndThumb(filesDir, id))
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.CheckErr(w, db.Delete(id), 500)
}

func executeCommand(w http.ResponseWriter, r *http.Request) {
	switch command := r.FormValue("command"); command {
	case "zip-all-files":
		message, err := zipAllFiles()
		if goutil.CheckErr(w, err, 500) {
			return
		}
		goutil.JsonResponse(w, message, 200)
	case "delete-all-files":
		if goutil.CheckErr(w, deleteAllFiles(), 500) {
			return
		}
	case "delete-5-files":
		if goutil.CheckErr(w, deleteOldFiles(5), 500) {
			return
		}
	default:
		goutil.JsonMessage(w, "unknown command", 400)
	}
}

// zipAllFiles 把全部文件打包，打包后的文件将会在列表中显示，因此用户可以下载和删除。
// zipAllFiles 会自动剔除使用 zipAllFiles 等函数打包的文件，避免重复打包。
func zipAllFiles() (message *Message, err error) {
	message, err = db.NewZipMsg("gosend_all_files")
	if err != nil {
		return
	}
	allFiles, err := db.AllFiles()
	if err != nil {
		return
	}
	zipFilePath := common.LocalFilePath(filesDir, message.ID)
	err = zipper.Create(zipFilePath, zipperFiles(allFiles))
	if err != nil {
		return
	}
	stat, err := os.Lstat(zipFilePath)
	if err != nil {
		return
	}
	message.FileSize = stat.Size()
	err = db.Insert(message)
	return
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
			Path: common.LocalFilePath(filesDir, message.ID),
		}
		files = append(files, file)
	}
	return
}

func deleteOldFiles(n int) error {
	files, err := db.OldFiles(n)
	if err != nil {
		return err
	}
	if err := deleteFilesAndThumb(files); err != nil {
		return err
	}
	return db.DeleteOldFiles(n)
}

func deleteAllFiles() error {
	allFiles, err := db.AllFiles()
	if err != nil {
		return err
	}
	if err := deleteFilesAndThumb(allFiles); err != nil {
		return err
	}
	return db.DeleteAllFiles()
}

func deleteFilesAndThumb(files []Message) error {
	var filePaths []string
	for _, file := range files {
		originFile, thumb := common.GetFileAndThumb(filesDir, file.ID)
		filePaths = append(filePaths, originFile, thumb)
	}
	return goutil.DeleteFiles(filePaths...)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if isLoggedIn(r) {
		goutil.JsonMessage(w, "Already logged in.", 400)
		return
	}

	password := r.FormValue("password")
	if password != defaultPassword {
		passwordTry++
		if checkPasswordTry(w) {
			return
		}
		goutil.JsonMessage(w, "Wrong Password", 400)
		return
	}

	passwordTry = 0
	db.Sess.Add(w, goutil.NewID())
}

func updateDatetime(w http.ResponseWriter, r *http.Request) {
	id, ok := goutil.GetID(w, r)
	if !ok {
		return
	}
	goutil.CheckErr(w, db.UpdateDatetime(id), 500)
}
