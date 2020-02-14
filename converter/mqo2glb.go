package converter

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/mqo"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"image"
	"image/png"

	_ "image/gif"
	_ "image/jpeg"

	_ "github.com/ftrvxmtrx/tga"
	_ "golang.org/x/image/bmp"
)

type mqoToGltf struct {
	*modeler.Modeler
	scale       float32
	convertBone bool
}

func NewMQOToGLTFConverter() *mqoToGltf {
	return &mqoToGltf{
		Modeler:     modeler.NewModeler(),
		scale:       0.001,
		convertBone: true,
	}
}

func (m *mqoToGltf) addMatrices(bufferIndex uint32, mat [][4][4]float32) uint32 {
	a := make([][4]float32, len(mat)*4)
	for i, m := range mat {
		a[i*4+0] = m[0]
		a[i*4+1] = m[1]
		a[i*4+2] = m[2]
		a[i*4+3] = m[3]
	}
	acc := m.AddTangent(bufferIndex, a)
	m.Accessors[acc].Type = gltf.AccessorMat4
	m.Accessors[acc].Count /= 4
	m.BufferViews[*m.Accessors[acc].BufferView].ByteStride *= 4
	return acc
}

func (m *mqoToGltf) addBoneNodes(bones []*mqo.Bone) (map[int]uint32, map[uint32]*mqo.Bone) {
	scale := m.scale
	idmap := map[int]uint32{}
	idmapr := map[uint32]*mqo.Bone{}
	bonemap := map[int]*mqo.Bone{}
	for _, b := range bones {
		idmap[b.ID] = uint32(len(m.Nodes))
		idmapr[uint32(len(m.Nodes))] = b
		bonemap[b.ID] = b
		m.Nodes = append(m.Nodes, &gltf.Node{Name: b.Name, Translation: [3]float64{0, 0, 0}, Rotation: [4]float64{0, 0, 0, 1}})
	}

	for _, b := range bones {
		node := m.Nodes[idmap[b.ID]]
		if b.Parent > 0 {
			parent := bonemap[b.Parent]
			node.Translation[0] = float64((b.Pos.X - parent.Pos.X) * scale)
			node.Translation[1] = float64((b.Pos.Y - parent.Pos.Y) * scale)
			node.Translation[2] = float64((b.Pos.Z - parent.Pos.Z) * scale)
			parentNode := m.Nodes[idmap[parent.ID]]
			parentNode.Children = append(parentNode.Children, idmap[b.ID])
		} else {
			node.Translation[0] = float64((b.Pos.X) * scale)
			node.Translation[1] = float64((b.Pos.Y) * scale)
			node.Translation[2] = float64((b.Pos.Z) * scale)
			m.Scenes[0].Nodes = append(m.Scenes[0].Nodes, idmap[b.ID])
		}
	}
	return idmap, idmapr
}

func (m *mqoToGltf) getWeights(bones []*mqo.Bone, obj *mqo.Object, vs int, boneIDToJoint map[int]uint32) ([]uint32, [][4]uint16, [][4]float32) {
	joints := make([][4]uint16, vs)
	weights := make([][4]float32, vs)
	njoint := make([]int, vs)
	var jointIds []uint32

	for _, b := range bones {
		for _, bw := range b.Weights {
			if bw.ObjectID != obj.UID {
				continue
			}
			jointIds = append(jointIds, boneIDToJoint[b.ID])
			for _, vw := range bw.Vertexes {
				v := obj.GetVertexIndexByID(vw.VertexID)
				if v < 0 || v >= vs || njoint[v] >= 4 {
					log.Fatal("invalid weight. V:", vw.VertexID, " O:", obj.Name)
				}
				joints[v][njoint[v]] = uint16(len(jointIds)) - 1
				weights[v][njoint[v]] = vw.Weight * 0.01
				njoint[v]++
			}
		}
	}
	return jointIds, joints, weights
}

func (m *mqoToGltf) addSkin(joints []uint32, jointToBone map[uint32]*mqo.Bone) uint32 {
	invmats := make([][4][4]float32, len(joints))
	scale := m.scale
	for i, j := range joints {
		b := jointToBone[j]
		invmats[i] = [4][4]float32{
			{1, 0, 0, 0},
			{0, 1, 0, 0},
			{0, 0, 1, 0},
			{-b.Pos.X * scale, -b.Pos.Y * scale, -b.Pos.Z * scale, 1},
		}
	}
	m.Skins = append(m.Skins, &gltf.Skin{
		Joints:              joints,
		InverseBindMatrices: gltf.Index(m.addMatrices(0, invmats)),
	})
	return uint32(len(m.Skins) - 1)
}

