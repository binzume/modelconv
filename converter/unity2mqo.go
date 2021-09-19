package converter

import (
	"log"

	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/mqo"
	"github.com/binzume/modelconv/unity"
)

type UnityToMQOOption struct {
}

type UnityToMQOConverter struct {
	options *UnityToMQOOption
}

type unityToMqoState struct {
	UnityToMQOOption
	src *unity.Scene
	dst *mqo.Document
}

func NewUnityToMQOConverter(options *UnityToMQOOption) *UnityToMQOConverter {
	if options == nil {
		options = &UnityToMQOOption{}
	}
	return &UnityToMQOConverter{
		options: options,
	}
}

func (conv *UnityToMQOConverter) Convert(secne *unity.Scene) (*mqo.Document, error) {
	state := unityToMqoState{
		UnityToMQOOption: *conv.options,
		src:              secne,
		dst:              mqo.NewDocument(),
	}

	// TODO
	state.dst.Materials = append(state.dst.Materials, &mqo.Material{Name: "dummy", Color: geom.Vector4{X: 1, Y: 1, Z: 1, W: 1}})

	for _, o := range secne.Objects {
		state.convertObject(o, 0)
	}

	return state.dst, nil
}

func (c *unityToMqoState) convertObject(o *unity.GameObject, d int) {
	dst := c.dst
	obj := mqo.NewObject(o.Name)
	dst.Objects = append(dst.Objects, obj)
	obj.UID = len(dst.Objects)
	obj.Depth = d

	tr := o.GetTransform()
	if tr == nil {
		return
	}

	var meshFilter *unity.MeshFilter
	var meshRenderer *unity.MeshRenderer
	if o.GetComponent(&meshFilter) && o.GetComponent(&meshRenderer) {
		if name, ok := unity.UnityMeshes[*meshFilter.Mesh]; ok {
			obj.Name += name
			if name == "Cube" {
				Cube(obj, tr.GetWorldMatrix(), 0)
			} else if name == "Plane" {
				Plane(obj, tr.GetWorldMatrix(), 0)
			} else if name == "Quad" {
				Quad(obj, tr.GetWorldMatrix(), 0)
			} else {
				log.Println("TODO:", name)
			}
		} else {
			asset := c.src.Assets.GetAsset(meshFilter.Mesh.GUID)
			if asset != nil {
				meta, _ := c.src.Assets.GetMetaFile(asset)
				log.Println("TODO:", asset.Path, meta.GetRecycleNameByFileID(meshFilter.Mesh.FileID))
			}
		}
	}

	for _, child := range tr.GetChildren() {
		c.convertObject(child.GetGameObject(), d+1)
	}
}

func AddGeometry(o *mqo.Object, tr *geom.Matrix4, mat int, vs []*geom.Vector3, faces [][]int) {
	voffset := len(o.Vertexes)
	for _, v := range vs {
		o.Vertexes = append(o.Vertexes, tr.ApplyTo(v))
	}
	for _, f := range faces {
		for i := range f {
			f[i] += voffset
		}
		o.Faces = append(o.Faces, &mqo.Face{Material: mat, Verts: f})
	}
}

func Cube(o *mqo.Object, tr *geom.Matrix4, mat int) {
	vs := []*geom.Vector3{
		{X: -0.5, Y: -0.5, Z: -0.5},
		{X: 0.5, Y: -0.5, Z: -0.5},
		{X: 0.5, Y: 0.5, Z: -0.5},
		{X: -0.5, Y: 0.5, Z: -0.5},
		{X: -0.5, Y: -0.5, Z: 0.5},
		{X: 0.5, Y: -0.5, Z: 0.5},
		{X: 0.5, Y: 0.5, Z: 0.5},
		{X: -0.5, Y: 0.5, Z: 0.5},
	}
	faces := [][]int{
		{0, 1, 2, 3}, {7, 6, 5, 4},
		{1, 0, 4, 5}, {3, 2, 6, 7},
		{2, 1, 5, 6}, {0, 3, 7, 4},
	}
	AddGeometry(o, tr, mat, vs, faces)
}

func Quad(o *mqo.Object, tr *geom.Matrix4, mat int) {
	vs := []*geom.Vector3{
		{X: -0.5, Y: -0.5},
		{X: 0.5, Y: -0.5},
		{X: 0.5, Y: 0.5},
		{X: -0.5, Y: 0.5},
	}
	faces := [][]int{
		{0, 1, 2, 3},
	}
	AddGeometry(o, tr, mat, vs, faces)
}

func Plane(o *mqo.Object, tr *geom.Matrix4, mat int) {
	vs := []*geom.Vector3{
		{X: -5, Y: 0, Z: -5},
		{X: 5, Y: 0, Z: -5},
		{X: 5, Y: 0, Z: 5},
		{X: -5, Y: 0, Z: 5},
	}
	faces := [][]int{
		{0, 1, 2, 3},
	}
	AddGeometry(o, tr, mat, vs, faces)
}
