package converter

import (
	"bytes"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/mqo"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"image"
	"image/color"
	"image/jpeg"
	"image/png"

	_ "image/gif"
	_ "image/jpeg"

	"github.com/blezek/tga"
	_ "github.com/oov/psd"
	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
)

var TextureUVEpsilon float32 = 0.0001

type MQOToGLTFOption struct {
	Scale      float32 // Default: 0.001
	ForceUnlit bool

	TextureReCompress      bool
	TextureBytesThreshold  int64 // 0: unlimited
	TextureResolutionLimit int   // 0: unlimited
	TextureScale           float32
	IgnoreObjectHierarchy  bool
	DetectAlphaTexture     bool

	ExportLights   bool
	ReuseGeometry  bool // experimental
	ConvertPhysics bool // experimental. BLENDER_physics?
}

type mqoToGltf struct {
	*MQOToGLTFOption
	*gltf.Document
	convertBone     bool
	convertMorph    bool
	JointNodeToBone map[uint32]*mqo.Bone
	extensions      map[string]bool
}

type textureCache struct {
	srcDir   string
	textures map[string]*textureInfo
}

type textureInfo struct {
	name string
	id   *uint32
	img  image.Image
	err  error
}

type uvrect struct {
	top, bottom, left, right float32
}

func newUVRectFromPoint(p geom.Vector2) *uvrect {
	x := p.X - float32(math.Floor(float64(p.X)))
	y := p.Y - float32(math.Floor(float64(p.Y)))
	return &uvrect{y, y, x, x}
}

func (b *uvrect) Add(p geom.Vector2) {
	x := p.X - float32(math.Floor(float64(p.X)))
	y := p.Y - float32(math.Floor(float64(p.Y)))
	if x < b.left {
		b.left = x
	}
	if x > b.right {
		b.right = x
	}
	if y < b.top {
		b.top = y
	}
	if y > b.bottom {
		b.bottom = y
	}
}

const BlenderPhysicsName = "BLENDER_physics"

type BlenderPhysicsBody struct {
	Shapes          []map[string]interface{} `json:"collisionShapes"`
	Mass            float32                  `json:"mass"`
	Static          bool                     `json:"static"`
	CollisionGroups int                      `json:"collisionGroups"`
	CollisionMasks  int                      `json:"collisionMasks"`
}

func (c *textureCache) get(name string) *textureInfo {
	if t, ok := c.textures[name]; ok {
		return t
	}
	t := &textureInfo{name: name}
	c.textures[name] = t
	return t
}

func (c *textureCache) getImage(name string) (image.Image, error) {
	t := c.get(name)
	if t.img != nil || t.err != nil {
		return t.img, t.err
	}

	f, err := os.Open(filepath.Join(c.srcDir, t.name))
	if err != nil {
		t.err = err
		return nil, err
	}
	defer f.Close()

	if strings.ToLower(filepath.Ext(t.name)) == ".tga" {
		t.img, t.err = tga.Decode(f)
	} else {
		t.img, _, t.err = image.Decode(f)
	}
	return t.img, t.err
}

type geomCache struct {
	count      int
	attributes map[string]uint32
	indices    []uint32
	matrix     *geom.Matrix4
}

func NewMQOToGLTFConverter(options *MQOToGLTFOption) *mqoToGltf {
	if options == nil {
		options = &MQOToGLTFOption{}
	}
	if options.Scale == 0 {
		options.Scale = 0.001
	}
	if options.TextureScale == 0 {
		options.TextureScale = 1.0
	}
	return &mqoToGltf{
		MQOToGLTFOption: options,
		Document:        gltf.NewDocument(),
		convertBone:     true,
		convertMorph:    true,
		extensions:      map[string]bool{},
	}
}

