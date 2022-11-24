package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hellflame/ueditor"
)

func main() {
	router := mux.NewRouter()
	demo, _ := os.ReadFile("demo.html")
	router.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Write(demo)
	})

	// create editor with storage
	editor := ueditor.NewEditor(nil, ueditor.NewLocalStorage("uploads"))
	// bind serve routes to editor backend & storage backend
	ueditor.BindMux(router, nil, editor)

	port := ":8080"
	log.Print("浏览器访问 http://127.0.0.1" + port + "/demo")
	if e := http.ListenAndServe(port, router); e != nil {
		log.Fatal(e)
	}
}
