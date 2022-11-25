package ueditor

import (
	"io"
	"mime/multipart"
)

const hashAsDirIndent = 2

type FileInfo struct {
	Name   string
	Path   string
	Modify int
}

type MetaInfo struct {
	Filename string
	MimeType string
	Size     int64
}

type Storage interface {
	Save(prefix string, h *multipart.FileHeader, f io.Reader) (path string, e error)
	Read(path string) (meta *MetaInfo, content []byte, e error)
	List(prefix string, offset, limit int) (files []FileInfo, total int)
}

func cutHashToPath(raw string) (head string, tail string) {
	return raw[:hashAsDirIndent], raw[hashAsDirIndent:]
}
