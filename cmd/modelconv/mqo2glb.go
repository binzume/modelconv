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

func mqo2gltf(doc *mqo.MQODocument, textureDir string) (*gltf.Document, error) {
	m := modeler.NewModeler()

	textures := map[string]uint32{}

	for _, obj := range doc.Objects {
		if !obj.Visible || len(obj.Faces) == 0 {
			continue
		}
		var vertexes [][3]float32
		for _, v := range obj.Vertexes {
			vertexes = append(vertexes, [3]float32{v.X * 0.001, v.Y * 0.001, v.Z * 0.001})
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

		// make primitive for each materials
		var primitives []*gltf.Primitive
		for mat, ind := range indices {
			indicesAccessor := m.AddIndices(0, ind)
			primitives = append(primitives, &gltf.Primitive{
				Indices: gltf.Index(indicesAccessor),
				Attributes: map[string]uint32{
					"POSITION":   positionAccessor,
					"TEXCOORD_0": texcoodAccessor,
				},
				Material: gltf.Index(uint32(mat)),
			})
		}
		mesh := &gltf.Mesh{
			Name:       obj.Name,
			Primitives: primitives,
		}
		meshIndex := uint32(len(m.Document.Meshes))
		m.Document.Meshes = append(m.Document.Meshes, mesh)
		nodeIndex := uint32(len(m.Nodes))
		m.Nodes = append(m.Nodes, &gltf.Node{Name: obj.Name, Mesh: gltf.Index(meshIndex)})
		m.Scenes[0].Nodes = append(m.Scenes[0].Nodes, nodeIndex)
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

	return m.Document, nil
}

func saveAsGlb(doc *mqo.MQODocument, path, textureDir string) error {
	gltfdoc, err := mqo2gltf(doc, textureDir)
	if err != nil {
		return err
	}
	return gltf.SaveBinary(gltfdoc, path)
}
