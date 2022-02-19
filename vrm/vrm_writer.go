package vrm

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/qmuntal/gltf"
)

type createfs struct {
	fs.FS
	dir string
}

func (f *createfs) Create(name string) (io.WriteCloser, error) {
	return os.Create(f.dir + "/" + name)
}

// Write vrm file
func Write(doc *Document, w io.Writer, path string) error {
	e := gltf.NewEncoderFS(w, &createfs{os.DirFS(filepath.Dir(path)), path})
	e.AsBinary = true
	if err := e.Encode((*gltf.Document)(doc)); err != nil {
		return err
	}
	return nil
}
