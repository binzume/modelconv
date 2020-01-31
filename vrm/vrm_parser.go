package vrm

import (
	"io"
	"path/filepath"

	"github.com/qmuntal/gltf"
)

// Parse vrm data
func Parse(r io.Reader, path string) (*VRMDocument, error) {
	var doc gltf.Document
	dec := gltf.NewDecoder(r).WithReadHandler(&gltf.RelativeFileHandler{Dir: filepath.Dir(path)})
	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}
	return (*VRMDocument)(&doc), nil
}
