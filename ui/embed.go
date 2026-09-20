package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var files embed.FS

func Handler() http.Handler {
	// 固定目录由 go:embed 在编译时检查，fs.Sub 的路径始终有效。
	content, _ := fs.Sub(files, "dist")
	return http.FileServer(http.FS(content))
}
