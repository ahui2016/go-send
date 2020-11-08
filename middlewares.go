package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ahui2016/goutil"
)

// 限制从前端传输过来的数据大小。
func setMaxBytes(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if goutil.CheckErr(w, checkContentLength(w, r, maxBytes), 500) {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		fn(w, r)
	}
}

// Check the Content-Length header immediately when the request comes in.
func checkContentLength(w http.ResponseWriter, r *http.Request, length int64) error {
	size, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return err
	}
	if size > maxBytes {
		return errors.New("File Too Large")
	}
	return nil
}

func checkLoginForFileServer(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if passwordTry >= passwordMaxTry {
			db.Close()
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
			fmt.Fprint(w, HTML["login"])
			return
		}
		fn(w, r)
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
		db.Close()
		goutil.JsonMessage(w, "No more try. Input wrong password too many times.", 403)
		return true
	}
	return false
}
