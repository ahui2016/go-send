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
	app.Use(limiter.New(limiter.Config{
		Max: 300,
	}))

	app.Static("/public", "./public")

	app.Use(favicon.New(favicon.Config{File: "public/icons/favicon.ico"}))

	app.Use("/static", checkLoginHtml)
	app.Static("/static", "./static")
	app.Use("/files", checkLoginHtml)
	app.Static("/files", filesDir)

	app.Use("/home", checkLoginHtml)
	app.Get("/home", homePage)

	app.Get("/", redirectToHome)
	app.Post("/login", loginHandler)

	api := app.Group("/api", checkLoginJson)
	api.Get("/all", getAllHandler)
	api.Get("/total-size", getTotalSize)
	api.Get("/all-bookmarks", getAllAnchors)
	api.Get("/all-clips", getAllClips)
	api.Get("/delete-all-clips", deleteAllClips)
	api.Post("/add-text-msg", addTextMsg)
	api.Post("/delete", deleteHandler)
	api.Post("/update-datetime", updateDatetime)
	api.Post("/execute-command", executeCommand)
	api.Post("/delete-clip", deleteClip)
	api.Post("/update-clip-datetime", updateClipDatetime)

	cli := app.Group("/cli", checkPassword)
	cli.Get("/last-text", getLastText)
	cli.Post("/add-clip", addClipMsg)
	cli.Post("/add-text", addTextMsg)
	cli.Post("/add-photo", simpleUploadHandler)

	log.Print(config.Address)
	log.Fatal(app.Listen(config.Address))
}
