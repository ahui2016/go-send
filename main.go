package main

import (
	"fmt"
	"github.com/ahui2016/goutil/graphics"
	"html"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ahui2016/go-send/model"
	"github.com/ahui2016/goutil"
	"github.com/ahui2016/goutil/zipper"
)

type (
	// Message .
	Message = model.Message
)

func main() {
	defer func() { _ = db.Close() }()

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	staticFS := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFS))

	filesFS := http.FileServer(http.Dir(filesDir))
	filesFS = http.StripPrefix("/files/", filesFS)
	http.Handle("/files/", checkLoginForFileServer(filesFS))

	// 本程序有一个简单的 webdav 功能，删除下面这行开头的双斜杠即可使用。
	// http.HandleFunc("/webdav/", maxBodyLimit(authWebDav(dav)))

	http.HandleFunc("/", homePage)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/api/login", bodyLimit(loginHandler))

	http.HandleFunc("/send-file", checkLogin(addFilePage))
	http.HandleFunc("/api/checksum", bodyLimit(checkLogin(checksumHandler)))
	http.HandleFunc("/api/upload-file", maxBodyLimit(checkLogin(uploadHandler)))

	http.HandleFunc("/messages", checkLogin(messagesPage))
	http.HandleFunc("/api/all", bodyLimit(checkLogin(getAllHandler)))
	http.HandleFunc("/api/add-text-msg", bodyLimit(checkLogin(addTextMsg)))
	http.HandleFunc("/api/delete", bodyLimit(checkLogin(deleteHandler)))

	http.HandleFunc("/api/update-datetime", bodyLimit(checkLogin(updateDatetime)))
	http.HandleFunc("/api/execute-command", bodyLimit(checkLogin(executeCommand)))
	http.HandleFunc("/api/total-size", bodyLimit(checkLogin(getTotalSize)))

	http.HandleFunc("/bookmarks", checkLogin(bookmarksPage))
	http.HandleFunc("/api/all-bookmarks", bodyLimit(checkLogin(getAllAnchors)))

	http.HandleFunc("/clips", checkLogin(clipsPage))
	http.HandleFunc("/api/all-clips", bodyLimit(checkLogin(getAllClips)))
	http.HandleFunc("/api/add-clip", bodyLimit(checkPassword(addClipMsg)))
	http.HandleFunc("/api/delete-clip", bodyLimit(checkLogin(deleteClip)))
	http.HandleFunc("/api/delete-all-clips", bodyLimit(checkLogin(deleteAllClips)))
	http.HandleFunc("/api/update-clip-datetime", bodyLimit(checkLogin(updateClipDatetime)))

	http.HandleFunc("/api/add-text", bodyLimit(checkPassword(addTextMsg)))
	http.HandleFunc("/api/last-text", bodyLimit(checkPassword(getLastText)))
	http.HandleFunc("/api/add-photo", maxBodyLimit(checkPassword(simpleUploadHandler)))

	log.Print(config.Address)
	log.Fatal(http.ListenAndServe(config.Address, nil))
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

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/icons/favicon.ico")
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/login.html")
}

func addFilePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/send-file.html")
}

func messagesPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/messages.html")
}

func bookmarksPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/bookmarks.html")
}

func clipsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/clips.html")
}

func addTextMsg(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()

	textMsg, ok := createAnchor(r.FormValue("text-msg"))
	message, err := db.InsertTextMsg(textMsg)
	if goutil.CheckErr(w, err, 500) {
		return
	}

	// 如果 ok, 表示 textMsg 是一个 anchor.
	if ok {
		message.FileType = model.GosendAnchor
		if goutil.CheckErr(w, db.DB.Save(message), 500) {
			return
		}
	}
	goutil.JsonResponse(w, message, 200)
}

// createAnchor 为了安全必须在 getTitle 成功后才生成 html,
// 否则原封不动返回 s。
func createAnchor(s string) (anchor string, ok bool) {
	var link, title string
	link, ok = isHttpURL(s)
	if ok {
		title, ok = getTitle(link)
		if ok {
			title = html.EscapeString(title)
			anchor = fmt.Sprintf(`<a href="%s">%s</a>`, link, title)
			return
		}
	}
	return s, false
}

