package ueditor

import (
	"io/fs"
	"strings"
)

type ServiceConfig struct {
	EditorHome     string `default:"/assets/"`
	Asset          fs.FS
	ApiPath        string `default:"/ueditor-api"`
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
