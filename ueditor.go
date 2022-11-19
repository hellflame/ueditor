package ueditor

import (
	"io"
	"mime/multipart"
	"path"
	"strings"
)

type UEditor struct {
	config  *Config
	storage Storage
}

type Config struct {
	ConfigActionName string `default:"config"`

	ImageActionName     string   `default:"up-image"`
	ImageFieldName      string   `default:"upfile"`
	ImageMaxSize        int      `default:"5000000"`
	ImageAllowFiles     []string `default:".png|.jpg|.jpeg|.gif|.bmp"`
	ImageCompressEnable bool     `default:"true"`
	ImageCompressBorder int      `default:"1600"`
	ImageInsertAlign    string   `default:"none"`
	ImageUrlPrefix      string
	// ImagePathFormat     string

	ScrawlActionName string `default:"up-scrawl"`
	ScrawlFieldName  string `default:"upfile"`
	// ScrawlPathFormat  string
	ScrawlMaxSize     int `default:"2000000"`
	ScrawlUrlPrefix   string
	ScrawlInsertAlign string   `default:"none"`
	ScrawlAllowFiles  []string `default:".png"`

	SnapscreenActionName string `default:"up-image"`
	// SnapscreenPathFormat  string
	SnapscreenUrlPrefix   string
	SnapscreenMaxSize     int    `default:"2000000"`
	SnapscreenInsertAlign string `default:"none"`

	CatcherLocalDomain []string
	CatcherActionName  string `default:"catch-image"`
	CatcherFieldName   string `default:"source"`
	// CatcherPathFormat  string
	CatcherUrlPrefix  string
	CatcherMaxSize    int      `default:"5000000"`
	CatcherAllowFiles []string `default:".png|.jpg|.jpeg|.gif|.bmp"`

	VideoActionName string `default:"up-video"`
	VideoFieldName  string `default:"upfile"`
	// VideoPathFormatstring string
	VideoUrlPrefix  string
	VideoMaxSize    int      `default:"100000000"`
	VideoAllowFiles []string `default:".flv|.swf|.mkv|.avi|.rm|.rmvb|.mpeg|.mpg|.ogg|.ogv|.mov|.wmv|.mp4|.webm|.mp3|.wav|.mid"`

	FileActionName string `default:"up-file"`
	FileFieldName  string `default:"upfile"`
	// FilePathFormat string
	FileUrlPrefix  string
	FileMaxSize    int      `default:"50000000"`
	FileAllowFiles []string `default:".png|.jpg|.jpeg|.gif|.bmp|.flv|.swf|.mkv|.avi|.rm|.rmvb|.mpeg|.mpg|.ogg|.ogv|.mov|.wmv|.mp4|.webm|.mp3|.wav|.mid|.rar|.zip|.tar|.gz|.7z|.bz2|.cab|.iso|.doc|.docx|.xls|.xlsx|.ppt|.pptx|.pdf|.txt|.md|.xml"`

	ImageManagerActionName string `default:"list-image"`
	// ImageManagerListPath    string
	ImageManagerListSize    int `default:"20"`
	ImageManagerUrlPrefix   string
	ImageManagerInsertAlign string   `default:"none"`
	ImageManagerAllowFiles  []string `default:".png|.jpg|.jpeg|.gif|.bmp"`

	FileManagerActionName string `default:"list-file"`
	// FileManagerListPath   string
	FileManagerUrlPrefix  string
	FileManagerListSize   int      `default:"20"`
	FileManagerAllowFiles []string `default:".png|.jpg|.jpeg|.gif|.bmp|.flv|.swf|.mkv|.avi|.rm|.rmvb|.mpeg|.mpg|.ogg|.ogv|.mov|.wmv|.mp4|.webm|.mp3|.wav|.mid|.rar|.zip|.tar|.gz|.7z|.bz2|.cab|.iso|.doc|.docx|.xls|.xlsx|.ppt|.pptx|.pdf|.txt|.md|.xml"`
}

func NewEditor(c *Config, s Storage) *UEditor {
	if c == nil {
		c = &Config{}
	}
	applyDefault(c)
	// do some config check

	return &UEditor{config: c, storage: s}
}

func (u *UEditor) onUploadFile(name string, f io.Reader) {}

func (u *UEditor) GetConfig() []byte {
	return lowerCamalMarshal(*u.config)
}

func (u *UEditor) GetActions() (c, ui, fi, li, lf string) {
	config := u.config
	c = config.ConfigActionName
	ui = config.ImageActionName
	fi = config.FileActionName
	li = config.ImageManagerActionName
	lf = config.FileManagerActionName
	return
}

func (u *UEditor) SaveFile(base, prefix string, h *multipart.FileHeader, f io.Reader) UploadResp {
	resource, e := u.storage.Save(prefix, h, f)
	if e != nil {
		return UploadResp{
			State: e.Error(),
		}
	}
	// non external
	if !strings.Contains(resource, "://") {
		resource = path.Join(base, resource)
	}
	return UploadResp{
		State:    StateOK,
		Url:      resource,
		Original: h.Filename,
		Size:     int(h.Size),
	}
}

func (u *UEditor) OnUploadImage(srvPrefix string, h *multipart.FileHeader, f io.Reader) UploadResp {
	return u.SaveFile(srvPrefix, "images/", h, f)
}

func (u *UEditor) OnUploadFile(srvPrefix string, h *multipart.FileHeader, f io.Reader) UploadResp {
	return u.SaveFile(srvPrefix, "files/", h, f)
}

func (u *UEditor) OnListImages(offset, size int) ListResp {
	return ListResp{}
}

func (u *UEditor) OnListFiles(offset, size int) ListResp {
	return ListResp{}
}

func (u *UEditor) ReadFile(path string) (meta *MetaInfo, raw []byte, e error) {
	return u.storage.Read(path)
}
