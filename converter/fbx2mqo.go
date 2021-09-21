package converter

import (
	"log"
	"strings"

	"github.com/binzume/modelconv/fbx"
	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/mqo"
)

type FBXToMQOOption struct {
	DisableBlendShape bool
	DisableBone       bool
	ConvertWholeNode  bool
	ObjectDepth       int
	RootTransform     *geom.Matrix4
	TargetModelName   string
	MaterialOverride  []int
}

type FBXToMQOConverter struct {
	options *FBXToMQOOption
}

type fbxToMqoState struct {
	*FBXToMQOOption
	src           *fbx.Document
	dst           *mqo.Document
	materialIDMap map[*fbx.Material]int
	boneNodeMap   map[*fbx.Model]int
	boneNodes     []*fbx.Model
	bones         []*mqo.Bone
	coordMat      *geom.Matrix4
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
	return conv.ConvertTo(mqo.NewDocument(), src)
}

func (conv *FBXToMQOConverter) ConvertTo(dst *mqo.Document, src *fbx.Document) (*mqo.Document, error) {
	gs := src.GlobalSettings
	mat := geom.Matrix4{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 1,
	}
	mat[gs.GetProperty70("CoordAxis").ToInt(0)*4] = gs.GetProperty70("CoordAxisSign").ToFloat32(1)
	mat[gs.GetProperty70("UpAxis").ToInt(1)*4+1] = gs.GetProperty70("UpAxisSign").ToFloat32(1)
	mat[gs.GetProperty70("FrontAxis").ToInt(2)*4+2] = gs.GetProperty70("FrontAxisSign").ToFloat32(1)
	c := &fbxToMqoState{
		FBXToMQOOption: conv.options,
		src:            src,
		dst:            dst,
		boneNodeMap:    map[*fbx.Model]int{},
		materialIDMap:  map[*fbx.Material]int{},
		coordMat:       &mat,
	}
	c.convert()
	return c.dst, nil
}

func (c *fbxToMqoState) convert() {
	if c.MaterialOverride == nil {
		for _, mat := range c.src.Materials {
			c.materialIDMap[mat] = len(c.dst.Materials)
			c.dst.Materials = append(c.dst.Materials, c.convertMaterial(mat))
		}
	}

	transform := c.coordMat.Mul(c.src.Scene.GetMatrix())
	for _, m := range c.src.Scene.GetChildModels() {
		if c.TargetModelName != "" && c.TargetModelName+"::Model" != m.Name() {
			continue
		}
		if c.ConvertWholeNode || (m.Kind() != "LimbNode" && hasGeometryNode(m)) {
			c.convertModel(m, c.ObjectDepth, transform)
		}
	}

	if len(c.bones) > 0 {
		mqo.GetBonePlugin(c.dst).SetBones(c.bones)
	}
}

func (c *fbxToMqoState) convertMaterial(m *fbx.Material) *mqo.Material {
	mat := &mqo.Material{}
	mat.Name = m.Name()
	col := m.GetColor("DiffuseColor", &geom.Vector3{X: 1, Y: 1, Z: 1})
	opacity := m.GetFactor("Opacity", 1)
	mat.Color = geom.Vector4{X: col.X, Y: col.Y, Z: col.Z, W: opacity}
	mat.Diffuse = m.GetFactor("DiffuseFactor", 1)
	mat.Specular = m.GetColor("SpecularColor", &geom.Vector3{}).Scale(m.GetFactor("SpecularFactor", 0)).X
	mat.Emission = m.GetColor("EmissiveColor", &geom.Vector3{}).Scale(m.GetFactor("EmissiveFactor", 0)).X * 0.1 // ?
	mat.Ambient = m.GetColor("AmbientColor", &geom.Vector3{}).Scale(m.GetFactor("AmbientFactor", 0)).X
	mat.Power = m.GetFactor("ShininessExponent", 0)
	texture := m.GetTexture("DiffuseColor")
	if texture != nil {
		mat.Texture = texture.FindChild("RelativeFilename").GetString()
	}
	return mat
}

func (c *fbxToMqoState) convertModel(m *fbx.Model, d int, parentTransform *geom.Matrix4) {
	dst := c.dst
	obj := mqo.NewObject(strings.TrimPrefix(m.Name(), "Model::"))
	dst.Objects = append(dst.Objects, obj)
	obj.UID = len(dst.Objects)
	obj.Depth = d

	var materialIDs []int
	for _, mat := range m.FindRefs("Material") {
		materialIDs = append(materialIDs, c.materialIDMap[mat.(*fbx.Material)])
	}

	transform := parentTransform.Mul(m.GetMatrix())
	if c.ObjectDepth > 0 && d == c.ObjectDepth {
		// FIXME
		if c.RootTransform != nil {
			transform = c.RootTransform.Mul(c.coordMat)
		}
		if len(c.MaterialOverride) > 0 {
			materialIDs = c.MaterialOverride
		}
	}
	geometry := m.GetGeometry()
	if geometry != nil {
		shapes := c.convertGeometry(geometry, obj, transform, materialIDs)
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
		if c.ConvertWholeNode || m.Kind() != "LimbNode" {
			c.convertModel(m, d+1, transform)
		}
	}
}

