package vrm

import (
	"os"
	"testing"

	"github.com/qmuntal/gltf"
)

func TestLoad(t *testing.T) {
	vrmPath := "../testdata/AliciaSolid.vrm"
	if _, err := os.Stat(vrmPath); err != nil {
		t.Skip()
	}
	doc, _ := gltf.Open(vrmPath)
	if v, ok := doc.Extensions[ExtensionName].(*VRM); ok {
		t.Log(v.Meta.Title)
		t.Log(v.Meta.Author)
		t.Log(v.ExporterVersion)
	} else {
		t.Error("VRM extension not found.")
	}
}
