package main

import (
	"log"
	"net/http"
	"os"

	"github.com/hellflame/ueditor"
)

func main() {
	demo, _ := os.ReadFile("demo.html")
	http.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Write(demo)
	})

	// create editor with storage
	editor := ueditor.NewEditor(nil, ueditor.NewLocalStorage("uploads"))
	// bind serve routes to editor backend & storage backend
	ueditor.BindHTTP(nil, nil, editor)

	port := ":8080"
	log.Print("浏览器访问 http://" + port + "/demo")
	if e := http.ListenAndServe(port, nil); e != nil {
		log.Fatal(e)
	}
}
