package vrm

import (
	"io"
	"path/filepath"

	"github.com/qmuntal/gltf"
)

type VRMDocument gltf.Document

func (doc *VRMDocument) VRMExt() *VRMExt {
	if v, ok := doc.Extensions[ExtensionName].(*VRMExt); ok {
		return v
	}
	ext := &VRMExt{MaterialProperties: []interface{}{}}
	if doc.Extensions == nil {
		doc.Extensions = gltf.Extensions{}
	}
	doc.ExtensionsUsed = append(doc.ExtensionsUsed, ExtensionName)
	doc.Extensions[ExtensionName] = ext
	return ext
}

func (doc *VRMDocument) Title() string {
	return doc.VRMExt().Meta.Title
}

func (doc *VRMDocument) Author() string {
	return doc.VRMExt().Meta.Author
}

func Parse(r io.Reader, path string) (*VRMDocument, error) {
	dec := gltf.NewDecoder(r).WithReadHandler(&gltf.RelativeFileHandler{Dir: filepath.Dir(path)})
	doc := new(VRMDocument)
	if err := dec.Decode((*gltf.Document)(doc)); err != nil {
		return nil, err
	}
	return doc, nil
}
