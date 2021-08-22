package converter

import (
	"math"
	"os"

	"github.com/binzume/modelconv/fbx"
	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/mqo"
)

type FBXToMQOOption struct {
}

type fbxToMqo struct {
	options *FBXToMQOOption
}

func NewFBXToMQOConverter(options *FBXToMQOOption) *fbxToMqo {
	if options == nil {
		options = &FBXToMQOOption{}
	}
	return &fbxToMqo{
		options: options,
	}
}

func (c *fbxToMqo) convertMaterial(src *fbx.Document, m *fbx.Material) *mqo.Material {
	mat := &mqo.Material{}
	mat.Name = m.Name()
	mat.Color = geom.Vector4{X: 1, Y: 1, Z: 1, W: 1}
	mat.Diffuse = 1
	mat.Texture = "textures.png"
	return mat
}

func (c *fbxToMqo) convertGeom(m *fbx.Mesh) *mqo.Object {
	obj := mqo.NewObject(m.Name())
	obj.Vertexes = m.Vertices

	var matByPolygon []int32

	matnode := m.Child("LayerElementMaterial")
	if matnode.Child("MappingInformationType").PropString(0) == "ByPolygon" {
		matByPolygon = matnode.Child("Materials").Prop(0).ToInt32Array()
	}

	var uv []*geom.Vector2
	var uvIndex []int32
	uvnode := m.Child("LayerElementUV")
	if uvnode.Child("MappingInformationType").PropString(0) == "ByPolygonVertex" {
		uv = uvnode.Child("UV").Prop(0).ToVec2Array()
		if uv != nil {
			uvIndex = uvnode.Child("UVIndex").Prop(0).ToInt32Array()
		}
	}

	vcount := 0
	for i, f := range m.Faces {
		face := &mqo.Face{Verts: f}
		if len(matByPolygon) == len(m.Faces) {
			face.Material = int(matByPolygon[i])
		}
		if len(uvIndex) > vcount+len(f) {
			for n := range f {
				v := uv[uvIndex[vcount+n]]
				face.UVs = append(face.UVs, geom.Vector2{X: v.X, Y: 1 - v.Y})
			}
		}
		vcount += len(f)
		obj.Faces = append(obj.Faces, face)
	}

	return obj
}

func (c *fbxToMqo) convertModel(dst *mqo.Document, m *fbx.Model, d int, parentMat *geom.Matrix4) {
	obj := mqo.NewObject(m.Name())
	g := m.FindRefs("Geometry")
	if len(g) > 0 {
		if mm, ok := g[0].(*fbx.Mesh); ok {
			obj = c.convertGeom(mm)
		}
	}
	// log.Println(m.Rotation, m.Scaling, m.Translation)
	rotv := m.Rotation.Scale(math.Pi / 180)
	tr := geom.NewTranslateMatrix4(m.Translation.X, m.Translation.Y, m.Translation.Z)
	sacle := geom.NewScaleMatrix4(m.Scaling.X, m.Scaling.Y, m.Scaling.Z)
	rot := geom.NewEulerRotationMatrix4(rotv.X, rotv.Y, rotv.Z, 1)
	mat := tr.Mul(rot).Mul(sacle)
	obj.Transform(func(v *geom.Vector3) {
		*v = *mat.ApplyTo(v)
	})
	obj.Depth = d
	dst.Objects = append(dst.Objects, obj)
	for _, o := range m.FindRefs("Model") {
		if m, ok := o.(*fbx.Model); ok {
			c.convertModel(dst, m, d+1, mat)
		}
	}
}

func (c *fbxToMqo) Convert(src *fbx.Document) (*mqo.Document, error) {
	src.Dump(os.Stdout, false)

	mqdoc := mqo.NewDocument()

	for _, mat := range src.Materials {
		mqdoc.Materials = append(mqdoc.Materials, c.convertMaterial(src, mat))
	}

	mm := geom.NewMatrix4()
	for _, o := range src.Scene.FindRefs("Model") {
		if m, ok := o.(*fbx.Model); ok {
			c.convertModel(mqdoc, m, 0, mm)
		}
	}

	return mqdoc, nil
}
