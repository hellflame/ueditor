package ueditor

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const metaSuffix = ".meta"

type LocalStorage struct {
	Base string
}

func NewLocalStorage(base string) *LocalStorage {
	return &LocalStorage{Base: base}
}

func (l *LocalStorage) List(prefix string, offset, limit int) (files []FileInfo, total int) {
	saveDir := path.Join(l.Base, prefix)
	if exist, e := dirExist(saveDir); e != nil || !exist {
		return
	}
	metaFiles, e := filepath.Glob(path.Join(saveDir, "*"+metaSuffix))
	if e != nil {
		return
	}
	metaInfo := []FileInfo{}
	for _, f := range metaFiles {
		s, _ := os.Stat(f)
		h := multipart.FileHeader{}
		meta, _ := os.ReadFile(f)
		if e := json.Unmarshal(meta, &h); e != nil {
			continue
		}
		metaInfo = append(metaInfo, FileInfo{
			Modify: int(s.ModTime().Unix()), Name: h.Filename,
			Path: strings.TrimRight(strings.TrimLeft(f, l.Base), metaSuffix),
		})
	}
	sort.Slice(metaInfo, func(i, j int) bool {
		return metaInfo[i].Modify > metaInfo[j].Modify
	})
	total = len(metaInfo)
	if offset >= total-1 {
		return
	}
	if offset+limit >= total {
		return metaInfo[offset:], total
	}
	return metaInfo[offset : offset+limit], total
}

func (l *LocalStorage) Read(p string) (*MetaInfo, []byte, error) {
	contentPath := path.Join(l.Base, p)
	if !fileExist(contentPath) {
		return nil, nil, ErrFileMissing
	}
	metaPath := path.Join(l.Base, p+metaSuffix)
	if !fileExist(metaPath) {
		return nil, nil, ErrFileMetaMissing
	}
	meta, e := os.ReadFile(metaPath)
	if e != nil {
		return nil, nil, e
	}
	info := &MetaInfo{}
	if e := json.Unmarshal(meta, info); e != nil {
		return nil, nil, e
	}
	content, e := os.ReadFile(contentPath)
	return info, content, e
}

func (l *LocalStorage) Save(prefix string, h *multipart.FileHeader, f io.Reader) (string, error) {
	content, e := io.ReadAll(f)
	if e != nil {
		return "", e
	}
	contentHash := fmt.Sprintf("%x", md5.Sum(content))
	saveDir := path.Join(l.Base, prefix)
	contentPath := path.Join(saveDir, contentHash)
	metaPath := path.Join(saveDir, fmt.Sprintf("%s%s", contentHash, metaSuffix))
	if dirExist, e := dirExist(saveDir); !dirExist {
		if createE := os.MkdirAll(saveDir, os.ModePerm); createE != nil {
			return "", createE
		}
	} else if e != nil {
		return "", e
	}
	if !fileExist(contentPath) {
		if e := saveFileContent(contentPath, content); e != nil {
			return "", e
		}
	}
	if !fileExist(metaPath) {
		mtype := h.Header.Get("Content-Type")
		if mtype == "" {
			mtype = "application/octet-stream"
		}
		meta := MetaInfo{Filename: h.Filename, MimeType: mtype, Size: h.Size}
		content, e := json.Marshal(meta)
		if e != nil {
			return "", e
		}
		if e := saveFileContent(metaPath, content); e != nil {
			return "", e
		}
	}
	return path.Join(prefix, contentHash), nil
}
