package gltfutil

import (
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/geom"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/binary"
	"github.com/qmuntal/gltf/modeler"
)

func Load(path string) (*gltf.Document, error) {
	return gltf.Open(path)
}

func ToSingleFile(doc *gltf.Document, srcDir string) error {
	for _, b := range doc.Buffers {
		b.URI = ""
	}
	for _, m := range doc.Images {
		if m.BufferView == nil && m.URI != "" {
			path, _ := filepath.Rel(srcDir, m.URI)
			f, err := os.Open(path)
			if err != nil {
				log.Print(err)
				continue
			}
			defer f.Close()
			buf, err := ioutil.ReadAll(f)
			if err != nil {
				log.Print(err)
				continue
			}
			if m.MimeType == "" {
				if strings.HasSuffix(strings.ToLower(m.URI), ".png") {
					m.MimeType = "image/png"
				} else {
					m.MimeType = "image/jpeg"
				}
			}
			m.BufferView = gltf.Index(modeler.WriteBufferView(doc, gltf.TargetNone, buf))
			m.URI = ""
		}
	}
	return nil
}

func Transform(doc *gltf.Document, scale *geom.Vector3, offset *geom.Vector3) {
	if scale == nil && offset == nil {
		return
	}
	scaleMat := geom.NewMatrix4()
	if scale != nil {
		scaleMat = geom.NewScaleMatrix4(scale.X, scale.Y, scale.Z)
	}
	scaleOffsetMat := scaleMat
	if offset != nil {
		scaleOffsetMat = geom.NewTranslateMatrix4(offset.X, offset.Y, offset.Z).Mul(scaleMat)
	}

	accs := map[uint32]bool{}
	for _, m := range doc.Meshes {
		for _, p := range m.Primitives {
			if a, ok := p.Attributes["POSITION"]; ok {
				accs[a] = false
			}
			for _, t := range p.Targets {
				if a, ok := t["POSITION"]; ok {
					accs[a] = true
				}
			}
		}
	}
	for a, diff := range accs {
		acr := doc.Accessors[a]
		pos, err := modeler.ReadPosition(doc, acr, [][3]float32{})
		if err != nil {
			log.Fatalf("err %v", err)
			continue
		}
		if acr.Sparse != nil {
			log.Fatal("TODO: support sparsed accessor")
		}

		acr.Min = []float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
		acr.Max = []float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
		for i := range pos {
			if diff {
				scaleMat.ApplyTo(geom.NewVector3FromArray(pos[i])).ToArray(pos[i][:])
			} else {
				scaleOffsetMat.ApplyTo(geom.NewVector3FromArray(pos[i])).ToArray(pos[i][:])
			}
			for t, v := range pos[i] {
				acr.Min[t] = float32(math.Min(float64(acr.Min[t]), float64(v)))
				acr.Max[t] = float32(math.Max(float64(acr.Max[t]), float64(v)))
			}
		}
		bufferView := doc.BufferViews[*acr.BufferView]
		buffer := doc.Buffers[bufferView.Buffer]
		err = binary.Write(buffer.Data[bufferView.ByteOffset+acr.ByteOffset:], bufferView.ByteStride, pos)
		if err != nil {
			log.Fatalf("Write err %v", err)
		}
	}
	for _, node := range doc.Nodes {
		scaleMat.ApplyTo(geom.NewVector3FromArray(node.Translation)).ToArray(node.Translation[:])
	}
	for _, skin := range doc.Skins {
		if skin.InverseBindMatrices != nil {
			accessor := doc.Accessors[*skin.InverseBindMatrices]
			if accessor.BufferView != nil {
				bufferView := doc.BufferViews[*accessor.BufferView]
				// TODO: support sparse data.
				data := doc.Buffers[bufferView.Buffer].Data
				if len(data) == 0 {
					continue
				}
				for i := range skin.Joints {
					offset := bufferView.ByteOffset + uint32(i)*64
					mat := readMatrix(data[offset : offset+64])
					// apply scale
					geom.NewMatrix4FromSlice(mat[:]).Mul(scaleMat).ToArray(mat[:])
					// normalize rotation
					geom.NewVector3FromSlice(mat[0:3]).Normalize().ToArray(mat[0:3])
					geom.NewVector3FromSlice(mat[4:7]).Normalize().ToArray(mat[4:7])
					geom.NewVector3FromSlice(mat[8:11]).Normalize().ToArray(mat[8:11])
					writeMatrix(data[offset:offset+64], mat)
				}
			}
		}
	}
}
