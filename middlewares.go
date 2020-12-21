package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ahui2016/goutil"
	"github.com/gofiber/fiber/v2"
)

// setBodySize(fn, defaultBodySize)
func bodyLimit(fn http.HandlerFunc) http.HandlerFunc {
	return setBodySize(fn, defaultBodySize)
}

// setBodySize(fn, maxBodySize)
func maxBodyLimit(fn http.HandlerFunc) http.HandlerFunc {
	return setBodySize(fn, maxBodySize)
}

// 限制从前端传输过来的数据大小。
func setBodySize(fn http.HandlerFunc, max int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if goutil.CheckErr(w, checkContentLength(r, max), 500) {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, max)
		fn(w, r)
	}
}

// func checkContentLength(c *fiber.Ctx) error {

// }

// Check the Content-Length header immediately when the request comes in.
func checkContentLength(r *http.Request, length int64) error {
	if r.Header.Get("Content-Length") == "" {
		return nil
	}
	size, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return err
	}
	if size > length {
		return errors.New("file too large")
	}
	return nil
}

func checkLoginHtml(c *fiber.Ctx) error {
	if isLoggedOut(c) {
		if err := checkPasswordTry(c); err != nil {
			return err
		}
		return c.Redirect("/public/login.html")
	}
	return c.Next()
}

func checkLoginJson(c *fiber.Ctx) error {
	if isLoggedOut(c) {
		return jsonError(c, "Require Login", fiber.StatusUnauthorized)
	}
	return c.Next()
}

func checkPassword(c *fiber.Ctx) error {
	if c.FormValue("password") != config.Password {
		return jsonError(c, "Wrong Password", 400)
	}
	return nil
}

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