func (m *mqoToGltf) addMatrices(mat [][4][4]float32) uint32 {
	a := make([][4]float32, len(mat)*4)
	for i, m := range mat {
		a[i*4+0] = m[0]
		a[i*4+1] = m[1]
		a[i*4+2] = m[2]
		a[i*4+3] = m[3]
	}
	acc := modeler.WriteTangent(m.Document, a)
	m.Accessors[acc].Type = gltf.AccessorMat4
	m.Accessors[acc].Count /= 4
	m.BufferViews[*m.Accessors[acc].BufferView].ByteStride *= 4
	return acc
}

func (m *mqoToGltf) addBoneNodes(bones []*mqo.Bone) (map[int]uint32, map[uint32]*mqo.Bone) {
	scale := m.Scale
	idmap := map[int]uint32{}
	idmapr := map[uint32]*mqo.Bone{}
	bonemap := map[int]*mqo.Bone{}
	for _, b := range bones {
		idmap[b.ID] = uint32(len(m.Nodes))
		idmapr[uint32(len(m.Nodes))] = b
		bonemap[b.ID] = b
		m.Nodes = append(m.Nodes, &gltf.Node{Name: b.Name, Translation: [3]float32{0, 0, 0}, Rotation: [4]float32{0, 0, 0, 1}})
	}

	for _, b := range bones {
		node := m.Nodes[idmap[b.ID]]
		if b.Parent > 0 {
			parent := bonemap[b.Parent]
			node.Translation[0] = ((b.Pos.X - parent.Pos.X) * scale)
			node.Translation[1] = ((b.Pos.Y - parent.Pos.Y) * scale)
			node.Translation[2] = ((b.Pos.Z - parent.Pos.Z) * scale)
			parentNode := m.Nodes[idmap[parent.ID]]
			parentNode.Children = append(parentNode.Children, idmap[b.ID])
		} else {
			node.Translation[0] = ((b.Pos.X) * scale)
			node.Translation[1] = ((b.Pos.Y) * scale)
			node.Translation[2] = ((b.Pos.Z) * scale)
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
				jindex := njoint[v]
				njoint[v]++
				if jindex >= 4 {
					// Overwrite smallest weight.
					minWeight := vw.Weight * 0.01
					for i, w := range weights[v] {
						if w < minWeight {
							minWeight = w
							jindex = i
						}
					}
					if jindex >= 4 {
						continue
					}
				}
				if v < 0 || v >= vs {
					log.Fatal("invalid weight. V:", vw.VertexID, " O:", obj.Name)
				}
				joints[v][jindex] = uint16(len(jointIds)) - 1
				weights[v][jindex] = vw.Weight * 0.01
			}
			for _, vw := range bw.Vertexes {
				v := obj.GetVertexIndexByID(vw.VertexID)
				if njoint[v] > 4 {
					log.Println("WWARNING: njoint > 4. V:", vw.VertexID, " O:", obj.Name)
					var sum float32 = 0
					for _, w := range weights[v] {
						sum += w
					}
					if sum > 0 {
						for i := range weights[v] {
							weights[v][i] /= sum
						}
					}
				}
			}

		}
	}
	return jointIds, joints, weights
}

func (m *mqoToGltf) addSkin(joints []uint32, jointToBone map[uint32]*mqo.Bone) uint32 {
	invmats := make([][4][4]float32, len(joints))
	scale := m.Scale
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
		InverseBindMatrices: gltf.Index(m.addMatrices(invmats)),
	})
	return uint32(len(m.Skins) - 1)
}