func (m *mqoToGltf) addTexture(textureDir string, texture string) uint32 {
	f, err := os.Open(filepath.Join(textureDir, texture))
	if err != nil {
		log.Print("Texture file not found:", texture)
	}
	defer f.Close()
	var r io.Reader = f
	var mimeType string
	ext := strings.ToLower(filepath.Ext(texture))
	if ext == ".jpg" || ext == ".jpeg" {
		mimeType = "image/jpeg"
	} else if ext == ".png" {
		mimeType = "image/png"
	} else {
		mimeType = "image/png"
		img, _, err := image.Decode(r)
		if err != nil {
			log.Fatal("Texture read error:", err, texture)
		}
		w := new(bytes.Buffer)
		err = png.Encode(w, img)
		if err != nil {
			log.Fatal("Texture encode error:", err, texture)
		}
		r = w
	}
	img, err := m.AddImage(0, filepath.Base(texture), mimeType, r)
	if err != nil {
		log.Fatal("Texture read error:", err, texture)
	}
	m.Buffers[0].ByteLength = uint32(len(m.Buffers[0].Data)) // avoid AddImage bug
	m.Textures = append(m.Textures,
		&gltf.Texture{Sampler: gltf.Index(0), Source: gltf.Index(img)})

	return uint32(len(m.Textures)) - 1
}

func (m *mqoToGltf) convertMaterial(mat *mqo.Material) *gltf.Material {
	mm := &gltf.Material{
		Name: mat.Name,
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &gltf.RGBA{R: float64(mat.Color.X), G: float64(mat.Color.Y), B: float64(mat.Color.Z), A: float64(mat.Color.W)},
		},
		DoubleSided: mat.DoubleSided,
	}
	if mat.EmissionColor != nil {
		mm.EmissiveFactor = [3]float64{float64(mat.EmissionColor.X), float64(mat.EmissionColor.Y), float64(mat.EmissionColor.Z)}
	} else if mat.Emission > 0 {
		mm.EmissiveFactor = [3]float64{float64(mat.Emission), float64(mat.Emission), float64(mat.Emission)}
	}
	if mat.Ex2 != nil && mat.Ex2.ShaderName == "glTF" {
		metallicFactor := mat.Ex2.FloatParam("Metallic")
		mm.PBRMetallicRoughness.MetallicFactor = &metallicFactor
		roughnessFactor := mat.Ex2.FloatParam("Roughness")
		mm.PBRMetallicRoughness.RoughnessFactor = &roughnessFactor
		switch mat.Ex2.IntParam("AlphaMode") {
		case 1:
			mm.AlphaMode = gltf.AlphaOpaque
		case 2:
			mm.AlphaMode = gltf.AlphaMask
			cutoff := mat.Ex2.FloatParam("AlphaCutoff")
			mm.AlphaCutoff = &cutoff
		case 3:
			mm.AlphaMode = gltf.AlphaBlend
		}
		if metallicFactor == 0 && roughnessFactor == 1 {
			var unlitMaterialExt = "KHR_materials_unlit"
			mm.Extensions = map[string]interface{}{unlitMaterialExt: map[string]string{}}
		}
	} else if mat.Color.W != 1 {
		mm.AlphaMode = gltf.AlphaBlend
	} else if strings.HasSuffix(mat.Texture, ".tga") || strings.HasSuffix(mat.Texture, ".png") {
		// TODO: check texture alpha.
		mm.AlphaMode = gltf.AlphaMask
	}
	return mm
}

