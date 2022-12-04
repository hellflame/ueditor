//go:build !nostorage || (nostorage && onlygorm)

package ueditor

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormStorage struct {
	base string
	db   *gorm.DB
}

// gorm 模型定义 - 表结构
type Resources struct {
	Category string `gorm:"uniqueIndex:ukey,size:50"`
	Hash     string `gorm:"uniqueIndex:ukey,size:50"`
	Filename string `gorm:"size:256"`
	Mimetype string `gorm:"size:50"`
	Size     int64
	Created  int64
	Chunks   int `gorm:"default:0"`
}

// NewGormStorage create a *gormStorage instance which implemented Storage interface
//
// File info is stored in given database instance, using table 'resources'
func NewGormStorage(base string, db *gorm.DB) *gormStorage {
	// do some db check
	mg := db.Migrator()
	if !mg.HasTable(&Resources{}) {
		if e := mg.CreateTable(&Resources{}); e != nil {
			panic(e)
		}
	}
	return &gormStorage{base, db}
}

func (s *gormStorage) List(prefix string, offset, limit int) (files []FileInfo, total int) {
	result := []Resources{}
	s.db.Where("category = ?", prefix).Select("filename", "hash", "created").Order("created desc").Limit(limit).Offset(offset).Find(&result)

	for _, src := range result {
		head, tail := cutToPieces(src.Hash)
		files = append(files, FileInfo{
			Name:   src.Filename,
			Modify: int(src.Size),
			Path:   path.Join(prefix, head, tail),
		})
	}
	var cnt int64
	s.db.Where("category = ?", prefix).Count(&cnt)
	total = int(cnt)
	return
}

func (s *gormStorage) Save(prefix string, h *multipart.FileHeader, f io.Reader) (string, error) {
	content, e := io.ReadAll(f)
	if e != nil {
		return "", e
	}
	contentHash := fmt.Sprintf("%x", md5.Sum(content))
	head, tail := cutToPieces(contentHash)
	saveDir := path.Join(s.base, prefix, head)
	contentPath := path.Join(saveDir, tail)

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

	mtype := h.Header.Get("Content-Type")
	if mtype == "" {
		mtype = "application/octet-stream"
	}
	target := Resources{
		Category: prefix,
		Filename: h.Filename,
		Mimetype: mtype,
		Hash:     contentHash,
		Size:     h.Size,
		Created:  time.Now().Unix(),
	}
	e = s.db.Clauses(clause.OnConflict{
		DoNothing: true,
		Columns:   []clause.Column{{Name: "category"}, {Name: "hash"}},
	}).Create(&target).Error

	return path.Join(prefix, head, tail), e
}

func (s *gormStorage) Read(p string) (*MetaInfo, []byte, error) {
	contentPath := path.Join(s.base, p)
	if !fileExist(contentPath) {
		return nil, nil, ErrFileMissing
	}
	cuts := strings.Split(p, "/")
	if len(cuts) < 2 {
		return nil, nil, ErrPathMalform
	}
	category := cuts[0]
	contentHash := strings.Join(cuts[1:], "")

	fetch := Resources{}
	e := s.db.Where("category = ? and hash = ?", category, contentHash).First(&fetch).Error
	if e != nil {
		return nil, nil, e
	}
	meta := &MetaInfo{
		Filename: fetch.Filename,
		MimeType: fetch.Mimetype,
		Size:     fetch.Size,
	}
	content, e := os.ReadFile(contentPath)
	return meta, content, e
}