func hasGeometryNode(m *fbx.Model) bool {
	if m.GetGeometry() != nil {
		return true
	}
	for _, child := range m.GetChildModels() {
		if hasGeometryNode(child) {
			return true
		}
	}
	return false
}

func (c *fbxToMqoState) convertGeometry(g *fbx.Geometry, obj *mqo.Object, transform *geom.Matrix4, materialIDs []int) []*mqo.Object {
	for _, v := range g.Vertices {
		obj.Vertexes = append(obj.Vertexes, transform.ApplyTo(v))
	}

	matnode := g.GetLayerElementMaterial()
	matArray := matnode.GetIndexes()
	matType := matnode.GetMappingInformationType()
	matByPolygon := matType == "ByPolygon" && len(matArray) >= len(g.Polygons)
	matAllSame := matType == "AllSame" && len(matArray) > 0

	uvnode := g.GetLayerElementUV()
	var uv []*geom.Vector2
	var uvIndex []int32
	if uvnode.GetMappingInformationType() == "ByPolygonVertex" {
		uv = uvnode.Array.GetVec2Array()
		if uvnode.GetReferenceInformationType() == "IndexToDirect" {
			uvIndex = uvnode.GetIndexes()
			if len(uvIndex) < g.PolygonVertexCount {
				uv = nil
				uvIndex = nil
			}
		} else {
			if len(uv) < g.PolygonVertexCount {
				uv = nil
			}
		}
	}

	vcount := 0
	for i, f := range g.Polygons {
		face := &mqo.Face{Verts: f}
		var matIdx int
		if matByPolygon {
			matIdx = int(matArray[i])
		} else if matAllSame {
			matIdx = int(matArray[0])
		}
		if matIdx < len(materialIDs) {
			face.Material = materialIDs[matIdx]
		}
		if uvIndex != nil { // Indexed
			for n := range f {
				v := uv[uvIndex[vcount+n]]
				face.UVs = append(face.UVs, geom.Vector2{X: v.X, Y: 1 - v.Y})
			}
		} else if uv != nil {
			for n := range f {
				v := uv[vcount+n]
				face.UVs = append(face.UVs, geom.Vector2{X: v.X, Y: 1 - v.Y})
			}
		}
		vcount += len(f)
		face.Flip()
		obj.Faces = append(obj.Faces, face)
	}

	var shapes []*mqo.Object

	for _, deformer := range g.GetDeformers() {
		if deformer.Kind() == "BlendShapeChannel" {
			// TODO: Apply deformer.FullWeights
			if !c.DisableBlendShape {
				for _, shape := range deformer.GetShapes() {
					shapes = append(shapes, c.convertShape(shape, obj, g, transform))
				}
			}
		} else {
			if !c.DisableBone {
				c.convertDeformer(deformer, obj.UID)
			}
		}
	}

	if !c.DisableBlendShape {
		for _, shape := range g.GetShapes() {
			shapes = append(shapes, c.convertShape(shape, obj, g, transform))
		}
	}
	return shapes
}

func (c *fbxToMqoState) convertDeformer(sub *fbx.Deformer, objID int) {
	model := sub.GetTarget()
	if model == nil {
		log.Println("ERR: Deformer has no model: ", sub.ID(), sub.Kind())
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

		pos := c.coordMat.Mul(m.GetWorldMatrix()).ApplyTo(&geom.Vector3{})
		b := &mqo.Bone{
			ID:     len(c.bones) + 1,
			Name:   strings.TrimPrefix(m.Name(), "Model::"),
			Pos:    mqo.Vector3Attr{Vector3: *pos},
			Parent: c.boneNodeMap[m.Parent],
		}
		c.boneNodeMap[m] = b.ID
		c.bones = append(c.bones, b)
	}
	bone := c.bones[c.boneNodeMap[model]-1]
	bone.Pos.Vector3 = *c.coordMat.Mul(sub.GetTransformLink()).ApplyTo(&geom.Vector3{})
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

func (c *fbxToMqoState) convertShape(node *fbx.GeometryShape, src *mqo.Object, g *fbx.Geometry, transfrom *geom.Matrix4) *mqo.Object {
	vertices := node.GetVertices()
	indexes := node.GetIndexes()

	obj := src.Clone()
	obj.Name = node.Name()

	if len(vertices) != len(indexes) {
		log.Println("ERROR: Shape ", node.Name(), len(vertices), len(indexes))
		return obj
	}

	for i, idx := range indexes {
		if int(idx) < len(obj.Vertexes) {
			obj.Vertexes[idx] = transfrom.ApplyTo(g.Vertices[idx].Add(vertices[i]))
		}
	}

	return obj
}
