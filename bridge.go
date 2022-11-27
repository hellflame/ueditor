package ueditor

import (
	"io/fs"
	"strings"
)

type ServiceConfig struct {
	// ueditor编辑器基地址路径，其他 js, css资源等从该路径获取
	EditorHome string `default:"/assets/"`
	// 资源文件，默认使用内嵌资源文件
	Asset fs.FS
	// 编辑器所用上传等功能的接口地址，默认地址已与编辑器配置保持一致，如编辑器资源有变，此处需修改
	ApiPath string `default:"/ueditor-api"`
	// 本地提供上传资源的文件服务时使用该路径
	SrcServePrefix string `default:"/resource/"`
}

func tidyConfig(raw *ServiceConfig) *ServiceConfig {
	if raw == nil {
		raw = &ServiceConfig{}
	}
	applyDefault(raw)
	if !strings.HasSuffix(raw.EditorHome, "/") {
		raw.EditorHome += "/"
	}
	if !strings.HasSuffix(raw.SrcServePrefix, "/") {
		raw.SrcServePrefix += "/"
	}
	if raw.Asset == nil {
		raw.Asset = assets
	}
	return raw
}
