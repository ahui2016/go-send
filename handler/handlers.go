package handler

import (
	"github.com/ahui2016/goutil"
	"github.com/gofiber/fiber/v2"
)

func homePage(c *fiber.Ctx) error {
	return c.Redirect("/messages.html")
}

func loginHandler(c *fiber.Ctx) error {
	if isLoggedIn(c) {
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
