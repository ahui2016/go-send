package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ahui2016/goutil"
)

func main() {
	defer db.Close()

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", homePage)

	http.HandleFunc("/add-file", addFilePage)
	http.HandleFunc("/messages", messagesPage)
	http.HandleFunc("/api/add-text-msg", setMaxBytes(addTextMsg))
	http.HandleFunc("/api/all", getAllHandler)

	addr := "127.0.0.1:80"
	log.Print(addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		fallthrough
	case "/home":
		http.Redirect(w, r, "/messages", 302)
	default:
		http.NotFound(w, r)
	}
}

func addFilePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, HTML["add-file"])
}

func messagesPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, HTML["messages"])
}

func addTextMsg(w http.ResponseWriter, r *http.Request) {
	textMsg := strings.TrimSpace(r.FormValue("text-msg"))
	if textMsg == "" {
		goutil.JsonMessage(w, "the message is empty", 400)
	}
	message, err := db.NewTextMsg(textMsg)
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.CheckErr(w, db.Insert(message), 500)
}

func getAllHandler(w http.ResponseWriter, r *http.Request) {
	all, err := db.AllByUpdatedAt()
	if goutil.CheckErr(w, err, 500) {
		return
	}
	goutil.JsonResponse(w, all, 200)
}
