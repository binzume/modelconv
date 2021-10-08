package fbx

import (
	"os"
	"testing"

	"github.com/binzume/modelconv/geom"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument()

	model := NewModel("model01", "Mesh")
	model.SetTranslation(&geom.Vector3{X: 1, Y: 0, Z: 0})
	model.SetScaling(&geom.Vector3{X: 1, Y: 2, Z: 1})
	doc.AddObject(model)

	g := NewGeometry("model01mesh", []*geom.Vector3{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 0}, {X: 1, Y: 1, Z: 0}}, [][]int{{0, 1, 2}})
	g.SetLayerElementMaterialIndex([]int32{0}, AllSame)
	g.SetLayerElementNormal([]*geom.Vector3{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 0}, {X: 1, Y: 1, Z: 0}}, ByVertice)
	doc.AddObject(g)

	mat := NewMaterial("mat01")
	mat.SetColor("DiffuseColor", &geom.Vector3{X: 1, Y: 0, Z: 0.5})
	doc.AddObject(mat)

	doc.AddConnection(doc.Scene, model) // add model to scene
	doc.AddConnection(model, g)         // set geometry to model
	doc.AddConnection(model, mat)       // set material to model

	w, _ := os.Create("../testdata/test01.fbx")
	//w := os.Stdout
	Write(w, doc)
}
