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

func BindHTTP(mux *http.ServeMux, c *ServiceConfig, editor *UEditor) *http.ServeMux {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	// init config
	c = tidyConfig(c)
	srvPrefix := c.SrcServePrefix // => /resource/
	editor.SetSrvPrefix(srvPrefix)
	actions := editor.GetActions()

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
			f, h, e := r.FormFile("upfile")
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
			content, e := base64.StdEncoding.DecodeString(r.FormValue("upfile"))
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
			size, e := strconv.Atoi(r.URL.Query().Get("size"))
			if e != nil {
				panic("invalid size")
			}
			offset, e := strconv.Atoi(r.URL.Query().Get("start"))
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
			sendJsonPRespons(w, callback, resp)
			return
		}
		sendJsonResponse(w, resp)
	})
	return mux
}

func sendJsonResponse(w http.ResponseWriter, resp []byte) {
	w.Header().Add("Content-Type", "application/json")
	w.Write(resp)
}

func sendJsonPRespons(w http.ResponseWriter, callback string, resp []byte) {
	if !isFullAlpha(callback) {
		panic("invalid jsonp method")
	}
	resp = append(resp, []byte(");")...)
	resp = append([]byte(callback+"("), resp...)
	w.Header().Add("Content-Type", "application/javascript")
	w.Write(resp)
}
