package fbx

import (
	"os"
	"testing"

	"github.com/binzume/modelconv/geom"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument()
	g := NewGeometry("Hello", []*geom.Vector3{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 0}, {X: 1, Y: 1, Z: 0}}, [][]int{{0, 1, 2}})
	doc.AddObject(g)

	mat := NewMaterial("mat1")
	mat.SetColorProperty("DiffuseColor", 1, 1, 1)
	doc.AddObject(mat)

	model := NewModel("model01")
	model.SetTranslation(&geom.Vector3{X: 1, Y: 0, Z: 0})
	model.SetScaling(&geom.Vector3{X: 1, Y: 2, Z: 1})
	doc.AddObject(model)

	doc.AddConnection(doc.Scene, model)
	doc.AddConnection(model, g)
	doc.AddConnection(model, mat)

	//w, _ := os.Create("test.fbx")
	w := os.Stdout
	Write(w, doc)
}
