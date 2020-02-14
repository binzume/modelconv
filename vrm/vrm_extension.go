package vrm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qmuntal/gltf"
)

const ExtensionName = "VRM"

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

func Unmarshal(data []byte) (interface{}, error) {
	var vrmext VRM
	if err := json.Unmarshal([]byte(data), &vrmext); err != nil {
		return nil, err
	}
	return &vrmext, nil
}

type Document gltf.Document

func (doc *Document) VRM() *VRM {
	if ext, ok := doc.Extensions[ExtensionName].(*VRM); ok {
		return ext
	}
	ext := NewVRM()
	if doc.Extensions == nil {
		doc.Extensions = gltf.Extensions{}
	}
	doc.Extensions[ExtensionName] = ext
	if !doc.IsExtentionUsed(ExtensionName) {
		doc.ExtensionsUsed = append(doc.ExtensionsUsed, ExtensionName)
	}
	return ext
}

func (doc *Document) IsExtentionUsed(extname string) bool {
	for _, ex := range doc.ExtensionsUsed {
		if ex == extname {
			return true
		}
	}
	return false
}

func (doc *Document) ValidateBones() error {
	errorBones := doc.VRM().CheckRequiredBones()
	if len(errorBones) > 0 {
		return fmt.Errorf("Bone error. Missing bones: %v", strings.Join(errorBones, ","))
	}
	return nil
}
