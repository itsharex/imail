package assets

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/midoks/imail/internal/tools"
)

type fileSystem struct {
	files []macaron.TemplateFile
}

func (fs *fileSystem) ListFiles() []macaron.TemplateFile {
	return fs.files
}

func (fs *fileSystem) Get(name string) (io.Reader, error) {
	for i := range fs.files {
		if fs.files[i].Name()+fs.files[i].Ext() == name {
			return bytes.NewReader(fs.files[i].Data()), nil
		}
	}
	return nil, fmt.Errorf("file %q not found", name)
}

func NewTemplateFileSystem(dir, customDir string) macaron.TemplateFileSystem {
	if dir != "" && !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	var files []macaron.TemplateFile
	names, _ := WalkTemplates()
	for _, name := range names {
		if !strings.HasPrefix(name, dir) {
			continue
		}

		var err error
		var data []byte
		fpath := path.Join(customDir, name)
		if tools.IsFile(fpath) {
			data, err = ioutil.ReadFile(fpath)
		} else {
			data, err = ReadTemplateFile(name)
		}
		if err != nil {
			panic(err)
		}

		name = strings.TrimPrefix(name, dir)
		ext := path.Ext(name)
		name = strings.TrimSuffix(name, ext)
		files = append(files, macaron.NewTplFile(name, data, ext))
	}
	return &fileSystem{files: files}
}
