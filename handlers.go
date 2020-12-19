package main

import (
	"io/ioutil"

	"github.com/ahui2016/goutil"
	"github.com/gofiber/fiber/v2"
)

func redirectToHome(c *fiber.Ctx) error {
	return c.Redirect("/home")
}

func homePage(c *fiber.Ctx) error {
	return c.SendFile("./static/messages.html")
}

func loginHandler(c *fiber.Ctx) error {
	if isLoggedIn(c) {
		return jsonMessage(c, "already logged in")
	}

	if c.FormValue("password") != config.Password {
		passwordTry++
		if err := checkPasswordTry(c); err != nil {
			return jsonErr500(c, err)
		}
		return jsonError(c, "Wrong Password", 400)
	}

	passwordTry = 0
	return db.SessionSet(c)
}

func getAllHandler(c *fiber.Ctx) error {
	all, err := db.AllByUpdatedAt()
	if err != nil {
		return err
	}
	return c.JSON(all)
}

func checksumHandler(c *fiber.Ctx) error {
	hashHex := c.FormValue("hashHex")
	var message Message
	err := db.DB.One("Checksum", hashHex, &message)

	if err != nil && err.Error() != "not found" {
		return jsonError(c, err.Error(), 500)
	}

	// 找不到，表示没有冲突。
	if err != nil && err.Error() == "not found" {
		return jsonMsgOK(c)
	}

	// err == nil, 正常找到已存在 hashHex, 表示发生文件冲突。
	return jsonError(c, "Checksum Already Exists", 400)
}

func uploadHandler(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	fileContents, err := getFileContents(c)
	if err != nil {
		return jsonErr500(c, err)
	}

	filename := c.FormValue("filename")
	message, err := db.NewFileMsg(filename)
	if err != nil {
		return jsonErr500(c, err)
	}
	message.Checksum = c.FormValue("checksum")
	message.FileSize = int64(len(fileContents))

	if err := checkImage(c, message, fileContents); err != nil {
		return jsonErr500(c, err)
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
