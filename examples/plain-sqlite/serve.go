//go:build ignore

package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/hellflame/ueditor"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, e := sql.Open("sqlite3", "resource.db")
	if e != nil {
		panic(e)
	}
	defer db.Close()

	demo, _ := os.ReadFile("demo.html")
	http.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Write(demo)
	})

	// create editor with storage
	editor := ueditor.NewEditor(nil, ueditor.NewSqliteStorage("uploads", db))

	// bind serve routes to editor backend & storage backend
	ueditor.BindHTTP(nil, nil, editor)

	port := ":8080"
	log.Print("浏览器访问 http://127.0.0.1" + port + "/demo")
	if e := http.ListenAndServe(port, nil); e != nil {
		log.Fatal(e)
	}
}
