package main

import (
	"github.com/gofiber/fiber/v2"
)

func homePage(c *fiber.Ctx) error {
	return c.Redirect("/messages.html")
}

func loginHandler(c *fiber.Ctx) error {
	if isLoggedIn2(c) {
		return jsonMessage(c, "already logged in")
	}

	if c.FormValue("password") != config.Password {
		passwordTry++
		if err := checkPasswordTry2(c); err != nil {
			return err
		}
		return jsonError(c, "Wrong Password", 400)
	}

	passwordTry = 0
	return db.SessionSet(c)
}
