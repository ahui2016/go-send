package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	defer func() { _ = db.Close() }()

	app := fiber.New(fiber.Config{
		BodyLimit:    maxBodySize,
		Concurrency:  5,
		ErrorHandler: errorHandler,
	})

	// app.Use(maxBodyLimit)
	app.Use(responseNoCache)
	app.Use(limiter.New(limiter.Config{
		Max: 300,
	}))

	app.Static("/public", "./public")

	app.Use(favicon.New(favicon.Config{File: "public/icons/favicon.ico"}))

	app.Use("/static", checkLoginHTML)
	app.Static("/static", "./static")
	app.Use("/files", checkLoginHTML)
	app.Static("/files", filesDir)

	app.Get("/", redirectToHome)
	app.Use("/home", checkLoginHTML)
	app.Get("/home", homePage)
	app.Post("/login", loginHandler)

	api := app.Group("/api", checkLoginJSON)
	api.Get("/all", getAllHandler)
	api.Get("/total-size", getTotalSize)
	api.Get("/all-bookmarks", getAllAnchors)
	api.Get("/all-clips", getAllClips)
	api.Get("/delete-all-clips", deleteAllClips)
	api.Post("/checksum", checksumHandler)
	api.Post("/upload-file", uploadHandler)
	api.Post("/add-text-msg", addTextMsg)
	api.Post("/delete", deleteHandler)
	api.Post("/update-datetime", updateDatetime)
	api.Post("/execute-command", executeCommand)
	api.Post("/delete-clip", deleteClip)
	api.Post("/update-clip-datetime", updateClipDatetime)

	cli := app.Group("/cli", checkPassword)
	cli.Post("/last-text", getLastText)
	cli.Post("/add-clip", addClipMsg)
	cli.Post("/add-text", addTextMsg)
	cli.Post("/add-photo", simpleUploadHandler)

	log.Fatal(app.Listen(config.Address))
}
