package converter

import (
	"log"
	"strings"

	"github.com/binzume/modelconv/fbx"
	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/mqo"
)

type FBXToMQOOption struct {
	disableBlendShape bool
	disableBone       bool
}

type FBXToMQOConverter struct {
	options *FBXToMQOOption
}

type fbxToMqoState struct {
	*FBXToMQOOption
	src         *fbx.Document
	boneNodeMap map[*fbx.Model]int
	boneNodes   []*fbx.Model
	bones       []*mqo.Bone
	upAxis      int
}

func NewFBXToMQOConverter(options *FBXToMQOOption) *FBXToMQOConverter {
	if options == nil {
		options = &FBXToMQOOption{}
	}
	return &FBXToMQOConverter{
		options: options,
	}
}

func (conv *FBXToMQOConverter) Convert(src *fbx.Document) (*mqo.Document, error) {

	c := &fbxToMqoState{
		FBXToMQOOption: conv.options,
		src:            src,
		boneNodeMap:    map[*fbx.Model]int{},
		upAxis:         src.GlobalSettings.GetProperty70("OriginalUpAxis").ToInt(1),
	}

	mqdoc := mqo.NewDocument()

	for _, mat := range src.Materials {
		mqdoc.Materials = append(mqdoc.Materials, c.convertMaterial(mat))
	}

	transform := geom.NewMatrix4()
	for _, o := range src.Scene.FindRefs("Model") {
		if m, ok := o.(*fbx.Model); ok {
			c.convertModel(mqdoc, m, 0, transform)
		}
	}

	mqo.GetBonePlugin(mqdoc).SetBones(c.bones)

	return mqdoc, nil
}

func (c *fbxToMqoState) convertCoord(v *geom.Vector3) *geom.Vector3 {
	// TODO
	if c.upAxis == 2 {
		return &geom.Vector3{X: v.X, Y: v.Z, Z: v.Y}
	}
	return v
}

func (c *fbxToMqoState) convertMaterial(m *fbx.Material) *mqo.Material {
	mat := &mqo.Material{}
	mat.Name = m.Name()
	col := m.GetProperty70("DiffuseColor").ToVector3(1, 1, 1)
	opacity := m.GetProperty70("Opacity").ToFloat32(1)
	mat.Color = geom.Vector4{X: col.X, Y: col.Y, Z: col.Z, W: opacity}
	mat.Diffuse = m.GetProperty70("DiffuseFactor").ToFloat32(1)
	mat.Specular = m.GetProperty70("SpecularFactor").ToFloat32(0) * m.GetProperty70("SpecularColor").ToFloat32(0)
	// mat.Emission = m.GetProperty70("EmissiveFactor").ToFloat32(0) * m.GetProperty70("EmissiveColor").ToFloat32(0)
	mat.Ambient = m.GetProperty70("AmbientFactor").ToFloat32(0) * m.GetProperty70("AmbientColor").ToFloat32(0)
	textures := m.FindRefs("Texture")
	if len(textures) > 0 {
		// TODO: GetPropertyRef("DiffuseColor").FindChild("RelativeFilename").PropString(0)
		mat.Texture = textures[0].(*fbx.Obj).FindChild("RelativeFilename").GetString("")
	}
	return mat
}

func (c *fbxToMqoState) convertModel(dst *mqo.Document, m *fbx.Model, d int, parentTransform *geom.Matrix4) {
	obj := mqo.NewObject(strings.TrimPrefix(m.Name(), "Model::"))
	dst.Objects = append(dst.Objects, obj)
	obj.UID = len(dst.Objects)
	obj.Depth = d

	transform := parentTransform.Mul(m.GetMatrix())
	geometry := m.FindRefs("Geometry")
	if len(geometry) > 0 {
		if g, ok := geometry[0].(*fbx.Geometry); ok {
			shapes := c.convertGeometry(g, obj, transform)
			if len(shapes) > 0 {
				morphPlugin := mqo.GetMorphPlugin(dst)
				var morphTargets mqo.MorphTargetList
				morphPlugin.MorphSet.Targets = append(morphPlugin.MorphSet.Targets, &morphTargets)
				morphTargets.Base = obj.Name
				for _, o := range shapes {
					dst.Objects = append(dst.Objects, o)
					o.UID = len(dst.Objects)
					o.Depth = d + 1
					o.Visible = false
					morphTargets.Target = append(morphTargets.Target, &mqo.MorphTarget{Name: o.Name})
				}
			}
		}
	}
	for _, o := range m.FindRefs("Model") {
		if m, ok := o.(*fbx.Model); ok {
			c.convertModel(dst, m, d+1, transform)
		}
	}
}

