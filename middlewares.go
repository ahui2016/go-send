package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ahui2016/goutil"
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

func checkLoginForFileServer(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if passwordTry >= passwordMaxTry {
			_ = db.Close()
			http.NotFound(w, r)
			return
		}
		if isLoggedOut(r) {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func checkLogin(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isLoggedOut(r) {
			// 凡是以 "/api/" 开头的请求都返回 json 消息。
			if strings.HasPrefix(r.URL.Path, "/api/") {
				goutil.JsonRequireLogin(w)
				return
			}
			// 不是以 "/api/" 开头的都是页面。
			if checkPasswordTry(w) {
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		fn(w, r)
	}
}

func checkPassword(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("password") != config.Password {
			goutil.JsonMessage(w, "Wrong Password", 400)
			return
		}
		fn(w, r)
	}
}

/*
func handlerToFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
*/

func authWebDav(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("password") != config.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Access to WebDav"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func isLoggedIn(r *http.Request) bool {
	return db.Sess.Check(r)
}

func isLoggedOut(r *http.Request) bool {
	return !isLoggedIn(r)
}

func checkPasswordTry(w http.ResponseWriter) bool {
	if passwordTry >= passwordMaxTry {
		// log.Fatal()
		_ = db.Close()
		goutil.JsonMessage(w, "No more try. Input wrong password too many times.", 403)
		return true
	}
	return false
}
