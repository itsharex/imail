package assets

import (
	"embed"
	"io/fs"
	"path"
)

//go:embed conf
var ConfFS embed.FS

//go:embed templates
var TemplatesFS embed.FS

//go:embed public
var PublicFS embed.FS

// ReadConfFile reads a file from the embedded conf filesystem.
func ReadConfFile(name string) ([]byte, error) {
	return ConfFS.ReadFile(path.Join("conf", name))
}

// ReadTemplateFile reads a file from the embedded templates filesystem.
func ReadTemplateFile(name string) ([]byte, error) {
	return TemplatesFS.ReadFile(path.Join("templates", name))
}

// ReadPublicFile reads a file from the embedded public filesystem.
func ReadPublicFile(name string) ([]byte, error) {
	return PublicFS.ReadFile(path.Join("public", name))
}

// ConfDir returns the embedded conf filesystem.
func ConfDir() fs.FS {
	sub, _ := fs.Sub(ConfFS, "conf")
	return sub
}

// TemplatesDir returns the embedded templates filesystem.
func TemplatesDir() fs.FS {
	sub, _ := fs.Sub(TemplatesFS, "templates")
	return sub
}

// PublicDir returns the embedded public filesystem.
func PublicDir() fs.FS {
	sub, _ := fs.Sub(PublicFS, "public")
	return sub
}

// PublicSub returns a sub-filesystem of the embedded public filesystem.
func PublicSub(dir string) (fs.FS, error) {
	return fs.Sub(PublicFS, path.Join("public", dir))
}

// WalkConf walks the conf filesystem and returns all file paths.
func WalkConf() ([]string, error) {
	return walkFS(ConfFS, "conf")
}

// WalkTemplates walks the templates filesystem and returns all file paths.
func WalkTemplates() ([]string, error) {
	return walkFS(TemplatesFS, "templates")
}

// WalkPublic walks the public filesystem and returns all file paths.
func WalkPublic() ([]string, error) {
	return walkFS(PublicFS, "public")
}

func walkFS(fsys embed.FS, root string) ([]string, error) {
	var files []string
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			// Remove the root prefix to get relative path
			relPath := path[len(root)+1:]
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}
