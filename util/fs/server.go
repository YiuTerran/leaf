package fs

import (
	"net/http"
	"strings"
)

//加强版静态文件服务
//如果打开文件夹里面有index.html则返回内容，否则直接返回文件
type NeuteredFileSystem struct {
	FileSystem http.FileSystem
}

func (nfs NeuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.FileSystem.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := nfs.FileSystem.Open(index); err != nil {
			return nil, err
		}
	}
	return f, nil
}
