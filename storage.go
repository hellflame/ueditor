package ueditor

import (
	"io"
	"mime/multipart"
)

const hashAsDirIndent = 2

// 文件列表数据结构，与 http 服务输出相关联
type FileInfo struct {
	Name   string
	Path   string
	Modify int
}

// 文件元信息数据结构，关联元数据存储与 http 服务输出
type MetaInfo struct {
	Filename string
	MimeType string
	Size     int64
}

// 存储接口，用于支持多种存储方式
type Storage interface {
	Save(prefix string, h *multipart.FileHeader, f io.Reader) (path string, e error)
	Read(path string) (meta *MetaInfo, content []byte, e error)
	List(prefix string, offset, limit int) (files []FileInfo, total int)
}

// 当有额外存储设施 (sqlite, mysql等) 时，可分散资源存储
func cutToPieces(raw string) (head string, tail string) {
	return raw[:hashAsDirIndent], raw[hashAsDirIndent:]
}
