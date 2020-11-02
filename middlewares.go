package main

import "net/http"

// 限制从前端传输过来的数据大小。
func setMaxBytes(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		fn(w, r)
	}
}
