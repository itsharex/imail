package assets

import (
	"bytes"
	"io/fs"
	"net/http"
	"os"
	"time"
)

type fileInfo struct {
	name string
	size int64
}

func (d fileInfo) Name() string {
	return d.name
}

func (d fileInfo) Size() int64 {
	return d.size
}

func (d fileInfo) Mode() os.FileMode {
	return os.FileMode(0644) | os.ModeDir
}

func (d fileInfo) ModTime() time.Time {
	return time.Time{}
}

func (d *fileInfo) IsDir() bool {
	return true
}

func (d fileInfo) Sys() interface{} {
	return nil
}

type file struct {
	name string
	*bytes.Reader

	children       []os.FileInfo
	childrenOffset int
}

func (f *file) Close() error {
	return nil
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	if len(f.children) == 0 {
		return nil, os.ErrNotExist
	}

	if count <= 0 {
		return f.children, nil
	}

	if f.childrenOffset+count > len(f.children) {
		count = len(f.children) - f.childrenOffset
	}
	offset := f.childrenOffset
	f.childrenOffset += count
	return f.children[offset : offset+count], nil
}

func (f *file) Stat() (os.FileInfo, error) {
	childCount := len(f.children)
	if childCount != 0 {
		return &fileInfo{
			name: f.name,
			size: int64(childCount),
		}, nil
	}

	data, err := ReadPublicFile(f.name)
	if err != nil {
		return nil, err
	}
	return &fileInfo{
		name: f.name,
		size: int64(len(data)),
	}, nil
}

type httpFileSystem struct{}

func (f *httpFileSystem) Open(name string) (http.File, error) {
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}

	data, err := ReadPublicFile(name)
	if err == nil {
		return &file{
			name:   name,
			Reader: bytes.NewReader(data),
		}, nil
	}

	subFS, err := PublicSub(name)
	if err != nil {
		return nil, err
	}

	readDirFS, ok := subFS.(fs.ReadDirFS)
	if !ok {
		return nil, os.ErrNotExist
	}

	entries, err := readDirFS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}

	return &file{
		name:     name,
		children: infos,
	}, nil
}

func NewFileSystem() http.FileSystem {
	return &httpFileSystem{}
}
