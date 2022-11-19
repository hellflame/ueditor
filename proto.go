package ueditor

const (
	StateOK             = "SUCCESS"
	StateUploadExceed   = "文件大小超出 upload_max_filesize 限制"
	StateFileSizeExceed = "文件大小超出 MAX_FILE_SIZE 限制"
	StateFileIncomplete = "文件未被完整上传"
	StateFileNotUpload  = "没有文件被上传"
	StateFileEmpty      = "上传文件为空"
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
