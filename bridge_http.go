//go:build !nobridge || (nobridge && onlyhttp)

package ueditor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// 绑定路由到系统库 http
func BindHTTP(mux *http.ServeMux, c *ServiceConfig, editor *UEditor) *http.ServeMux {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	// init config
	c = tidyConfig(c)
	srvPrefix := c.SrcServePrefix // => /resource/
	editor.SetSrvPrefix(srvPrefix)
	actions := editor.GetActions()
	fieldName := editor.GetUploadFieldName()

	// editor home assets
	mux.Handle(c.EditorHome, http.FileServer(http.FS(c.Asset)))

	// editor uploaded resource service
	mux.HandleFunc(srvPrefix, func(w http.ResponseWriter, r *http.Request) {
		meta, raw, e := editor.ReadFile(strings.TrimPrefix(r.URL.Path, srvPrefix))
		if e != nil {
			w.WriteHeader(404)
			return
		}
		w.Header().Add("Content-Type", meta.MimeType)
		w.Header().Add("Content-Disposition", "inline; filename=\""+url.QueryEscape(meta.Filename)+"\"")
		w.Write(raw)
	})

	// editor api
	mux.HandleFunc(c.ApiPath, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		action := query.Get("action")
		var resp []byte
		switch action {
		case actions.Config:
			resp = editor.GetConfig()
		case actions.UploadImage, actions.UploadFile, actions.UploadVideo:
			f, h, e := r.FormFile(fieldName)
			if e != nil {
				sendError(w, "invalid file")
				return
			}
			defer f.Close()
			switch action {
			case actions.UploadImage:
				resp = LowerCamelMarshal(editor.OnUploadImage(h, f))
			case actions.UploadFile:
				resp = LowerCamelMarshal(editor.OnUploadFile(h, f))
			case actions.UploadVideo:
				resp = LowerCamelMarshal(editor.OnUploadVideo(h, f))
			}
		case actions.Uploadscrawl:
			content, e := base64.StdEncoding.DecodeString(r.FormValue(fieldName))
			if e != nil {
				sendError(w, "invalid base64 => "+e.Error())
				return
			}
			resp = LowerCamelMarshal(editor.OnUploadScrawl(
				&multipart.FileHeader{
					Filename: fmt.Sprintf("%d", time.Now().UnixNano()),
					Header: textproto.MIMEHeader{
						"Content-Type": []string{"image/png"},
					},
					Size: int64(len(content)),
				},
				bytes.NewBuffer(content)))

		case actions.ListImages, actions.ListFiles:
			size, e := strconv.Atoi(r.URL.Query().Get("size"))
			if e != nil {
				sendError(w, "invalid size")
				return
			}
			offset, e := strconv.Atoi(r.URL.Query().Get("start"))
			if e != nil {
				offset = 0
			}
			if action == actions.ListImages {
				resp = LowerCamelMarshal(editor.OnListImages(offset, size))
			} else {
				resp = LowerCamelMarshal(editor.OnListFiles(offset, size))
			}
		default:
			sendError(w, "unknown action => "+action)
			return
		}

		callback := query.Get("callback")
		if callback != "" {
			SendJsonPResponse(w, callback, resp)
			return
		}
		SendJsonResponse(w, resp)
	})
	return mux
}
