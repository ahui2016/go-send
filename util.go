package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ahui2016/go-send/model"
	"github.com/ahui2016/goutil"
	"github.com/ahui2016/goutil/graphics"
	"github.com/ahui2016/goutil/zipper"
	"github.com/gofiber/fiber/v2"
)

type (
	// Message .
	Message = model.Message
)

func jsonMessage(c *fiber.Ctx, msg string) error {
	return c.Status(200).JSON(fiber.Map{"message": msg})
}

func jsonMsgOK(c *fiber.Ctx) error {
	return jsonMessage(c, "OK")
}

func jsonError(c *fiber.Ctx, msg string, status int) error {
	return c.Status(status).JSON(fiber.Map{"message": msg})
}

/*
func jsonErr500(c *fiber.Ctx, err error) error {
	return jsonError(c, err.Error(), 500)
}
*/

func getFileHeaderContents(c *fiber.Ctx, key string) (
	header *multipart.FileHeader, contents []byte, err error) {

	header, err = c.FormFile(key)
	if err != nil {
		return nil, nil, err
	}
	file, err := header.Open()
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	// 将文件内容全部读入内存
	contents, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}
	return header, contents, nil
}

/*
func getFile(c *fiber.Ctx, key string) (multipart.File, error) {
	fileHeader, err := c.FormFile(key)
	if err != nil {
		return nil, err
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	return file, nil
}
*/

// getFileContents gets contents from FormFile("file").
// It also verifies the file has not been corrupted.
func getFileContents(c *fiber.Ctx) ([]byte, error) {
	_, contents, err := getFileHeaderContents(c, "file")
	if err != nil {
		return nil, err
	}

	// 根据文件内容生成 checksum 并检查其是否正确
	if Sha256Hex(contents) != c.FormValue("checksum") {
		return nil, errors.New("checksums do not match")
	}
	return contents, nil
}

// Sha256Hex .
func Sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// checkImage 在 message 是图片是检查该图片能否正常使用，
func checkImage(c *fiber.Ctx, message *Message, img []byte) error {
	if message.IsImage() {
		if _, err := graphics.Thumbnail(img, 0, 0); err != nil {
			return errors.New("该图片有问题，拒绝接收")
		}
	}
	return nil
}

func getThumbnail(c *fiber.Ctx) ([]byte, error) {
	_, contents, err := getFileHeaderContents(c, "thumbnail")
	return contents, err
}

// getFormValue checks if the c.FormValue(key) is empty or not,
// if it is empty, write error message and return false;
// if it is not empty, return the id and true.
func getFormValue(c *fiber.Ctx, key string) (string, error) {
	value := c.FormValue(key)
	if value == "" {
		return "", errors.New(key + " is empty")
	}
	return value, nil
}

func getID(c *fiber.Ctx) (string, error) {
	return getFormValue(c, "id")
}

// errorContains returns NoCaseContains(err.Error(), substr)
// Returns false if err is nil.
func errorContains(err error, substr string) bool {
	if err == nil {
		return false
	}
	return noCaseContains(err.Error(), substr)
}

// noCaseContains reports whether substr is within s case-insensitive.
func noCaseContains(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
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

	reTitle := regexp.MustCompile(`<title ?.*>(.+)</title>`)
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
