//go:build ignore

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/hellflame/ueditor"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	demo, _ := os.ReadFile("demo.html")
	http.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Write(demo)
	})

	// use your own minio instance
	client, e := minio.New("192.168.1.8:9000", &minio.Options{
		Creds: credentials.NewStaticV4("minioadmin", "minioadmin", ""),
	})
	if e != nil {
		panic(e)
	}
	// create editor with storage
	editor := ueditor.NewEditor(nil, ueditor.NewMinioStorage(client))
	// bind serve routes to editor backend & storage backend
	ueditor.BindHTTP(nil, nil, editor)

	port := ":8080"
	log.Print("浏览器访问 http://127.0.0.1" + port + "/demo")
	if e := http.ListenAndServe(port, nil); e != nil {
		log.Fatal(e)
	}
}