// isHttpURL 当 s 是一个有效网址时返回该网址与 true, 否则返回 false.
func isHttpURL(s string) (addr string, ok bool) {
	reAddr := regexp.MustCompile(`^https?://[-a-zA-Z0-9_+~@#$%&=?/;:,.]+$`)
	addr = strings.TrimSpace(s)
	if !reAddr.MatchString(addr) {
		return "", false
	}
	return addr, true
}

// getTitle 获取一个有效网址 addr 的网页的 title.
func getTitle(addr string) (title string, ok bool) {
	res, err := http.Get(addr)
	if err != nil {
		return "", false
	}
	defer func() { _ = res.Body.Close() }()

	reTitle := regexp.MustCompile(`<title>(.+)</title>`)
	blob := make([]byte, 1024)
	for {
		_, err := res.Body.Read(blob)
		// 我服了，有的网站在 title 里加换行符
		headStr := strings.ReplaceAll(string(blob), "\n", " ")
		matches := reTitle.FindStringSubmatch(headStr)
		// 这个 matches 要么为空，要么包含两个元素
		if len(matches) >= 2 {
			return matches[1], true
		}

		// 由于 EOF 也属于错误，但即使 EOF 也有可能读取出一些数据，因此 err 延后处理。
		if err != nil {
			return "", false
		}
	}
}

func addClipMsg(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()

	textMsg := r.FormValue("text-msg")
	_, err := db.InsertClip(textMsg, config.ClipsLimit)
	if goutil.CheckErr(w, err, 500) {
		return
	}
}

func getAllHandler(w http.ResponseWriter, _ *http.Request) {
	all, err := db.AllByUpdatedAt()
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.JsonResponse(w, all, 200)
}

func getAllAnchors(w http.ResponseWriter, _ *http.Request) {
	all, err := db.AllAnchors()
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.JsonResponse(w, all, 200)
}

func getAllClips(w http.ResponseWriter, _ *http.Request) {
	all, err := db.AllClips()
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
	db.Lock()
	defer db.Unlock()

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

	if checkImage(w, message, fileContents) {
		return
	}
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
	if thumbFile, err := getThumbnail(r); err == nil {
		err = ioutil.WriteFile(thumbFilePath(message.ID), thumbFile, 0600)
		if goutil.CheckErr(w, err, 500) {
			return
		}
	}

	// 自动删除过期条目
	goutil.CheckErr(w, deleteExpiredItems(), 500)
}

