//go:build ignore

package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hellflame/ueditor"
)

func main() {
	router := gin.Default()

	router.StaticFile("/demo", "demo.html")

	// create editor with storage
	editor := ueditor.NewEditor(nil, ueditor.NewLocalStorage("uploads"))
	// bind serve routes to editor backend & storage backend
	ueditor.BindGin(router, nil, editor)

	port := ":8080"
	log.Print("浏览器访问 http://127.0.0.1" + port + "/demo")

	if e := router.Run(port); e != nil {
		log.Fatal(e)
	}
}
