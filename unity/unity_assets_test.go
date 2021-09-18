package unity

import (
	"os"
	"testing"
)

const PackagePath = "../testdata/unity/test.unitypackage"
const ExtractedPackagePath = "../testdata/unity/extracted"
const SceneAssetPath = "Assets/world.unity"

func TestExtractPackage(t *testing.T) {
	_ = os.RemoveAll(ExtractedPackagePath)
	err := extractPackage(PackagePath, ExtractedPackagePath)
	if err != nil {
		t.Error("Cannot extract package.", err)
	}
}

func TestOpenPackage(t *testing.T) {
	assets, err := OpenPackage(ExtractedPackagePath)
	if err != nil {
		t.Error("Cannot open package.", err)
	}
	defer assets.Close()

	for _, a := range assets.GetAllAssets() {
		t.Log(a.GUID, a.Path)
	}
}

func TestLoadScene(t *testing.T) {
	assets, err := OpenPackage(ExtractedPackagePath)
	if err != nil {
		t.Error("Cannot open package.", err)
	}
	defer assets.Close()

	scene, err := LoadScene(assets, assets.GetAssetByPath(SceneAssetPath))
	if err != nil {
		t.Error("Cannot open scene.", err)
	}
	DumpScene(scene)
}