func (m *mqoToGltf) Convert(doc *mqo.Document, textureDir string) (*gltf.Document, error) {
	scale := m.scale

	objectByName := map[string]*mqo.Object{}
	morphTargets := map[string]*mqo.Object{}
	morphBases := map[string]*mqo.MorphTargetList{}
	for _, obj := range doc.Objects {
		objectByName[obj.Name] = obj
	}

	morphs := mqo.GetMorphPlugin(doc).Morphs()
	for _, m := range morphs {
		morphBases[m.Base] = m
		for _, t := range m.Target {
			morphTargets[t.Name] = objectByName[t.Name]
		}
	}

	doc.FixObjectID()
	var targetObjects []*mqo.Object
	for _, obj := range doc.Objects {
		if obj.Visible && len(obj.Faces) > 0 && morphTargets[obj.Name] == nil {
			targetObjects = append(targetObjects, obj)
		}
	}

	m.Nodes = make([]*gltf.Node, len(targetObjects))

	var bones []*mqo.Bone
	if m.convertBone {
		bones = mqo.GetBonePlugin(doc).Bones()
	}
	boneIDToJoint, jointToBone := m.addBoneNodes(bones)

	for i, obj := range targetObjects {
		var vertexes [][3]float32
		for _, v := range obj.Vertexes {
			vertexes = append(vertexes, [3]float32{v.X * scale, v.Y * scale, v.Z * scale})
		}

		texcood := make([][2]float32, len(vertexes))
		indices := map[int][]uint32{}
		for _, f := range obj.Faces {
			indices[f.Material] = append(indices[f.Material], uint32(f.Verts[2]), uint32(f.Verts[1]), uint32(f.Verts[0]))
			for i, index := range f.Verts {
				texcood[index] = [2]float32{f.UVs[i].X, f.UVs[i].Y}
			}
		}

		attributes := map[string]uint32{
			"POSITION":   m.AddPosition(0, vertexes),
			"TEXCOORD_0": m.AddTextureCoord(0, texcood),
		}

		joints, j, w := m.getWeights(bones, obj, len(vertexes), boneIDToJoint)
		if len(joints) > 0 {
			attributes["JOINTS_0"] = m.AddJoints(0, j)
			attributes["WEIGHTS_0"] = m.AddWeights(0, w)
		}

		var targets []map[string]uint32
		var targetNames []string
		if morph, ok := morphBases[obj.Name]; ok {
			for _, t := range morph.Target {
				var mv [][3]float32
				for i, v := range objectByName[t.Name].Vertexes {
					mv = append(mv, [3]float32{v.X*scale - vertexes[i][0], v.Y*scale - vertexes[i][1], v.Z*scale - vertexes[i][2]})
				}
				targets = append(targets, map[string]uint32{
					"POSITION": m.AddPosition(0, mv),
				})
				targetNames = append(targetNames, t.Name)
			}
		}

		// make primitive for each materials
		var primitives []*gltf.Primitive
		for mat, ind := range indices {
			indicesAccessor := m.AddIndices(0, ind)
			primitives = append(primitives, &gltf.Primitive{
				Indices:    gltf.Index(indicesAccessor),
				Attributes: attributes,
				Material:   gltf.Index(uint32(mat)),
				Targets:    targets,
			})
		}
		mesh := &gltf.Mesh{
			Name:       obj.Name,
			Primitives: primitives,
			Extras:     map[string]interface{}{"targetNames": targetNames},
		}
		meshIndex := uint32(len(m.Document.Meshes))
		m.Document.Meshes = append(m.Document.Meshes, mesh)
		m.Nodes[i] = &gltf.Node{Name: obj.Name, Mesh: gltf.Index(meshIndex)}
		m.Scenes[0].Nodes = append(m.Scenes[0].Nodes, uint32(i))

		if len(joints) > 0 {
			m.Nodes[i].Skin = gltf.Index(m.addSkin(joints, jointToBone))
		}
	}

	textures := map[string]uint32{}
	useUnlit := false
	for _, mat := range doc.Materials {
		mm := m.convertMaterial(mat)
		if mat.Texture != "" {
			if _, exist := textures[mat.Texture]; !exist {
				textures[mat.Texture] = m.addTexture(textureDir, mat.Texture)
			}
			mm.PBRMetallicRoughness.BaseColorTexture = &gltf.TextureInfo{
				Index: textures[mat.Texture],
			}
		}
		if mm.Extensions["KHR_materials_unlit"] != nil {
			useUnlit = true
		}
		m.Document.Materials = append(m.Document.Materials, mm)
	}
	if useUnlit {
		m.ExtensionsUsed = append(m.ExtensionsUsed, "KHR_materials_unlit")
	}

	if len(m.Document.Textures) > 0 {
		m.Document.Samplers = []*gltf.Sampler{{}}
	}

	return m.Document, nil
}