func getThumbnail(r *http.Request) ([]byte, error) {
	file, _, err := r.FormFile("thumbnail")
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	return ioutil.ReadAll(file)
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

// checkImage 在 message 是图片是检查该图片能否正常使用，
// 如果不能正常使用，向 w 写入错误消息并返回 true 表示有错误。
func checkImage(w http.ResponseWriter, message *Message, img []byte) (hasErr bool) {
	if message.IsImage() {
		if _, err := graphics.Thumbnail(img, 0, 0); err != nil {
			goutil.JsonMessage(w, "该图片有问题，拒绝接收", 500)
			return true
		}
	}
	return false
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()

	id, ok := goutil.GetID(w, r)
	if !ok {
		return
	}
	err := goutil.DeleteFiles(getFileAndThumb(id))
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.CheckErr(w, db.Delete(id), 500)
}

func deleteClip(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()

	id, ok := goutil.GetID(w, r)
	if !ok {
		return
	}
	goutil.CheckErr(w, db.DeleteClip(id), 500)
}

func deleteAllClips(w http.ResponseWriter, _ *http.Request) {
	db.Lock()
	defer db.Unlock()
	goutil.CheckErr(w, db.DeleteAllClips(), 500)
}

func executeCommand(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()

	switch command := r.FormValue("command"); command {
	case "zip-all-files":
		message, err := zipAllFiles()
		if goutil.CheckErr(w, err, 500) {
			return
		}
		goutil.JsonResponse(w, message, 200)
	case "delete-all-files":
		goutil.CheckErr(w, deleteAllFiles(), 500)
	case "delete-10-files":
		goutil.CheckErr(w, deleteOldFiles(10), 500)
	case "delete-10-items":
		goutil.CheckErr(w, deleteOldItems(10), 500)
	case "delete-grey-items":
		err := deleteGreyItems()
		if goutil.ErrorContains(err, "not found") {
			goutil.JsonMessage(w, "暂时没有文件变灰", 404)
			return
		}
		goutil.CheckErr(w, err, 500)
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
	zipFilePath := localFilePath(message.ID)
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

// zipperFiles 将文件转换为 zipper.File 形式，会剔除 GosendZip, 避免重复打包。
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

func deleteOldFiles(n int) error {
	files, err := db.OldFiles(n)
	if err != nil {
		return err
	}
	return deleteItems(files)
}

func deleteOldItems(n int) error {
	items, err := db.OldItems(n)
	if err != nil {
		return err
	}
	return deleteItems(items)
}

func deleteGreyItems() error {
	items, err := db.GreyItems()
	if err != nil {
		return err
	}
	return deleteItems(items)
}

func deleteExpiredItems() error {
	items, err := db.ExpiredItems()
	if goutil.ErrorContains(err, "not found") {
		return nil
	}
	if err != nil {
		return err
	}
	return deleteItems(items)
}

func deleteItems(items []Message) error {
	if err := deleteFilesAndThumb(items); err != nil {
		return err
	}
	return db.DeleteMessages(items)
}

func deleteAllFiles() error {
	err1 := os.RemoveAll(filesDir)
	err2 := os.Mkdir(filesDir, 0700)
	err3 := db.DeleteAllFiles()
	return goutil.WrapErrors(err1, err2, err3)
}

func deleteFilesAndThumb(files []Message) error {
	var filePaths []string
	for _, file := range files {
		originFile, thumb := getFileAndThumb(file.ID)
		filePaths = append(filePaths, originFile, thumb)
	}
	return goutil.DeleteFiles(filePaths...)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if isLoggedIn(r) {
		goutil.JsonMessage(w, "already logged in", 200)
		return
	}

	if r.FormValue("password") != config.Password {
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
	db.Lock()
	defer db.Unlock()

	id, ok := goutil.GetID(w, r)
	if !ok {
		return
	}
	goutil.CheckErr(w, db.UpdateDatetime(id), 500)
}

func updateClipDatetime(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()

	id, ok := goutil.GetID(w, r)
	if !ok {
		return
	}
	goutil.CheckErr(w, db.UpdateClipDatetime(id), 500)
}

func getTotalSize(w http.ResponseWriter, _ *http.Request) {
	size, _ := db.GetTotalSize()
	resp := make(map[string]int64)
	resp["totalSize"] = size
	resp["capacity"] = databaseCapacity
	goutil.JsonResponse(w, resp, 200)
}

func getLastText(w http.ResponseWriter, _ *http.Request) {
	textMsg, err := db.LastTextMsg()
	if goutil.CheckErr(w, err, 500) {
		return
	}
	_, _ = fmt.Fprint(w, textMsg)
}

func simpleUploadHandler(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()

	header, contents, err := getHeaderContents(r)
	if goutil.CheckErr(w, err, 400) {
		return
	}

	message, err := db.NewFileMsg(header.Filename)
	if goutil.CheckErr(w, err, 500) {
		return
	}

	if checkImage(w, message, contents) {
		return
	}

	message.FileSize = header.Size

	// 至此，message 的全部内容都已经填充完毕，可以写入数据库。
	if goutil.CheckErr(w, db.Insert(message), 500) {
		return
	}

	// 数据库操作成功，保存文件（如果是图片，则顺便生成缩略图）。
	// 不可在数据库操作结束之前保存文件，因为数据库操作发生错误时不应保存文件。
	goutil.CheckErr(w, writeFile(message, contents), 500)
}

func getHeaderContents(r *http.Request) (
	header *multipart.FileHeader, contents []byte, err error) {
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = file.Close() }()

	// 将文件内容全部读入内存
	contents, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}
	return header, contents, nil
}