func (m *mqoToGltf) hasAlpha(texture string, textures *textureCache, rect *uvrect) bool {
	if texture == "" || strings.HasSuffix(texture, ".jpg") || strings.HasSuffix(texture, ".bmp") {
		return false
	}
	img, err := textures.getImage(texture)
	if err != nil {
		return false
	}
	if rect != nil {
		if s, ok := img.(interface {
			SubImage(r image.Rectangle) image.Image
		}); ok {
			b := img.Bounds()
			r := image.Rectangle{
				Min: image.Point{X: int(float32(b.Dx()) * rect.left), Y: int(float32(b.Dy()) * rect.top)},
				Max: image.Point{X: int(float32(b.Dx()) * rect.right), Y: int(float32(b.Dy()) * rect.bottom)},
			}
			img = s.SubImage(r)
		}
	}
	switch img.ColorModel() {
	case color.YCbCrModel, color.CMYKModel:
		return false
	case color.RGBAModel:
		return !img.(*image.RGBA).Opaque()
	case color.NRGBAModel:
		return !img.(*image.NRGBA).Opaque()
	}
	return false
}

func scaleTexture(texture string, mime string, textures *textureCache, scale float32, limit int) (io.Reader, error) {
	img, err := textures.getImage(texture)
	if err != nil {
		return nil, err
	}
	rect := img.Bounds()

	if limit > 0 {
		sz := int(float32(rect.Dx()) * scale)
		if sz > limit {
			scale *= float32(limit) / float32(sz)
		}
	}

	if scale != 1.0 {
		dst := image.NewRGBA(image.Rect(0, 0, int(float32(rect.Dx())*scale), int(float32(rect.Dy())*scale)))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, rect, draw.Over, nil)
		img = dst
	}

	w := new(bytes.Buffer)
	if mime == "image/png" {
		err = png.Encode(w, img)
	} else {
		err = jpeg.Encode(w, img, nil)
	}
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (m *mqoToGltf) addTexture(texture string, textures *textureCache) (*uint32, error) {
	t := textures.get(texture)
	if t.id != nil {
		return t.id, nil
	}
	ext := strings.ToLower(filepath.Ext(texture))
	path := texture
	if !filepath.IsAbs(path) {
		path = filepath.Join(textures.srcDir, texture)
	}

	encode := m.TextureReCompress
	if m.TextureBytesThreshold > 0 {
		stat, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if stat.Size() > m.TextureBytesThreshold {
			encode = true
		}
	}

	var mimeType string
	if ext == ".jpg" || ext == ".jpeg" {
		mimeType = "image/jpeg"
	} else if ext == ".png" {
		mimeType = "image/png"
	} else {
		mimeType = "image/png"
		encode = true
	}

	var r io.Reader
	if encode {
		r2, err := scaleTexture(texture, mimeType, textures, m.TextureScale, m.TextureResolutionLimit)
		if err != nil {
			return nil, err
		}
		r = r2
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	img, err := modeler.WriteImage(m.Document, filepath.Base(texture), mimeType, r)
	if err != nil {
		return nil, err
	}
	m.Buffers[0].ByteLength = uint32(len(m.Buffers[0].Data)) // avoid AddImage bug
	m.Textures = append(m.Textures,
		&gltf.Texture{Sampler: gltf.Index(0), Source: gltf.Index(img)})

	t.id = gltf.Index(uint32(len(m.Textures)) - 1)

	return t.id, nil
}
func (m *mqoToGltf) tryAddTexture(texturePath string, textures *textureCache) *gltf.TextureInfo {
	if texturePath == "" {
		return nil
	}
	if tex, err := m.addTexture(texturePath, textures); err == nil {
		return &gltf.TextureInfo{Index: *tex}
	} else {
		log.Print("Texture read error:", texturePath, err)
	}
	return nil
}

func (m *mqoToGltf) convertMaterial(mat *mqo.Material, textures *textureCache, bounds *uvrect) *gltf.Material {
	var unlitMaterialExt = "KHR_materials_unlit"
	var rf float32 = 0.4
	var mf = mat.Specular
	mm := &gltf.Material{
		Name: mat.Name,
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{mat.Color.X, mat.Color.Y, mat.Color.Z, mat.Color.W},
			RoughnessFactor: &rf,
			MetallicFactor:  &mf,
		},
		DoubleSided: mat.DoubleSided,
	}
	if mat.EmissionColor != nil || mat.Emission > 0 {
		if mat.EmissionColor != nil {
			mm.EmissiveFactor = [3]float32{(mat.EmissionColor.X), (mat.EmissionColor.Y), (mat.EmissionColor.Z)}
		} else {
			mm.EmissiveFactor = [3]float32{(mat.Emission), (mat.Emission), (mat.Emission)}
		}
		s := geom.Max(geom.Max(mm.EmissiveFactor[0], mm.EmissiveFactor[1]), mm.EmissiveFactor[2])
		if s > 1 {
			mm.EmissiveFactor[0] /= s
			mm.EmissiveFactor[1] /= s
			mm.EmissiveFactor[2] /= s
		}
	}

	addExt := func(name string, ext interface{}) {
		if mm.Extensions == nil {
			mm.Extensions = map[string]interface{}{}
		}
		mm.Extensions[name] = ext
		m.extensions[name] = true
	}
	if mat.GetShaderName() == "glTF" {
		metallicFactor := float32(mat.Ex2.FloatParam("Metallic"))
		mm.PBRMetallicRoughness.MetallicFactor = &metallicFactor
		roughnessFactor := float32(mat.Ex2.FloatParam("Roughness"))
		mm.PBRMetallicRoughness.RoughnessFactor = &roughnessFactor
		switch mat.Ex2.IntParam("AlphaMode") {
		case 1:
			mm.AlphaMode = gltf.AlphaOpaque
		case 2:
			mm.AlphaMode = gltf.AlphaMask
			cutoff := float32(mat.Ex2.FloatParam("AlphaCutOff"))
			mm.AlphaCutoff = &cutoff
		case 3:
			mm.AlphaMode = gltf.AlphaBlend
		}

		if mat.Ex2.BoolParam("Extensions.Unlit") || metallicFactor == 0 && roughnessFactor == 1 {
			addExt(unlitMaterialExt, map[string]string{})
		}
		if mat.Ex2.BoolParam("Extensions.SpecularExt") {
			ext := map[string]interface{}{}
			if v := mat.Ex2.FloatParam("Extensions.SpecularFactor"); v != 0 {
				ext["specularFactor"] = v
			}
			if c := mat.Ex2.ColorParam("Extensions.SpecularColorFactor"); c != nil {
				ext["specularColorFactor"] = c[0:3]
			}
			if t := m.tryAddTexture(mat.Ex2.Mapping("Specular"), textures); t != nil {
				ext["specularTexture"] = t
			}
			if t := m.tryAddTexture(mat.Ex2.Mapping("SpecularColor"), textures); t != nil {
				ext["specularColorTexture"] = t
			}
			addExt("KHR_materials_specular", ext)
		}
		if mat.Ex2.BoolParam("Extensions.Clearcoat") {
			ext := map[string]interface{}{}
			if v := mat.Ex2.FloatParam("Extensions.ClearcoatFactor"); v != 0 {
				ext["clearcoatFactor"] = v
			}
			if v := mat.Ex2.FloatParam("Extensions.ClearcoatRoughnessFactor"); v != 0 {
				ext["clearcoatRoughnessFactor"] = v
			}
			if t := m.tryAddTexture(mat.Ex2.Mapping("Clearcoat"), textures); t != nil {
				ext["clearcoatTexture"] = t
			}
			if t := m.tryAddTexture(mat.Ex2.Mapping("ClearcoatRoughness"), textures); t != nil {
				ext["clearcoatRoughnessTexture"] = t
			}
			if t := m.tryAddTexture(mat.Ex2.Mapping("ClearcoatNormal"), textures); t != nil {
				ext["clearcoatNormalTexture"] = t
			}
			addExt("KHR_materials_clearcoat", ext)
		}
		if mat.Ex2.BoolParam("Extensions.Sheen") {
			ext := map[string]interface{}{}
			if c := mat.Ex2.ColorParam("Extensions.SheenColorFactor"); c != nil {
				ext["sheenColorFactor"] = c[0:3]
			}
			if v := mat.Ex2.FloatParam("Extensions.SheenRoughnessFactor"); v != 0 {
				ext["sheenRoughnessFactor"] = v
			}
			if t := m.tryAddTexture(mat.Ex2.Mapping("SheenColor"), textures); t != nil {
				ext["sheenColorTexture"] = t
			}
			if t := m.tryAddTexture(mat.Ex2.Mapping("SheenRoughness"), textures); t != nil {
				ext["sheenRoughnessTexture"] = t
			}
			addExt("KHR_materials_sheen", ext)
		}
		if mat.Ex2.BoolParam("Extensions.IorExt") {
			ext := map[string]interface{}{}
			if v := mat.Ex2.FloatParam("Extensions.Ior"); v != 0 {
				ext["ior"] = v
			}
			addExt("KHR_materials_ior", ext)
		}
		if mat.Ex2.BoolParam("Extensions.Volume") {
			ext := map[string]interface{}{}
			if v := mat.Ex2.FloatParam("Extensions.ThicknessFactor"); v != 0 {
				ext["thicknessFactor"] = v
			}
			if v := mat.Ex2.FloatParam("Extensions.AttenuationDistance"); v != 0 {
				ext["attenuationDistance"] = v
			}
			if c := mat.Ex2.ColorParam("Extensions.AttenuationColor"); c != nil {
				ext["attenuationColor"] = c[0:3]
			}
			if t := m.tryAddTexture(mat.Ex2.Mapping("Thickness"), textures); t != nil {
				ext["thicknessTexture"] = t
			}
			addExt("KHR_materials_volume", ext)
		}
		if mat.Ex2.BoolParam("Extensions.Transmission") {
			ext := map[string]interface{}{}
			if v := mat.Ex2.FloatParam("Extensions.TransmissionFactor"); v != 0 {
				ext["transmissionFactor"] = v
			}
			if t := m.tryAddTexture(mat.Ex2.Mapping("Transmission"), textures); t != nil {
				ext["transmissionTexture"] = t
			}
			addExt("KHR_materials_transmission", ext)
		}
	} else if mat.Color.W < 0.99 || (m.DetectAlphaTexture && m.hasAlpha(mat.Texture, textures, bounds)) {
		mm.AlphaMode = gltf.AlphaBlend
	}
	if m.ForceUnlit || mat.GetShaderName() == "Constant" {
		addExt(unlitMaterialExt, map[string]string{})
	}

	mm.PBRMetallicRoughness.BaseColorTexture = m.tryAddTexture(mat.Texture, textures)
	if t := m.tryAddTexture(mat.BumpTexture, textures); t != nil {
		mm.NormalTexture = &gltf.NormalTexture{Index: &t.Index}
	}
	return mm
}