func (c *fbxToMqoState) convertGeometry(g *fbx.Geometry, obj *mqo.Object, transform *geom.Matrix4) []*mqo.Object {
	for _, v := range g.Vertices {
		obj.Vertexes = append(obj.Vertexes, c.convertCoord(transform.ApplyTo(v)))
	}

	var matByPolygon []int32

	matnode := g.FindChild("LayerElementMaterial")
	if matnode.FindChild("MappingInformationType").GetString("") == "ByPolygon" {
		matByPolygon = matnode.FindChild("Materials").GetInt32Array()
	}

	var uv []*geom.Vector2
	var uvIndex []int32
	uvnode := g.FindChild("LayerElementUV")
	if uvnode.FindChild("MappingInformationType").GetString("") == "ByPolygonVertex" {
		uv = uvnode.FindChild("UV").Attr(0).ToVec2Array()
		if uv != nil {
			uvIndex = uvnode.FindChild("UVIndex").GetInt32Array()
		}
	}

	vcount := 0
	for i, f := range g.Faces {
		face := &mqo.Face{Verts: f}
		if len(matByPolygon) == len(g.Faces) {
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

	if !c.disableBone {
		for _, node := range g.FindRefs("Deformer") {
			c.convertDeformer(node, obj.UID)
		}
	}

	var shapes []*mqo.Object
	if !c.disableBlendShape {
		for _, node := range g.GetChildren() {
			if node.Name == "Shape" {
				shapes = append(shapes, c.convertShape(node, obj, g, transform))
			}
		}
	}
	return shapes
}

func (c *fbxToMqoState) convertDeformer(node fbx.Object, objID int) {
	for _, sub := range node.FindRefs("Deformer") {
		models := sub.FindRefs("Model")
		if len(models) != 1 {
			log.Println("ERR: Deformer models: ", sub.ID(), len(models))
		}
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
					Name:   strings.TrimPrefix(m.Name(), "Model::"),
					Group:  0,
					Pos:    mqo.Vector3Attr{Vector3: *pos},
					Parent: c.boneNodeMap[m.Parent],
				}
				c.boneNodeMap[m] = b.ID
				c.bones = append(c.bones, b)
			}
			bone := c.bones[c.boneNodeMap[model]-1]
			weights := sub.(*fbx.Obj).FindChild("Weights").GetFloat32Array()
			indexes := sub.(*fbx.Obj).FindChild("Indexes").GetInt32Array()
			if len(weights) == len(indexes) {
				for i := range indexes {
					bone.SetVertexWeight(objID, int(indexes[i])+1, weights[i])
				}
			} else {
				log.Println("ERR: Deformer weights ", len(weights), len(indexes))
			}
		}
	}
}

func (c *fbxToMqoState) convertShape(node *fbx.Node, src *mqo.Object, g *fbx.Geometry, transfrom *geom.Matrix4) *mqo.Object {
	vertices := node.FindChild("Vertices").Attr(0).ToVec3Array()
	indexes := node.FindChild("Indexes").Attr(0).ToInt32Array()
	_ = node.FindChild("Normals").Attr(0).ToVec3Array()

	obj := src.Clone()
	obj.Name = node.GetString("")

	if len(vertices) != len(indexes) {
		log.Println("ERROR: Shape ", node.GetString(""), len(vertices), len(indexes))
		return obj
	}

	for i, idx := range indexes {
		if int(idx) < len(obj.Vertexes) {
			obj.Vertexes[idx] = c.convertCoord(transfrom.ApplyTo(g.Vertices[idx].Add(vertices[i])))
		}
	}

	return obj
}
