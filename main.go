package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/add-file", addFilePage)

	addr := "127.0.0.1:80"
	log.Print(addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func addFilePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, HTML["add-file"])
}
