package main

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

/*
func maxBodyLimit(c *fiber.Ctx) error {
	if err := checkContentLength(c, maxBodySize); err != nil {
		return c.Status(413).JSON(err.Error())
	}
	return nil
}

func checkContentLength(c *fiber.Ctx, length int) error {
	if c.Request().Header.ContentLength() > length {
		return errors.New("Requset Entity Too Large")
	}
	return nil
}
*/

func responseNoCache(c *fiber.Ctx) error {
	c.Response().Header.Set(
		fiber.HeaderCacheControl,
		"no-store, no-cache",
	)
	return c.Next()
}

func checkLoginHTML(c *fiber.Ctx) error {
	if isLoggedOut(c) {
		if err := checkPasswordTry(c); err != nil {
			return err
		}
		return c.SendFile("./public/login.html")
	}
	return c.Next()
}

func checkLoginJSON(c *fiber.Ctx) error {
	if isLoggedOut(c) {
		return jsonError(c, "Require Login", fiber.StatusUnauthorized)
	}
	return c.Next()
}

func checkPassword(c *fiber.Ctx) error {
	if c.FormValue("password") != config.Password {
		return jsonError(c, "Wrong Password", 400)
	}
	return c.Next()
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

/*
func authWebDav(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Access to WebDav"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if password != config.Password {
			http.Error(w, "WabDAV: wrong password", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}
*/
