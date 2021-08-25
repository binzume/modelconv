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
	options     *FBXToMQOOption
	boneNodeMap map[*fbx.Model]int
	boneNodes   []*fbx.Model
	bones       []*mqo.Bone
	upAxis      int
}

func NewFBXToMQOConverter(options *FBXToMQOOption) *fbxToMqo {
	if options == nil {
		options = &FBXToMQOOption{}
	}
	return &fbxToMqo{
		options:     options,
		boneNodeMap: map[*fbx.Model]int{},
		upAxis:      1,
	}
}

func (c *fbxToMqo) convertCoord(v *geom.Vector3) *geom.Vector3 {
	if c.upAxis == 2 {
		return &geom.Vector3{X: v.X, Y: v.Z, Z: v.Y}
	}
	return v
}

func (c *fbxToMqo) convertMaterial(src *fbx.Document, m *fbx.Material) *mqo.Material {
	mat := &mqo.Material{}
	mat.Name = m.Name()
	col := m.GetProperty70("DiffuseColor").ToVector3(1, 1, 1)
	opacity := m.GetProperty70("Opacity").ToFloat32(1)
	mat.Color = geom.Vector4{X: col.X, Y: col.Y, Z: col.Z, W: opacity}
	mat.Diffuse = m.GetProperty70("DiffuseFactor").ToFloat32(1)
	mat.Specular = m.GetProperty70("SpecularFactor").ToFloat32(0)
	textures := m.FindRefs("Texture")
	if len(textures) > 0 {
		// TODO: GetPropertyRef("DiffuseColor").FindChild("RelativeFilename").PropString(0)
		mat.Texture = textures[0].(*fbx.Obj).FindChild("RelativeFilename").PropString(0)
	}
	return mat
}

func (c *fbxToMqo) convertGeom(m *fbx.Geometry, doc *fbx.Document, objID int) *mqo.Object {
	obj := mqo.NewObject(m.Name())

	if c.upAxis == 2 {
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
		for _, sub := range node.FindRefs("Deformer") {
			models := sub.FindRefs("Model")
			if len(models) > 0 {
				model := models[0].(*fbx.Model)
				var modelPath []*fbx.Model
				m := model
				for m != nil {
					if _, exists := c.boneNodeMap[m]; exists {
						break
					}
					c.boneNodeMap[m] = 0
					modelPath = append(modelPath, m)
					m = m.Parent
				}
				for i := range modelPath {
					m := modelPath[len(modelPath)-i-1]

					pos := c.convertCoord(m.GetWorldMatrix().ApplyTo(&geom.Vector3{}))
					b := &mqo.Bone{
						ID:     len(c.bones) + 1,
						Name:   m.Name(),
						Group:  0,
						Pos:    mqo.Vector3Attr{Vector3: *pos},
						Parent: c.boneNodeMap[m.Parent],
					}
					c.boneNodeMap[m] = b.ID
					c.bones = append(c.bones, b)
				}
				bone := c.bones[c.boneNodeMap[model]-1]
				weights := sub.(*fbx.Obj).FindChild("Weights").Prop(0).ToFloat32Array()
				indexes := sub.(*fbx.Obj).FindChild("Indexes").Prop(0).ToInt32Array()
				if len(weights) == len(indexes) {
					for i := range indexes {
						bone.SetVertexWeight(objID, int(indexes[i])+1, weights[i])
					}
				} else {
					log.Println("ERR: Deformer weights ", len(weights), len(indexes))
				}
			}
			if len(models) != 1 {
				log.Println("ERR: Deformer models: ", sub.ID(), len(models))
			}
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

func (c *fbxToMqo) convertModel(dst *mqo.Document, m *fbx.Model, d int, parentMat *geom.Matrix4, doc *fbx.Document) {
	obj := mqo.NewObject(m.Name())
	g := m.FindRefs("Geometry")
	if len(g) > 0 {
		if mm, ok := g[0].(*fbx.Geometry); ok {
			obj = c.convertGeom(mm, doc, len(dst.Objects)+1)
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

	c.upAxis = src.GlobalSettings.GetProperty70("OriginalUpAxis").ToInt(1)

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

	mqo.GetBonePlugin(mqdoc).SetBones(c.bones)

	return mqdoc, nil
}
