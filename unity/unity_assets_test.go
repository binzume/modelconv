package unity

import (
	"os"
	"testing"
)

const PackagePath = "../testdata/unity/unity2020.unitypackage"
const ProjectPath = "../testdata/unity/TestProject"
const ExtractedPackagePath = "../testdata/unity/extracted.generated"
const SceneAssetPath = "Assets/Scenes/SampleScene.unity"

func TestExtractPackage(t *testing.T) {
	if _, err := os.Stat(PackagePath); err != nil {
		t.Skip()
	}

	_ = os.RemoveAll(ExtractedPackagePath)
	err := extractPackage(PackagePath, ExtractedPackagePath)
	if err != nil {
		t.Error("Cannot extract package.", err)
	}
}

func TestOpenPackage(t *testing.T) {
	if _, err := os.Stat(ExtractedPackagePath); err != nil {
		t.Skip()
	}

	assets, err := OpenPackage(ExtractedPackagePath)
	if err != nil {
		t.Fatal("Cannot open package.", err)
	}
	defer assets.Close()

	for _, a := range assets.GetAllAssets() {
		t.Log(a.GUID, a.Path)
	}
}

func TestOpenProject(t *testing.T) {
	if _, err := os.Stat(ExtractedPackagePath); err != nil {
		t.Skip()
	}

	assets, err := OpenProject(ProjectPath)
	if err != nil {
		t.Fatal("Cannot open package.", err)
	}
	defer assets.Close()

	for _, a := range assets.GetAllAssets() {
		t.Log(a.GUID, a.Path)
	}
}

func TestLoadScene(t *testing.T) {
	if _, err := os.Stat(ExtractedPackagePath); err != nil {
		t.Skip()
	}

	assets, err := OpenPackage(ExtractedPackagePath)
	if err != nil {
		t.Fatal("Cannot open package.", err)
	}
	defer assets.Close()

	scene, err := LoadScene(assets, SceneAssetPath)
	if err != nil {
		t.Fatal("Cannot open scene.", err)
	}

	DumpScene(scene, true)

	var transform *Transform
	scene.Objects[0].GetComponent(&transform)
	if transform == nil {
		t.Error("GetComponent return nil", err)
	}
	t.Log(transform.GetGameObject() == scene.Objects[0])
}
