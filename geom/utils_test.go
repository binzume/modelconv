package geom

import (
	"testing"
)

func TestTriangulate(t *testing.T) {
	tris := Triangulate([]*Vector3{
		{0, 0, 0},
		{0, 1, 0},
		{0, 1, 1},
	})
	t.Log(tris)

	tris2 := Triangulate([]*Vector3{
		{0, 0, 0},
		{0, 1, 0},
		{0, 1, 1},
		{0, 0, 1},
	})
	t.Log(tris2)

	// non-convex
	tris3 := Triangulate([]*Vector3{
		{0, 0, 0},
		{0, 1, 0},
		{0, 1, 1},
		{0, 0.8, 0.2},
	})
	t.Log(tris3)

	// Empty
	if len(Triangulate(nil)) != 0 {
		t.Error("not empty")
	}
}
