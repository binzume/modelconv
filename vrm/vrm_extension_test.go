package vrm

import (
	"testing"

	"github.com/qmuntal/gltf"
)

func TestLoad(t *testing.T) {
	doc, _ := gltf.Open("../testdata/AliciaSolid.vrm")
	if v, ok := doc.Extensions[ExtensionName].(*VRM); ok {
		t.Log(v.Meta.Title)
		t.Log(v.Meta.Author)
		t.Log(v.ExporterVersion)
	} else {
		t.Error("VRM extension not found.")
	}
}
