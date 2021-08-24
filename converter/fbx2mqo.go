package converter

import (
	"log"
	"math"

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
	mat.Color = geom.Vector4{
		X: m.GetProperty70("DiffuseColor").Get(0).ToFloat32(1),
		Y: m.GetProperty70("DiffuseColor").Get(1).ToFloat32(1),
		Z: m.GetProperty70("DiffuseColor").Get(2).ToFloat32(1),
		W: m.GetProperty70("Opacity").Get(0).ToFloat32(1)}
	mat.Diffuse = 1
	mat.Specular = m.GetProperty70("SpecularFactor").Get(0).ToFloat32(0)
	textures := m.FindRefs("Texture")
	if len(textures) > 0 {
		// TODO: GetPropertyRef("DiffuseColor").FindChild("RelativeFilename").PropString(0)
		mat.Texture = textures[0].(*fbx.Obj).FindChild("RelativeFilename").PropString(0)
	}
	return mat
}

func (c *fbxToMqo) convertGeom(m *fbx.Geometry, doc *fbx.Document) *mqo.Object {
	obj := mqo.NewObject(m.Name())

	upAxis := doc.GlobalSettings.GetProperty70("OriginalUpAxis").Get(0).ToInt(1)

	if upAxis == 2 {
		for _, v := range m.Vertices {
			obj.Vertexes = append(obj.Vertexes, &geom.Vector3{X: v.X, Y: v.Z, Z: v.Y})
		}
	} else {
		obj.Vertexes = m.Vertices
	}

	var matByPolygon []int32

	matnode := m.FindChild("LayerElementMaterial")
	if matnode.FindChild("MappingInformationType").PropString(0) == "ByPolygon" {
		matByPolygon = matnode.FindChild("Materials").Prop(0).ToInt32Array()
	}

	var uv []*geom.Vector2
	var uvIndex []int32
	uvnode := m.FindChild("LayerElementUV")
	if uvnode.FindChild("MappingInformationType").PropString(0) == "ByPolygonVertex" {
		uv = uvnode.FindChild("UV").Prop(0).ToVec2Array()
		if uv != nil {
			uvIndex = uvnode.FindChild("UVIndex").Prop(0).ToInt32Array()
		}
	}
	for _, node := range m.FindRefs("Deformer") {
		sub := node.FindRefs("Deformer")
		name := ""
		if len(sub) > 0 && len(sub[0].FindRefs("Model")) > 0 {
			name = sub[0].FindRefs("Model")[0].Name()
		}
		log.Println("TODO: skinning", name, node.ID(), len(sub))
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

func (c *fbxToMqo) convertModel(dst *mqo.Document, m *fbx.Model, d int, parentMat *geom.Matrix4, doc *fbx.Document) {
	obj := mqo.NewObject(m.Name())
	g := m.FindRefs("Geometry")
	if len(g) > 0 {
		if mm, ok := g[0].(*fbx.Geometry); ok {
			obj = c.convertGeom(mm, doc)
			obj.Name = m.Name()
		}
	}
	rotv := m.Rotation.Scale(math.Pi / 180)
	tr := geom.NewTranslateMatrix4(m.Translation.X, m.Translation.Y, m.Translation.Z)
	rot := geom.NewEulerRotationMatrix4(rotv.X, rotv.Y, rotv.Z, 0) // XYZ order?
	sacle := geom.NewScaleMatrix4(m.Scaling.X, m.Scaling.Y, m.Scaling.Z)
	mat := parentMat.Mul(tr).Mul(rot).Mul(sacle)
	obj.Transform(func(v *geom.Vector3) {
		*v = *mat.ApplyTo(v)
	})
	obj.Depth = d
	dst.Objects = append(dst.Objects, obj)
	for _, o := range m.FindRefs("Model") {
		if m, ok := o.(*fbx.Model); ok {
			c.convertModel(dst, m, d+1, mat, doc)
		}
	}
}

func (c *fbxToMqo) Convert(src *fbx.Document) (*mqo.Document, error) {
	//  src.Dump(os.Stdout, false)

	mqdoc := mqo.NewDocument()

	for _, mat := range src.Materials {
		mqdoc.Materials = append(mqdoc.Materials, c.convertMaterial(src, mat))
	}

	mm := geom.NewMatrix4()
	for _, o := range src.Scene.FindRefs("Model") {
		if m, ok := o.(*fbx.Model); ok {
			c.convertModel(mqdoc, m, 0, mm, src)
		}
	}

	return mqdoc, nil
}
