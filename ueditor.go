package ueditor

import (
	"io"
	"mime/multipart"
	"path"
	"strings"
)

type UEditor struct {
	config    *Config
	storage   Storage
	srvPrefix string // resource servce prefix
	actions   Actions
}

const (
	ImageSaveBase  = "images"
	FileSaveBase   = "files"
	VideoSaveBase  = "videos"
	ScrawlSaveBase = "scrawls"
)

const (
	StateOK             = "SUCCESS"
	StateFileSizeExceed = "文件大小超出限制"
)

type UploadResp struct {
	State    string
	Url      string
	Title    string
	Original string
	Type     string
	Size     int
}

type ListResp struct {
	State string
	List  []FShard
	Start int
	Total int
}

type FShard struct {
	Url   string
	Mtime int
}

type Actions struct {
	Config string

	UploadImage string
	UploadFile  string

	UploadVideo  string
	Uploadscrawl string

	ListImages string
	ListFiles  string
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

	actions := Actions{
		Config: c.ConfigActionName,

		UploadImage:  c.ImageActionName,
		UploadFile:   c.FileActionName,
		UploadVideo:  c.VideoActionName,
		Uploadscrawl: c.ScrawlActionName,

		ListImages: c.ImageManagerActionName,
		ListFiles:  c.FileManagerActionName,
	}
	// do some config check

	return &UEditor{config: c, storage: s, actions: actions}
}

func (u *UEditor) onUploadFile(name string, f io.Reader) {}

func (u *UEditor) GetConfig() []byte {
	return LowerCamalMarshal(*u.config)
}

func (u *UEditor) SetSrvPrefix(prefix string) {
	u.srvPrefix = prefix
}

func (u *UEditor) GetActions() Actions {
	return u.actions
}

func (u *UEditor) SaveFile(prefix string, h *multipart.FileHeader, f io.Reader) UploadResp {
	resource, e := u.storage.Save(prefix, h, f)
	if e != nil {
		return UploadResp{
			State: e.Error(),
		}
	}
	base := u.srvPrefix
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

func (u *UEditor) ListFiles(prefix string, offset, size int) ListResp {
	files, total := u.storage.List(prefix, offset, size)
	tmp := make([]FShard, len(files))
	base := u.srvPrefix
	for idx, f := range files {
		// non external
		if !strings.Contains(f.Path, "://") {
			f.Path = path.Join(base, f.Path)
		}
		tmp[idx] = FShard{Url: f.Path, Mtime: f.Modify}
	}
	return ListResp{
		Total: total,
		State: StateOK,
		Start: offset,
		List:  tmp,
	}
}

func (u *UEditor) OnUploadImage(h *multipart.FileHeader, f io.Reader) UploadResp {
	if h.Size > int64(u.config.ImageMaxSize) {
		return UploadResp{State: StateFileSizeExceed}
	}
	mtype := h.Header.Get("Content-Type")
	if !strings.HasPrefix(mtype, "image/") || !isAllowedFileType(h.Filename, u.config.ImageAllowFiles) {
		return UploadResp{State: "非法图片类型"}
	}
	return u.SaveFile(ImageSaveBase, h, f)
}

func (u *UEditor) OnUploadScrawl(h *multipart.FileHeader, f io.Reader) UploadResp {
	if h.Size > int64(u.config.ScrawlMaxSize) {
		return UploadResp{State: StateFileSizeExceed}
	}
	return u.SaveFile(ScrawlSaveBase, h, f)
}

func (u *UEditor) OnUploadFile(h *multipart.FileHeader, f io.Reader) UploadResp {
	if h.Size > int64(u.config.FileMaxSize) {
		return UploadResp{State: StateFileSizeExceed}
	}
	if !isAllowedFileType(h.Filename, u.config.FileAllowFiles) {
		return UploadResp{State: "非法文件类型"}
	}
	return u.SaveFile(FileSaveBase, h, f)
}

func (u *UEditor) OnUploadVideo(h *multipart.FileHeader, f io.Reader) UploadResp {
	if h.Size > int64(u.config.VideoMaxSize) {
		return UploadResp{State: StateFileSizeExceed}
	}
	if !isAllowedFileType(h.Filename, u.config.VideoAllowFiles) {
		return UploadResp{State: "非法视频类型"}
	}
	return u.SaveFile(VideoSaveBase, h, f)
}

func (u *UEditor) OnListImages(offset, size int) ListResp {
	return u.ListFiles(ImageSaveBase, offset, size)
}

func (u *UEditor) OnListFiles(offset, size int) ListResp {
	return u.ListFiles(FileSaveBase, offset, size)
}

func (u *UEditor) ReadFile(path string) (meta *MetaInfo, raw []byte, e error) {
	return u.storage.Read(path)
}
