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
	convertWholeNode  bool
}

type FBXToMQOConverter struct {
	options *FBXToMQOOption
}

type fbxToMqoState struct {
	*FBXToMQOOption
	src         *fbx.Document
	dst         *mqo.Document
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
	dst := mqo.NewDocument()

	c := &fbxToMqoState{
		FBXToMQOOption: conv.options,
		src:            src,
		dst:            dst,
		boneNodeMap:    map[*fbx.Model]int{},
		upAxis:         src.GlobalSettings.GetProperty70("UpAxis").ToInt(1),
	}

	for _, mat := range src.Materials {
		dst.Materials = append(dst.Materials, c.convertMaterial(mat))
	}

	transform := geom.NewMatrix4()
	for _, m := range src.Scene.GetChildModels() {
		if c.convertWholeNode || c.containsGeometry(m) {
			c.convertModel(m, 0, transform)
		}
	}

	mqo.GetBonePlugin(dst).SetBones(c.bones)

	return dst, nil
}

func (c *fbxToMqoState) convertCoord(v *geom.Vector3) *geom.Vector3 {
	// TODO
	if c.upAxis == 2 {
		return &geom.Vector3{X: v.X, Y: v.Z, Z: v.Y}
	}
	return &geom.Vector3{X: v.X, Y: v.Y, Z: v.Z}
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
		mat.Texture = textures[0].(*fbx.Obj).FindChild("RelativeFilename").GetString()
	}
	return mat
}

func (c *fbxToMqoState) convertModel(m *fbx.Model, d int, parentTransform *geom.Matrix4) {
	dst := c.dst
	obj := mqo.NewObject(strings.TrimPrefix(m.Name(), "Model::"))
	dst.Objects = append(dst.Objects, obj)
	obj.UID = len(dst.Objects)
	obj.Depth = d

	transform := parentTransform.Mul(m.GetMatrix())
	geometry := m.GetGeometry()
	if geometry != nil {
		shapes := c.convertGeometry(geometry, obj, transform)
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
	for _, m := range m.GetChildModels() {
		c.convertModel(m, d+1, transform)
	}
}

func (c *fbxToMqoState) containsGeometry(m *fbx.Model) bool {
	if m.GetGeometry() != nil {
		return true
	}
	for _, child := range m.GetChildModels() {
		if c.containsGeometry(child) {
			return true
		}
	}
	return false
}

func (c *fbxToMqoState) convertGeometry(g *fbx.Geometry, obj *mqo.Object, transform *geom.Matrix4) []*mqo.Object {
	for _, v := range g.Vertices {
		obj.Vertexes = append(obj.Vertexes, c.convertCoord(transform.ApplyTo(v)))
	}

	var matByPolygon []int32

	matnode := g.FindChild("LayerElementMaterial")
	if matnode.FindChild("MappingInformationType").GetString() == "ByPolygon" {
		matByPolygon = matnode.FindChild("Materials").GetInt32Array()
	}

	var uv []*geom.Vector2
	var uvIndex []int32
	uvnode := g.FindChild("LayerElementUV")
	if uvnode.FindChild("MappingInformationType").GetString() == "ByPolygonVertex" {
		uv = uvnode.FindChild("UV").GetVec2Array()
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
		face.Flip()
		obj.Faces = append(obj.Faces, face)
	}

	if !c.disableBone {
		for _, deformer := range g.GetDeformers() {
			c.convertDeformer(deformer, obj.UID)
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

func (c *fbxToMqoState) convertDeformer(sub *fbx.Deformer, objID int) {
	model := sub.GetTarget()
	if model == nil {
		log.Println("ERR: Deformer model: ", sub.ID())
		return
	}
	var modelPath []*fbx.Model
	m := model
	for m != nil && m != c.src.Scene {
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
	weights := sub.GetWeights()
	indexes := sub.GetIndexes()
	if len(weights) == len(indexes) {
		for i := range indexes {
			bone.SetVertexWeight(objID, int(indexes[i])+1, weights[i])
		}
	} else {
		log.Println("ERR: Deformer weights ", len(weights), len(indexes))
	}
}

func (c *fbxToMqoState) convertShape(node *fbx.Node, src *mqo.Object, g *fbx.Geometry, transfrom *geom.Matrix4) *mqo.Object {
	vertices := node.FindChild("Vertices").GetVec3Array()
	indexes := node.FindChild("Indexes").GetInt32Array()
	_ = node.FindChild("Normals").GetVec3Array()

	obj := src.Clone()
	obj.Name = node.GetString()

	if len(vertices) != len(indexes) {
		log.Println("ERROR: Shape ", node.GetString(), len(vertices), len(indexes))
		return obj
	}

	for i, idx := range indexes {
		if int(idx) < len(obj.Vertexes) {
			obj.Vertexes[idx] = c.convertCoord(transfrom.ApplyTo(g.Vertices[idx].Add(vertices[i])))
		}
	}

	return obj
}
