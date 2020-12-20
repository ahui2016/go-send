package main

import (
	"io/ioutil"

	"github.com/ahui2016/go-send/model"
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
			return err
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
		return err
	}

	filename := c.FormValue("filename")
	message, err := db.NewFileMsg(filename)
	if err != nil {
		return err
	}
	message.Checksum = c.FormValue("checksum")
	message.FileSize = int64(len(fileContents))

	if err := checkImage(c, message, fileContents); err != nil {
		return jsonError(c, err.Error(), 400)
	}

	// 至此，message 的全部内容都已经填充完毕，可以写入数据库。
	if err := db.Insert(message); err != nil {
		return err
	}

	// 数据库操作成功，保存文件（如果是图片，则顺便生成缩略图）。
	// 不可在数据库操作结束之前保存文件，因为数据库操作发生错误时不应保存文件。
	if err := writeFile(message, fileContents); err != nil {
		return err
	}

	// 如果前端传来缩略图，就保存下来。如果没有，则忽略不管。
	if thumbFile, err := getThumbnail(c); err == nil {
		err = ioutil.WriteFile(thumbFilePath(message.ID), thumbFile, 0600)
		if err != nil {
			return err
		}
	}

	// 自动删除过期条目
	return deleteExpiredItems()
}

func addTextMsg(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	textMsg, ok := createAnchor(c.FormValue("text-msg"))
	message, err := db.InsertTextMsg(textMsg)
	if err != nil {
		return err
	}

	// 如果 ok, 表示 textMsg 是一个 anchor.
	if ok {
		message.FileType = model.GosendAnchor
		if err := db.DB.Save(message); err != nil {
			return err
		}
	}
	return c.JSON(message)
}

func deleteHandler(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id, err := getID(c)
	if err != nil {
		return jsonError(c, err.Error(), 400)
	}
	if err := goutil.DeleteFiles(getFileAndThumb(id)); err != nil {
		return err
	}
	return db.Delete(id)
}

func updateDatetime(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id, err := getID(c)
	if err != nil {
		return jsonError(c, err.Error(), 400)
	}
	return db.UpdateDatetime(id)
}

func executeCommand(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	switch command := c.FormValue("command"); command {
	case "zip-all-files":
		message, err := zipAllFiles()
		if err != nil {
			return err
		}
		return c.JSON(message)
	case "delete-all-files":
		if err := deleteAllFiles(); err != nil {
			return err
		}
	case "delete-10-files":
		if err := deleteOldFiles(10); err != nil {
			return err
		}
	case "delete-10-items":
		if err := deleteOldItems(10); err != nil {
			return err
		}
	case "delete-grey-items":
		err := deleteGreyItems()
		if errorContains(err, "not found") {
			return jsonError(c, "暂时没有文件变灰", 404)
		}
		if err != nil {
			return err
		}
	default:
		return jsonError(c, "unknown command", 400)
	}
	return nil
}

func getTotalSize(c *fiber.Ctx) error {
	size, _ := db.GetTotalSize()
	resp := make(map[string]int64)
	resp["totalSize"] = size
	resp["capacity"] = databaseCapacity
	return c.JSON(resp)
}

func getAllAnchors(c *fiber.Ctx) error {
	all, err := db.AllAnchors()
	if err != nil {
		return err
	}
	return c.JSON(all)
}

func getAllClips(c *fiber.Ctx) error {
	all, err := db.AllClips()
	if err != nil {
		return err
	}
	return c.JSON(all)
}

func addClipMsg(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	textMsg := c.FormValue("text-msg")
	_, err := db.InsertClip(textMsg, config.ClipsLimit)
	return err
}

func deleteClip(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id, err := getID(c)
	if err != nil {
		return jsonError(c, err.Error(), 400)
	}
	return db.DeleteClip(id)
}

func deleteAllClips(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()
	return db.DeleteAllClips()
}

func updateClipDatetime(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id, err := getID(c)
	if err != nil {
		return jsonError(c, err.Error(), 400)
	}
	return db.UpdateClipDatetime(id)
}

func getLastText(c *fiber.Ctx) error {
	textMsg, err := db.LastTextMsg()
	if err != nil {
		return err
	}
	return c.SendString(textMsg)
}

func simpleUploadHandler(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	header, contents, err := getFileHeaderContents(c, "file")
	if err != nil {
		return err
	}

	message, err := db.NewFileMsg(header.Filename)
	if err != nil {
		return err
	}
	if err := checkImage(c, message, contents); err != nil {
		return jsonError(c, err.Error(), 400)
	}

	message.FileSize = header.Size

	// 至此，message 的全部内容都已经填充完毕，可以写入数据库。
	if err := db.Insert(message); err != nil {
		return err
	}

	// 数据库操作成功，保存文件（如果是图片，则顺便生成缩略图）。
	// 不可在数据库操作结束之前保存文件，因为数据库操作发生错误时不应保存文件。
	return writeFile(message, contents)
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	err = c.Status(code).JSON(fiber.Map{"message": err.Error()})
	if err != nil {
		// In case the SendFile fails
		return c.Status(500).SendString("Internal Server Error")
	}
	return nil
}
