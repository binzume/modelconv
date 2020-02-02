package main

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
	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	_ "github.com/ftrvxmtrx/tga"
	_ "golang.org/x/image/bmp"
)

func addMatrices(m *modeler.Modeler, bufferIndex uint32, mat [][4][4]float32) uint32 {
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

func addGltfBones(m *modeler.Modeler, doc *mqo.MQODocument, scale float32) {
	bones := mqo.GetBonePlugin(doc).Bones()

	idmap := map[int]uint32{}
	idmapr := map[uint32]int{}
	bonemap := map[int]*mqo.Bone{}
	for _, b := range bones {
		idmap[b.ID] = uint32(len(m.Nodes))
		idmapr[uint32(len(m.Nodes))] = b.ID
		bonemap[b.ID] = b
		m.Nodes = append(m.Nodes, &gltf.Node{Name: b.Name, Translation: [3]float64{0, 0, 0}})
	}

	for _, b := range bonemap {
		if b.Parent > 0 {
			parent := bonemap[b.Parent]
			node := m.Nodes[idmap[b.ID]]
			node.Translation[0] = float64((b.Pos.X - parent.Pos.X) * scale)
			node.Translation[1] = float64((b.Pos.Y - parent.Pos.Y) * scale)
			node.Translation[2] = float64((b.Pos.Z - parent.Pos.Z) * scale)
			parentNode := m.Nodes[idmap[parent.ID]]
			parentNode.Children = append(parentNode.Children, idmap[b.ID])
		} else {
			m.Scenes[0].Nodes = append(m.Scenes[0].Nodes, idmap[b.ID])
		}
	}

	for _, skin := range m.Skins {
		if len(skin.Joints) == 0 {
			continue
		}
		invmats := make([][4][4]float32, len(skin.Joints))
		for i, j := range skin.Joints {
			b := bonemap[idmapr[j]]
			invmats[i] = [4][4]float32{
				{1, 0, 0, 0},
				{0, 1, 0, 0},
				{0, 0, 1, 0},
				{-b.Pos.X * scale, -b.Pos.Y * scale, -b.Pos.Z * scale, 1},
			}
		}
		skin.InverseBindMatrices = gltf.Index(addMatrices(m, 0, invmats))
	}
}

func getWeights(m *modeler.Modeler, doc *mqo.MQODocument, obj int, vs int) ([]uint32, [][4]uint16, [][4]float32) {
	joints := make([][4]uint16, vs)
	weights := make([][4]float32, vs)
	njoint := make([]int, vs)
	var jj []uint32
	bones := mqo.GetBonePlugin(doc).Bones()

	for _, b := range bones {
		for _, bw := range b.Weights {
			if bw.ObjectID != obj {
				continue
			}
			j := uint32(len(m.Nodes) + b.ID - 1) // TODO jointId
			jj = append(jj, j)
			for _, vw := range bw.Vertexes {
				v := vw.VertexID - 1
				if v < 0 || v >= vs || njoint[v] >= 4 {
					log.Fatal("invalid weight. V:", vw.VertexID, " O:", obj)
				}
				joints[v][njoint[v]] = uint16(len(jj)) - 1
				weights[v][njoint[v]] = vw.Weight * 0.01
				njoint[v]++
			}
		}
	}

	return jj, joints, weights
}

func mqo2gltf(doc *mqo.MQODocument, textureDir string) (*gltf.Document, error) {
	m := modeler.NewModeler()
	var scale float32 = 0.001

	textures := map[string]uint32{}

	var targetObjects []*mqo.Object
	var targetObjectIDs []int // TODO: Object.ID
	for i, obj := range doc.Objects {
		// TODO: remove Morph target.
		if obj.Visible && len(obj.Faces) > 0 {
			targetObjectIDs = append(targetObjectIDs, i+1)
			targetObjects = append(targetObjects, obj)
		}
	}

	m.Nodes = make([]*gltf.Node, len(targetObjects))

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

		positionAccessor := m.AddPosition(0, vertexes)
		texcoodAccessor := m.AddTextureCoord(0, texcood)
		attributes := map[string]uint32{
			"POSITION":   positionAccessor,
			"TEXCOORD_0": texcoodAccessor,
		}

		joints, j, w := getWeights(m, doc, targetObjectIDs[i], len(vertexes))
		if len(joints) > 0 {
			attributes["JOINTS_0"] = m.AddJoints(0, j)
			attributes["WEIGHTS_0"] = m.AddWeights(0, w)
		}

		// make primitive for each materials
		var primitives []*gltf.Primitive
		for mat, ind := range indices {
			indicesAccessor := m.AddIndices(0, ind)
			primitives = append(primitives, &gltf.Primitive{
				Indices:    gltf.Index(indicesAccessor),
				Attributes: attributes,
				Material:   gltf.Index(uint32(mat)),
			})
		}
		mesh := &gltf.Mesh{
			Name:       obj.Name,
			Primitives: primitives,
		}
		meshIndex := uint32(len(m.Document.Meshes))
		m.Document.Meshes = append(m.Document.Meshes, mesh)
		m.Nodes[i] = &gltf.Node{Name: obj.Name, Mesh: gltf.Index(meshIndex)}

		if len(joints) > 0 {
			m.Nodes[i].Skin = gltf.Index(uint32(len(m.Skins)))
			skin := &gltf.Skin{}
			skin.Joints = joints
			m.Skins = append(m.Skins, skin)
		}

		m.Scenes[0].Nodes = append(m.Scenes[0].Nodes, uint32(i))
	}

	for _, mat := range doc.Materials {
		mm := gltf.Material{
			Name: mat.Name,
			PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
				BaseColorFactor: &gltf.RGBA{R: float64(mat.Color.X), G: float64(mat.Color.Y), B: float64(mat.Color.Z), A: float64(mat.Color.W)},
			},
			DoubleSided: mat.DoubleSided,
		}
		if mat.Texture != "" {
			tex, exist := textures[mat.Texture]
			if !exist {
				f, err := os.Open(filepath.Join(textureDir, mat.Texture))
				if err != nil {
					log.Print("Texture file not found:", mat.Texture)
				}
				defer f.Close()
				var r io.Reader = f
				if !strings.HasSuffix(mat.Texture, ".png") {
					img, _, err := image.Decode(r)
					if err != nil {
						log.Fatal("Texture read error:", err, mat.Texture)
					}
					w := new(bytes.Buffer)
					err = png.Encode(w, img)
					if err != nil {
						log.Fatal("Texture encode error:", err, mat.Texture)
					}
					r = w
				}
				img, err := m.AddImage(0, filepath.Base(mat.Texture), "image/png", r)
				if err != nil {
					log.Fatal("Texture read error:", err, mat.Texture)
				}
				m.Document.Buffers[0].ByteLength = uint32(len(m.Document.Buffers[0].Data)) // avoid AddImage bug
				tex = uint32(len(m.Document.Textures))
				m.Document.Textures = append(m.Document.Textures,
					&gltf.Texture{Sampler: gltf.Index(0), Source: gltf.Index(img)})

				textures[mat.Texture] = tex
			}
			mm.PBRMetallicRoughness.BaseColorTexture = &gltf.TextureInfo{
				Index: tex,
			}
		}
		m.Document.Materials = append(m.Document.Materials, &mm)
	}

	if len(m.Document.Textures) > 0 {
		m.Document.Samplers = []*gltf.Sampler{
			{
				MagFilter: gltf.MagLinear,
				MinFilter: gltf.MinLinear,
				WrapS:     gltf.WrapRepeat,
				WrapT:     gltf.WrapRepeat,
			},
		}
	}

	addGltfBones(m, doc, scale)

	return m.Document, nil
}

func saveAsGlb(doc *mqo.MQODocument, path, textureDir string) error {
	gltfdoc, err := mqo2gltf(doc, textureDir)
	if err != nil {
		return err
	}
	return gltf.SaveBinary(gltfdoc, path)
}
