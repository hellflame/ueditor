//go:build !nostorage || (nostorage && onlysqlite)

package ueditor

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"
)

type sqliteStorage struct {
	base string
	db   *sql.DB

	tableName string
}

// NewSqliteStorage create a *sqliteStorage instance which implemented Storage interface
//
// File info is storged in given slite database, using table 'resources'
func NewSqliteStorage(base string, db *sql.DB) *sqliteStorage {
	tableName := "resources"
	// do some db check
	if _, e := db.Exec(`create table if not exists ` + tableName + ` (
		category varchar(50),
		filename varchar(256),
		mimetype varchar(50),
		size bigint,
		created bigint,
		hash varchar(50),
		chunks int,
		primary key (category, hash)
	)`); e != nil {
		panic(e)
	}

	return &sqliteStorage{base, db, tableName}
}

func (s *sqliteStorage) List(prefix string, offset, limit int) (files []FileInfo, total int) {
	rows, e := s.db.Query("select filename, hash, created from "+
		s.tableName+" where category = ? order by created desc limit ? offset ?",
		prefix, limit, offset)
	if e != nil {
		panic(e)
	}
	for rows.Next() {
		var fname, hash string
		var created int64
		if e := rows.Scan(&fname, &hash, &created); e != nil {
			panic(e)
		}
		head, tail := cutHashToPath(hash)
		files = append(files, FileInfo{
			Name:   fname,
			Modify: int(created),
			Path:   path.Join(prefix, head, tail),
		})
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	if e := s.db.QueryRow("select count(1) from "+s.tableName+" where category = ?",
		prefix).Scan(&total); e != nil {
		panic(e)
	}
	return
}

func (s *sqliteStorage) Save(prefix string, h *multipart.FileHeader, f io.Reader) (string, error) {
	content, e := io.ReadAll(f)
	if e != nil {
		return "", e
	}
	contentHash := fmt.Sprintf("%x", md5.Sum(content))
	head, tail := cutHashToPath(contentHash)
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
	if _, e := s.db.Exec("insert or ignore into "+s.tableName+
		" (category, filename, mimetype, hash, size, created) values (?,?,?,?,?,?)",
		prefix, h.Filename, mtype, contentHash, h.Size, time.Now().Unix()); e != nil {
		panic(e)
	}

	return path.Join(prefix, head, tail), nil
}

func (s *sqliteStorage) Read(p string) (*MetaInfo, []byte, error) {
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
	meta := &MetaInfo{}
	if e := s.db.QueryRow("select filename, mimetype, size from "+s.tableName+
		" where category = ? and hash = ?", category, contentHash).Scan(&meta.Filename, &meta.MimeType, &meta.Size); e != nil {
		return nil, nil, e
	}
	content, e := os.ReadFile(contentPath)
	return meta, content, e
}
