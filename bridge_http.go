package ueditor

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func BindHTTP(mux *http.ServeMux, c *ServiceConfig, editor *UEditor) *http.ServeMux {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	// init config
	c = tidyConfig(c)
	srvPrefix := c.SrcServePrefix // => /resource/
	editor.SetSrvPrefix(srvPrefix)
	actionConfig, actionUpImage, actionUpFile, actionLsImage, actionLsFile := editor.GetActions()

	// editor home asserts
	mux.Handle(c.EditorHome, http.FileServer(http.FS(c.Asset)))

	// editor uploaded resource service
	mux.HandleFunc(srvPrefix, func(w http.ResponseWriter, r *http.Request) {
		meta, raw, e := editor.ReadFile(strings.TrimLeft(r.URL.Path, srvPrefix))
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
		case actionConfig:
			resp = editor.GetConfig()
		case actionUpImage, actionUpFile:
			f, h, e := r.FormFile("upfile")
			if e != nil {
				panic("invalid file")
			}

			if action == actionUpImage {
				resp = lowerCamalMarshal(editor.OnUploadImage(h, f))
			} else {
				resp = lowerCamalMarshal(editor.OnUploadFile(h, f))
			}
		case actionLsImage, actionLsFile:
			size, e := strconv.Atoi(r.URL.Query().Get("size"))
			if e != nil {
				panic("invalid size")
			}
			offset, e := strconv.Atoi(r.URL.Query().Get("start"))
			if e != nil {
				offset = 0
			}
			if action == actionLsImage {
				resp = lowerCamalMarshal(editor.OnListImages(offset, size))
			} else {
				resp = lowerCamalMarshal(editor.OnListFiles(offset, size))
			}
		default:
			panic("unknown action")
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
