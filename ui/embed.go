package ui

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed dist
var files embed.FS

func Handler() http.Handler {
	// 固定目录由 go:embed 在编译时检查，fs.Sub 的路径始终有效。
	content, _ := fs.Sub(files, "dist")
	fileServer := http.FileServer(http.FS(content))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		_, err := fs.Stat(content, name)
		if errors.Is(err, fs.ErrNotExist) &&
			(r.Method == http.MethodGet || r.Method == http.MethodHead) &&
			path.Ext(name) == "" && !strings.HasPrefix(name, "assets/") {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
			r.URL.RawPath = ""
		}
		fileServer.ServeHTTP(w, r)
	})
}
