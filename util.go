package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/ioutil"

	"github.com/ahui2016/goutil/graphics"
	"github.com/gofiber/fiber/v2"
)

func jsonMessage(c *fiber.Ctx, msg string) error {
	return c.JSON(fiber.Map{"message": msg})
}

func jsonMsgOK(c *fiber.Ctx) error {
	return jsonMessage(c, "OK")
}

func jsonError(c *fiber.Ctx, msg string, status int) error {
	return c.Status(status).JSON(fiber.Map{"message": msg})
}

func jsonErr500(c *fiber.Ctx, err error) error {
	return jsonError(c, err.Error(), 500)
}

func isLoggedIn(c *fiber.Ctx) bool {
	return db.SessionCheck(c)
}

func isLoggedOut(c *fiber.Ctx) bool {
	return !isLoggedIn(c)
}

func checkPasswordTry(c *fiber.Ctx) error {
	if passwordTry >= passwordMaxTry {
		_ = db.Close()
		msg := "No more try. Input wrong password too many times."
		return errors.New(msg)
	}
	return nil
}

// GetFileContents gets contents from FormFile("file").
// It also verifies the file has not been corrupted.
func getFileContents(c *fiber.Ctx) ([]byte, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 将文件内容全部读入内存
	contents, err := ioutil.ReadAll(file)
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
// 如果不能正常使用，向 w 写入错误消息并返回 true 表示有错误。
func checkImage(c *fiber.Ctx, message *Message, img []byte) error {
	if message.IsImage() {
		if _, err := graphics.Thumbnail(img, 0, 0); err != nil {
			return errors.New("该图片有问题，拒绝接收")
		}
	}
	return nil
}
