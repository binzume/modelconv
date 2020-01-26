package vrm

import (
	"io"
	"path/filepath"

	"github.com/qmuntal/gltf"
)

// Write vrm file
func Write(doc *VRMDocument, w io.Writer, path string) error {
	e := gltf.NewEncoder(w).WithWriteHandler(&gltf.RelativeFileHandler{Dir: filepath.Dir(path)})
	e.AsBinary = true
	if err := e.Encode((*gltf.Document)(doc)); err != nil {
		return err
	}
	return nil
}
