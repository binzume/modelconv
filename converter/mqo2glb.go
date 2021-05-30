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
	"image/color"
	"image/png"

	_ "image/gif"
	_ "image/jpeg"

	_ "github.com/ftrvxmtrx/tga"
	"golang.org/x/image/bmp"
)

type MQOToGLTFOption struct {
	ForceUnlit bool
}

type mqoToGltf struct {
	*modeler.Modeler
	options         *MQOToGLTFOption
	scale           float32
	convertBone     bool
	convertMorph    bool
	JointNodeToBone map[uint32]*mqo.Bone
}

func NewMQOToGLTFConverter(options *MQOToGLTFOption) *mqoToGltf {
	if options == nil {
		options = &MQOToGLTFOption{}
	}
	return &mqoToGltf{
		Modeler:      modeler.NewModeler(),
		scale:        0.001,
		convertBone:  true,
		convertMorph: true,
		options:      options,
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

func (m *mqoToGltf) getNormal(obj *mqo.Object) [][3]float32 {
	normals := obj.GetSmoothNormals()
	normalArray := make([][3]float32, len(obj.Vertexes))
	for i, n := range normals {
		n.ToArray(normalArray[i][:])
	}
	return normalArray
}

func (m *mqoToGltf) getWeights(bones []*mqo.Bone, obj *mqo.Object, boneIDToJoint map[int]uint32) ([]uint32, [][4]uint16, [][4]float32) {
	var joints [][4]uint16
	var weights [][4]float32
	var njoint []int
	var jointIds []uint32

	vs := len(obj.Vertexes)
	for _, b := range bones {
		for _, bw := range b.Weights {
			if bw.ObjectID != obj.UID {
				continue
			}
			if joints == nil {
				joints = make([][4]uint16, vs)
				weights = make([][4]float32, vs)
				njoint = make([]int, vs)
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

func (m *mqoToGltf) hasAlpha(textureDir string, texture string) bool {
	if strings.HasSuffix(texture, ".jpg") || strings.HasSuffix(texture, ".bmp") {
		return false
	}
	f, err := os.Open(filepath.Join(textureDir, texture))
	if err != nil {
		return false
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return false
	}
	switch img.ColorModel() {
	case color.YCbCrModel, color.CMYKModel:
		return false
	case color.RGBAModel:
		return !img.(*image.RGBA).Opaque()
	}
	return false
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
			if ext == ".bmp" {
				// retry
				f.Seek(0, io.SeekStart)
				img, err = bmp.Decode(r)
			}
			if err != nil {
				log.Fatalf("Texture read error: %v texid: %v", err, texture)
			}
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

func (m *mqoToGltf) convertMaterial(textureDir string, mat *mqo.Material) *gltf.Material {
	var unlitMaterialExt = "KHR_materials_unlit"
	var rf = 0.4
	var mf = float64(mat.Specular)
	mm := &gltf.Material{
		Name: mat.Name,
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &gltf.RGBA{R: float64(mat.Color.X), G: float64(mat.Color.Y), B: float64(mat.Color.Z), A: float64(mat.Color.W)},
			RoughnessFactor: &rf,
			MetallicFactor:  &mf,
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
			cutoff := mat.Ex2.FloatParam("AlphaCutOff")
			mm.AlphaCutoff = &cutoff
		case 3:
			mm.AlphaMode = gltf.AlphaBlend
		}
		if metallicFactor == 0 && roughnessFactor == 1 {
			mm.Extensions = map[string]interface{}{unlitMaterialExt: map[string]string{}}
		}
	} else if mat.Color.W < 0.99 || m.hasAlpha(textureDir, mat.Texture) {
		mm.AlphaMode = gltf.AlphaBlend
	}
	if m.options.ForceUnlit {
		mm.Extensions = map[string]interface{}{unlitMaterialExt: map[string]string{}}
	}
	return mm
}

func (m *mqoToGltf) ConvertObject(obj *mqo.Object, bones []*mqo.Bone, boneIDToJoint map[int]uint32,
	morphObjs []*mqo.Object, ignoreMats map[int]bool) (*gltf.Mesh, []uint32) {
	scale := m.scale

	var vertexes [][3]float32
	var srcIndices []int
	for i, v := range obj.Vertexes {
		vertexes = append(vertexes, [3]float32{v.X * scale, v.Y * scale, v.Z * scale})
		srcIndices = append(srcIndices, i)
	}

	joints, joints0, weights0 := m.getWeights(bones, obj, boneIDToJoint)
	indices := map[int][]uint32{}
	normal := m.getNormal(obj)
	texcood0 := make([][2]float32, len(vertexes))
	type vertexKey struct {
		i  int
		uv mqo.Vector2
	}
	indicesMap := map[vertexKey]int{}
	useTexcood0 := false
	for _, f := range obj.Faces {
		if len(f.Verts) < 3 {
			continue
		}
		if _, ok := ignoreMats[f.Material]; ok {
			continue
		}
		verts := make([]int, len(f.Verts))
		copy(verts, f.Verts)
		if len(f.UVs) > 0 {
			useTexcood0 = true
			for i, index := range verts {
				if (texcood0[index][0] != 0 || texcood0[index][1] != 0) && (texcood0[index][0] != f.UVs[i].X || texcood0[index][1] != f.UVs[i].Y) {
					if ii, ok := indicesMap[vertexKey{index, f.UVs[i]}]; ok {
						verts[i] = ii
					} else {
						// copy attrs.
						verts[i] = len(vertexes)
						srcIndices = append(srcIndices, index)
						vertexes = append(vertexes, vertexes[index])
						texcood0 = append(texcood0, texcood0[index])
						normal = append(normal, normal[index])
						if len(joints) > 0 {
							joints0 = append(joints0, joints0[index])
							weights0 = append(weights0, weights0[index])
						}
						indicesMap[vertexKey{index, f.UVs[i]}] = verts[i]
					}
				}
				texcood0[verts[i]] = [2]float32{f.UVs[i].X, f.UVs[i].Y}
			}
		}
		// convex polygon only. TODO: triangulation.
		for n := 0; n < len(verts)-2; n++ {
			indices[f.Material] = append(indices[f.Material], uint32(verts[0]), uint32(verts[n+2]), uint32(verts[n+1]))
		}
	}

	attributes := map[string]uint32{
		"POSITION": m.AddPosition(0, vertexes),
	}
	if useTexcood0 {
		attributes["TEXCOORD_0"] = m.AddTextureCoord(0, texcood0)
	}
	if obj.Shading > 0 && !m.options.ForceUnlit {
		attributes["NORMAL"] = m.AddNormal(0, normal)
	}
	if len(joints) > 0 {
		attributes["JOINTS_0"] = m.AddJoints(0, joints0)
		attributes["WEIGHTS_0"] = m.AddWeights(0, weights0)
	}

	// morph
	var targets []map[string]uint32
	var targetNames []string
	for _, morphObj := range morphObjs {
		var mv [][3]float32
		for _, i := range srcIndices {
			v := morphObj.Vertexes[i]
			mv = append(mv, [3]float32{v.X*scale - vertexes[i][0], v.Y*scale - vertexes[i][1], v.Z*scale - vertexes[i][2]})
		}
		targets = append(targets, map[string]uint32{
			"POSITION": m.AddPosition(0, mv),
			"NORMAL":   attributes["NORMAL"], // for UniVRM. TODO
		})
		targetNames = append(targetNames, morphObj.Name)
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
	return &gltf.Mesh{
		Name:       obj.Name,
		Primitives: primitives,
		Extras:     map[string]interface{}{"targetNames": targetNames},
	}, joints
}

func (m *mqoToGltf) Convert(doc *mqo.Document, textureDir string) (*gltf.Document, error) {
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

	ignoreMats := map[int]bool{}
	for m, mat := range doc.Materials {
		if mat.Name == "$IGNORE" {
			ignoreMats[m] = true
		}
	}

	m.Nodes = make([]*gltf.Node, len(targetObjects))

	var bones []*mqo.Bone
	var boneIDToJoint map[int]uint32
	var jointToBone map[uint32]*mqo.Bone
	if m.convertBone {
		bones = mqo.GetBonePlugin(doc).Bones()
		boneIDToJoint, jointToBone = m.addBoneNodes(bones)
		m.JointNodeToBone = jointToBone
	}

	for i, obj := range targetObjects {
		var morphTargets []*mqo.Object
		if m.convertMorph {
			if morph, ok := morphBases[obj.Name]; ok {
				for _, t := range morph.Target {
					morphTargets = append(morphTargets, objectByName[t.Name])
				}
			}
		}
		mesh, joints := m.ConvertObject(obj, bones, boneIDToJoint, morphTargets, ignoreMats)
		node := &gltf.Node{Name: obj.Name}
		if len(mesh.Primitives) > 0 {
			node.Mesh = gltf.Index(uint32(len(m.Document.Meshes)))
			m.Document.Meshes = append(m.Document.Meshes, mesh)
		}
		if len(joints) > 0 {
			node.Skin = gltf.Index(m.addSkin(joints, jointToBone))
		}
		m.Nodes[i] = node
		m.Scenes[0].Nodes = append(m.Scenes[0].Nodes, uint32(i))
	}

	textures := map[string]uint32{}
	useUnlit := false
	for _, mat := range doc.Materials {
		mm := m.convertMaterial(textureDir, mat)
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