func (m *mqoToGltf) ConvertObject(obj *mqo.Object, bones []*mqo.Bone, boneIDToJoint map[int]uint32,
	morphObjs []*mqo.Object, materialMap map[int]int, shared *geomCache, uvBounds map[int]*uvrect) (*gltf.Mesh, []uint32) {
	scale := m.Scale
	obj.FixhNormals()
	obj.Triangulate()

	var vertexes [][3]float32
	var srcIndices []int
	for i, v := range obj.Vertexes {
		vertexes = append(vertexes, [3]float32{v.X * scale, v.Y * scale, v.Z * scale})
		srcIndices = append(srcIndices, i)
	}

	joints, joints0, weights0 := m.getWeights(bones, obj, boneIDToJoint)
	var materials []int
	indices := map[int][]uint32{}
	normals := make([][3]float32, len(vertexes))
	texcood0 := make([][2]float32, len(vertexes))
	indicesMap := map[int][]int{}
	useTexcood0 := false
	partial := false
	for _, f := range obj.Faces {
		if len(f.Verts) < 3 {
			continue
		}
		mat, ok := materialMap[f.Material]
		if !ok {
			partial = true
			continue
		}
		verts := make([]int, len(f.Verts))
		copy(verts, f.Verts)
		if len(f.UVs) > 0 {
			useTexcood0 = true
			bounds := uvBounds[f.Material]
			if bounds == nil {
				bounds = newUVRectFromPoint(f.UVs[0])
				uvBounds[f.Material] = bounds
			}

			for i, index := range verts {
				assigned := indicesMap[index]
				normal := f.Normals[i]
				if assigned != nil {
					vi := -1
					for _, v := range assigned {
						if f.UVs[i].Sub(&geom.Vector2{X: texcood0[v][0], Y: texcood0[v][1]}).LenSqr() < TextureUVEpsilon {
							verts[i] = v
							vi = v
						}
					}
					if vi >= 0 {
						continue
					}
					verts[i] = len(vertexes)
					srcIndices = append(srcIndices, index)
					vertexes = append(vertexes, vertexes[index])
					texcood0 = append(texcood0, [2]float32{})
					normals = append(normals, [3]float32{})
					if len(joints) > 0 {
						joints0 = append(joints0, joints0[index])
						weights0 = append(weights0, weights0[index])
					}
				}
				indicesMap[index] = append(indicesMap[index], verts[i])
				texcood0[verts[i]] = [2]float32{f.UVs[i].X, f.UVs[i].Y}
				normal.ToArray(normals[verts[i]][:])
				bounds.Add(f.UVs[i])
			}
		} else {
			for i, index := range verts {
				f.Normals[i].ToArray(normals[index][:])
			}
		}
		if _, exists := indices[mat]; !exists {
			materials = append(materials, mat)
		}
		indices[mat] = append(indices[mat], uint32(verts[2]), uint32(verts[1]), uint32(verts[0]))
	}

	attributes := map[string]uint32{}

	if !partial && shared != nil && shared.attributes != nil {
		attributes["POSITION"] = shared.attributes["POSITION"]
		if useTexcood0 {
			attributes["TEXCOORD_0"] = shared.attributes["TEXCOORD_0"]
		}
		if obj.Shading > 0 && !m.ForceUnlit {
			attributes["NORMAL"] = shared.attributes["NORMAL"]
		}
	} else if len(vertexes) > 0 {
		attributes["POSITION"] = modeler.WritePosition(m.Document, vertexes)
		if useTexcood0 {
			attributes["TEXCOORD_0"] = modeler.WriteTextureCoord(m.Document, texcood0)
		}
		if obj.Shading > 0 && !m.ForceUnlit {
			attributes["NORMAL"] = modeler.WriteNormal(m.Document, normals)
		}
	}

	if len(joints) > 0 {
		attributes["JOINTS_0"] = modeler.WriteJoints(m.Document, joints0)
		attributes["WEIGHTS_0"] = modeler.WriteWeights(m.Document, weights0)
	}

	if !partial && shared != nil && shared.attributes == nil {
		shared.attributes = attributes
		for _, mat := range materials {
			shared.indices = append(shared.indices, modeler.WriteIndices(m.Document, indices[mat]))
		}
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
			"POSITION": modeler.WritePosition(m.Document, mv),
			"NORMAL":   attributes["NORMAL"], // for UniVRM. TODO
		})
		targetNames = append(targetNames, morphObj.Name)
	}

	var indicesAccessors []uint32
	if !partial && shared != nil {
		indicesAccessors = shared.indices
	} else {
		for _, mat := range materials {
			indicesAccessors = append(indicesAccessors, modeler.WriteIndices(m.Document, indices[mat]))
		}
	}

	// make primitive for each materials
	var primitives []*gltf.Primitive
	for i, mat := range materials {
		primitives = append(primitives, &gltf.Primitive{
			Indices:    gltf.Index(indicesAccessors[i]),
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

func (m *mqoToGltf) checkMaterials(obj *mqo.Object, materials map[int]bool) {
	for _, f := range obj.Faces {
		materials[f.Material] = true
	}
}

func (m *mqoToGltf) Convert(doc *mqo.Document, textureDir string) (*gltf.Document, error) {
	objectByName := map[string]*mqo.Object{}
	morphTargets := map[string]*mqo.Object{}
	morphBases := map[string]*mqo.MorphTargetList{}
	uvBounds := map[int]*uvrect{}
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
	materialUsed := map[int]bool{}
	for _, obj := range doc.Objects {
		if (!m.IgnoreObjectHierarchy || obj.Visible && len(obj.Faces) > 0) && morphTargets[obj.Name] == nil {
			targetObjects = append(targetObjects, obj)
			if obj.Visible {
				m.checkMaterials(obj, materialUsed)
			}
		}
	}

	materialMap := map[int]int{}
	materialCount := 0
	for i, mat := range doc.Materials {
		if !materialUsed[i] {
			continue
		}
		if !strings.HasSuffix(mat.Name, "$IGNORE") && !strings.HasPrefix(mat.Name, "$MORPH:") {
			materialMap[i] = materialCount
			materialCount++
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
		if m.ConvertPhysics {
			physics := mqo.GetPhysicsPlugin(doc)
			for _, b := range physics.Bodies {
				if b.TargetBoneID == 0 {
					continue
				}
				if n, ok := boneIDToJoint[b.TargetBoneID]; ok {
					addPhysicsBody(m.Nodes[n], b, m.Scale, &jointToBone[n].Pos.Vector3)
				}
			}
		}
	}

	sharedGeoms := map[string]*geomCache{}
	for _, obj := range targetObjects {
		if hint := obj.SharedGeometryHint; hint != nil {
			if sharedGeoms[hint.Key] == nil {
				sharedGeoms[hint.Key] = &geomCache{}
			}
			sharedGeoms[hint.Key].count++
		}
	}

	var lights []map[string]interface{}
	var nodePath []*gltf.Node
	for i, obj := range targetObjects {
		var morphTargets []*mqo.Object
		if m.convertMorph {
			if morph, ok := morphBases[obj.Name]; ok {
				for _, t := range morph.Target {
					if len(objectByName[t.Name].Vertexes) != len(obj.Vertexes) {
						log.Print("unmached morph target: ", t.Name)
						continue
					}
					morphTargets = append(morphTargets, objectByName[t.Name])
				}
			}
		}

		node := &gltf.Node{Name: obj.Name}
		if obj.Visible {
			var shared *geomCache
			if hint := obj.SharedGeometryHint; m.ReuseGeometry && hint != nil {
				if sharedGeoms[hint.Key] != nil && sharedGeoms[hint.Key].count > 1 {
					shared = sharedGeoms[hint.Key]
					if shared.matrix == nil {
						shared.matrix = geom.NewScaleMatrix4(m.Scale, m.Scale, m.Scale).Mul(hint.Transform).Inverse()
					}
				}
			}
			mesh, joints := m.ConvertObject(obj, bones, boneIDToJoint, morphTargets, materialMap, shared, uvBounds)
			if len(mesh.Primitives) > 0 {
				node.Mesh = gltf.Index(uint32(len(m.Document.Meshes)))
				m.Document.Meshes = append(m.Document.Meshes, mesh)
			}
			if len(joints) > 0 {
				node.Skin = gltf.Index(m.addSkin(joints, jointToBone))
			}
			if shared != nil {
				geom.NewScaleMatrix4(m.Scale, m.Scale, m.Scale).Mul(obj.SharedGeometryHint.Transform).Mul(shared.matrix).ToArray(node.Matrix[:])
				// workaround: avoid missing mesh issue in Windows 3D viewer?
				node.Matrix[3] = 0
				node.Matrix[7] = 0
				node.Matrix[11] = 0
				node.Matrix[15] = 1
			}
			if m.ExportLights {
				if lightParam, ok := obj.Extra["light"].(map[string]interface{}); ok {
					if node.Extensions == nil {
						node.Extensions = gltf.Extensions{}
					}
					node.Extensions["KHR_lights_punctual"] = map[string]interface{}{"light": len(lights)}
					lights = append(lights, lightParam)
				}
			}
		}
		m.Nodes[i] = node
		// node.AddChild()
		if len(nodePath) > obj.Depth {
			nodePath = nodePath[:obj.Depth]
		}
		if !m.IgnoreObjectHierarchy && len(nodePath) > 0 {
			parent := nodePath[len(nodePath)-1]
			parent.Children = append(parent.Children, uint32(i))
		} else {
			m.Scenes[0].Nodes = append(m.Scenes[0].Nodes, uint32(i))
		}
		if node.MatrixOrDefault() == gltf.DefaultMatrix {
			nodePath = append(nodePath, node)
		}

		// Physics
		if m.ConvertPhysics {
			physics := mqo.GetPhysicsPlugin(doc)
			for _, b := range physics.Bodies {
				if b.TargetObjID != 0 && b.TargetObjID == obj.UID {
					addPhysicsBody(node, b, m.Scale, geom.NewVector3(0, 0, 0))
				}
			}
		}
	}

	if len(lights) > 0 {
		m.extensions["KHR_lights_punctual"] = true
		if m.Document.Extensions == nil {
			m.Document.Extensions = gltf.Extensions{}
		}
		m.Document.Extensions["KHR_lights_punctual"] = map[string]interface{}{"lights": lights}
	}

	textures := &textureCache{srcDir: textureDir, textures: map[string]*textureInfo{}}
	for i, mat := range doc.Materials {
		if _, ok := materialMap[i]; !ok {
			continue
		}
		mm := m.convertMaterial(mat, textures, uvBounds[i])
		m.Document.Materials = append(m.Document.Materials, mm)
	}
	if m.ConvertPhysics {
		m.extensions[BlenderPhysicsName] = true
	}
	for ext := range m.extensions {
		m.ExtensionsUsed = append(m.ExtensionsUsed, ext)
	}

	if len(m.Document.Textures) > 0 {
		m.Document.Samplers = []*gltf.Sampler{{}}
	}

	return m.Document, nil
}

func addPhysicsBody(node *gltf.Node, body *mqo.PhysicsBody, scale float32, nodePos *geom.Vector3) {
	var shapes []map[string]interface{}

	for _, s := range body.Shapes {
		shapes = append(shapes, map[string]interface{}{
			"boundingBox": [3]float32{s.Size.X * scale, s.Size.Y * scale, s.Size.Z * scale},
			"shapeType":   s.Type,
			"offsetTranslation": [3]float32{
				(s.Position.X - nodePos.X) * scale,
				(s.Position.Y - nodePos.Y) * scale,
				(s.Position.Z - nodePos.Z) * scale},
			"offsetScale": [3]float32{s.Size.X * scale, s.Size.Y * scale, s.Size.Z * scale},
			// "offsetRotation":    [4]float32{s.Rotation.X, s.Rotation.Y, s.Rotation.Z, 1}, // TODO
			// "primaryAxis": "Z",
		})
	}
	if len(shapes) > 0 {
		if node.Extensions == nil {
			node.Extensions = gltf.Extensions{}
		}
		if p, ok := node.Extensions[BlenderPhysicsName].(*BlenderPhysicsBody); ok {
			p.Shapes = append(p.Shapes, shapes...)
			p.Mass = p.Mass + body.Mass
			p.Static = p.Static || body.Kinematic
		} else {
			node.Extensions[BlenderPhysicsName] = &BlenderPhysicsBody{
				Shapes:          shapes,
				Mass:            body.Mass,
				Static:          body.Kinematic,
				CollisionGroups: body.CollisionGroup,
				CollisionMasks:  body.CollisionMask,
			}
		}
	}
}
