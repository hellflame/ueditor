//go:build !nobridge || (nobridge && onlygin)

package ueditor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func BindGin(engine *gin.Engine, c *ServiceConfig, editor *UEditor) {
	// init config
	c = tidyConfig(c)
	srvPrefix := c.SrcServePrefix // => /resource/
	editor.SetSrvPrefix(srvPrefix)
	actions := editor.GetActions()
	fieldName := editor.GetUploadFieldName()

	// editor home assets
	engine.GET(c.EditorHome+"*path", func(ctx *gin.Context) {
		path := ctx.Request.URL.Path[1:]
		file, e := c.Asset.Open(path)
		if e != nil {
			ctx.Writer.WriteHeader(404)
			return
		}
		defer file.Close()
		content, e := io.ReadAll(file)
		if e != nil {
			ctx.Writer.WriteHeader(500)
			return
		}
		ctype := mime.TypeByExtension(filepath.Ext(path))
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		ctx.Header("Content-Type", ctype)

		ctx.Writer.Write(content)
	})

	// editor uploaded resource service
	engine.GET(srvPrefix+"*path", func(ctx *gin.Context) {
		meta, raw, e := editor.ReadFile(ctx.Param("path"))
		writer := ctx.Writer
		if e != nil {
			writer.WriteHeader(404)
			return
		}
		writer.Header().Add("Content-Type", meta.MimeType)
		writer.Header().Add("Content-Disposition", "inline; filename=\""+url.QueryEscape(meta.Filename)+"\"")
		writer.Write(raw)
	})

	// editor api
	handler := func(ctx *gin.Context) {
		query := ctx.Request.URL.Query()
		action := query.Get("action")
		var resp []byte

		switch action {
		case actions.Config:
			resp = editor.GetConfig()
		case actions.UploadImage, actions.UploadFile, actions.UploadVideo:
			f, h, e := ctx.Request.FormFile(fieldName)
			if e != nil {
				panic("invalid file")
			}
			switch action {
			case actions.UploadImage:
				resp = LowerCamalMarshal(editor.OnUploadImage(h, f))
			case actions.UploadFile:
				resp = LowerCamalMarshal(editor.OnUploadFile(h, f))
			case actions.UploadVideo:
				resp = LowerCamalMarshal(editor.OnUploadVideo(h, f))
			}
		case actions.Uploadscrawl:
			content, e := base64.StdEncoding.DecodeString(ctx.Request.FormValue(fieldName))
			if e != nil {
				panic("invalid base64 => " + e.Error())
			}
			resp = LowerCamalMarshal(editor.OnUploadScrawl(
				&multipart.FileHeader{
					Filename: fmt.Sprintf("%d", time.Now().UnixNano()),
					Header: textproto.MIMEHeader{
						"Content-Type": []string{"image/png"},
					},
					Size: int64(len(content)),
				},
				bytes.NewBuffer(content)))

		case actions.ListImages, actions.ListFiles:
			size, e := strconv.Atoi(query.Get("size"))
			if e != nil {
				panic("invalid size")
			}
			offset, e := strconv.Atoi(query.Get("start"))
			if e != nil {
				offset = 0
			}
			if action == actions.ListImages {
				resp = LowerCamalMarshal(editor.OnListImages(offset, size))
			} else {
				resp = LowerCamalMarshal(editor.OnListFiles(offset, size))
			}
		default:
			panic("unknown action => " + action)
		}
		callback := query.Get("callback")
		if callback != "" {
			sendJsonPRespons(ctx.Writer, callback, resp)
			return
		}
		sendJsonResponse(ctx.Writer, resp)
	}
	engine.GET(c.ApiPath, handler).POST(c.ApiPath, handler)
}
