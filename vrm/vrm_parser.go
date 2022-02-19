package vrm

import (
	"io"
	"os"
	"path/filepath"

	"github.com/qmuntal/gltf"
)

// Parse vrm data
func Parse(r io.Reader, path string) (*Document, error) {
	var doc gltf.Document
	dec := gltf.NewDecoderFS(r, os.DirFS(filepath.Dir(path)))
	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}
	return (*Document)(&doc), nil
}
