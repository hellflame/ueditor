//go:build !nostorage || (nostorage && onlyminio)

package ueditor

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

type minioStorage struct {
	client *minio.Client
	base   string
	ctx    context.Context
}

func NewMinioStorage(client *minio.Client) *minioStorage {
	base := client.EndpointURL().String()
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return &minioStorage{client: client, base: base, ctx: context.Background()}
}

func (m *minioStorage) prepBucket(name string) error {
	exist, e := m.client.BucketExists(m.ctx, name)
	if e != nil || exist {
		return e
	}
	if e = m.client.MakeBucket(m.ctx, name, minio.MakeBucketOptions{}); e != nil {
		return e
	}
	return m.client.SetBucketPolicy(m.ctx, name,
		`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::`+name+`/*"]}]}`)

}

func (m *minioStorage) List(prefix string, offset, limit int) (files []FileInfo, total int) {
	ctx, cancel := context.WithCancel(m.ctx)
	defer cancel()
	for f := range m.client.ListObjects(ctx, prefix, minio.ListObjectsOptions{Recursive: true}) {
		files = append(files, FileInfo{
			Path:   m.base + path.Join(prefix, f.Key),
			Modify: int(f.LastModified.Unix()),
		})
	}
	total = len(files)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Modify > files[j].Modify
	})
	if offset >= total-1 {
		return
	}
	if offset+limit >= total {
		return files[offset:], total
	}
	return files[offset : offset+limit], total
}

// minio files should be served by minio server
func (m *minioStorage) Read(p string) (*MetaInfo, []byte, error) {
	return nil, nil, ErrNotImpled
}

func (m *minioStorage) Save(prefix string, h *multipart.FileHeader, f io.Reader) (string, error) {
	if e := m.prepBucket(prefix); e != nil {
		return "", e
	}
	content := make([]byte, 0, 409600)
	reader := bufio.NewReaderSize(f, 409600)
	for {
		if len(content) == cap(content) {
			// Add more capacity (let append pick how much).
			content = append(content, 0)[:len(content)]
		}
		n, err := reader.Read(content[len(content):cap(content)])
		content = content[:len(content)+n]
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
	}
	fname := url.PathEscape(h.Filename)
	present := time.Now()
	contentHash := fmt.Sprintf("%x", md5.Sum(content))
	full := fmt.Sprintf("%d/%d/%d/%s-%s", present.Year(), present.Month(), present.Day(), contentHash[:10], fname)
	_, e := m.client.PutObject(m.ctx, prefix, full, bytes.NewBuffer(content), h.Size, minio.PutObjectOptions{
		ContentType: h.Header.Get("Content-Type"),
	})
	if e != nil {
		return "", e
	}
	return m.base + path.Join(prefix, full), nil
}
