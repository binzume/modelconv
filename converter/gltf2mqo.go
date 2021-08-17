package converter

import (
	"log"

	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/mqo"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

type GLTFToMQOOption struct {
}

type gltfToMqo struct {
	options *GLTFToMQOOption
}

func NewGLTFToMQOConverter(options *GLTFToMQOOption) *gltfToMqo {
	if options == nil {
		options = &GLTFToMQOOption{}
	}
	return &gltfToMqo{
		options: options,
	}
}

func (c *gltfToMqo) convertMaterial(src *gltf.Document, m *gltf.Material) *mqo.Material {
	mat := &mqo.Material{}
	mat.Name = m.Name
	mat.DoubleSided = m.DoubleSided
	mat.Shader = 2
	mat.EmissionColor = &mqo.Vector3{X: m.EmissiveFactor[0], Y: m.EmissiveFactor[1], Z: m.EmissiveFactor[2]}
	mat.Ex2 = &mqo.MaterialEx2{
		ShaderType: "hlsl",
		ShaderName: "glTF",
		ShaderParams: map[string]interface{}{
			"AlphaMode":   m.AlphaMode,
			"AlphaCutoff": m.AlphaCutoffOrDefault(),
		},
	}
	if m.PBRMetallicRoughness != nil {
		col := m.PBRMetallicRoughness.BaseColorFactorOrDefault()
		mat.Color = mqo.Vector4{X: col[0], Y: col[1], Z: col[2], W: col[3]}
		mat.Specular = m.PBRMetallicRoughness.MetallicFactorOrDefault()
		mat.Ex2.ShaderParams["Metallic"] = m.PBRMetallicRoughness.MetallicFactorOrDefault()
		mat.Ex2.ShaderParams["Roughness"] = m.PBRMetallicRoughness.RoughnessFactorOrDefault()
		if m.PBRMetallicRoughness.BaseColorTexture != nil {
			i := m.PBRMetallicRoughness.BaseColorTexture.Index
			mat.Texture = src.Images[*src.Textures[i].Source].URI
		}
	}
	return mat
}

func (c *gltfToMqo) convertMesh(src *gltf.Document, m *gltf.Mesh) *mqo.Object {
	obj := mqo.NewObject(m.Name)
	for _, p := range m.Primitives {
		if p.Indices == nil {
			continue
		}
		if a, ok := p.Attributes["POSITION"]; ok {
			acr := src.Accessors[a]
			pos, err := modeler.ReadPosition(src, acr, [][3]float32{})
			if err != nil {
				log.Fatalf("err %v", err)
				continue
			}
			for _, v := range pos {
				obj.Vertexes = append(obj.Vertexes, &mqo.Vector3{X: v[0], Y: v[1], Z: v[2]})
			}
		}
		var texCoord [][2]float32
		if a, ok := p.Attributes["TEXCOORD_0"]; ok {
			acr := src.Accessors[a]
			t, err := modeler.ReadTextureCoord(src, acr, [][2]float32{})
			texCoord = t
			if err != nil {
				log.Fatalf("err %v", err)
				continue
			}
		}
		acr := src.Accessors[*p.Indices]
		indices, err := modeler.ReadIndices(src, acr, []uint32{})
		if err != nil {
			log.Fatalf("err %v", err)
			continue
		}
		mat := 0
		if p.Material != nil {
			mat = int(*p.Material)
		}
		for i := 0; i < len(indices)/3; i++ {
			f := &mqo.Face{Material: mat, Verts: []int{int(indices[i*3]), int(indices[i*3+2]), int(indices[i*3+1])}}
			if len(texCoord) > int(indices[i*3]) {
				f.UVs = []mqo.Vector2{
					{X: texCoord[indices[i*3]][0], Y: texCoord[indices[i*3]][1]},
					{X: texCoord[indices[i*3+2]][0], Y: texCoord[indices[i*3+2]][1]},
					{X: texCoord[indices[i*3+1]][0], Y: texCoord[indices[i*3+1]][1]}}
			}
			obj.Faces = append(obj.Faces, f)
		}
	}
	return obj
}

func (c *gltfToMqo) convertBones(src *gltf.Document, node uint32, parent int, bones []*mqo.Bone) []*mqo.Bone {
	n := src.Nodes[node]
	pos := geom.NewVector3FromArray(n.Translation)
	if parent > 0 {
		pos = pos.Add(&bones[parent-1].Pos.Vector3)
	}
	b := &mqo.Bone{
		ID:     len(bones) + 1,
		Name:   n.Name,
		Group:  0,
		Pos:    mqo.Vector3Attr{Vector3: *pos},
		Parent: parent,
	}
	bones = append(bones, b)
	for _, child := range n.Children {
		bones = c.convertBones(src, child, b.ID, bones)
	}
	return bones
}

func (c *gltfToMqo) Convert(src *gltf.Document) (*mqo.Document, error) {
	mqdoc := mqo.NewDocument()

	for _, mat := range src.Materials {
		mqdoc.Materials = append(mqdoc.Materials, c.convertMaterial(src, mat))
	}

	for _, node := range src.Nodes {
		if node.Mesh != nil {
			mqdoc.Objects = append(mqdoc.Objects, c.convertMesh(src, src.Meshes[*node.Mesh]))
		}
	}

	bonemap := map[uint32]bool{}
	var bones []*mqo.Bone
	for _, skin := range src.Skins {
		if skin.Skeleton != nil {
			if _, exist := bonemap[*skin.Skeleton]; !exist {
				bonemap[*skin.Skeleton] = true
				bones = c.convertBones(src, *skin.Skeleton, 0, bones)
			}
		}
	}
	mqo.GetBonePlugin(mqdoc).SetBones(bones)

	return mqdoc, nil
}
