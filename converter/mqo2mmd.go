package converter

import (
	"github.com/binzume/modelconv/mmd"
	"github.com/binzume/modelconv/mqo"
)

type mqoToMMD struct {
}

func NewMQOToMMDConverter() *mqoToMMD {
	return &mqoToMMD{}
}

func (c *mqoToMMD) Convert(doc *mqo.MQODocument) (*mmd.Document, error) {
	dst := mmd.NewDocument()
	for mi, m := range doc.Materials {
		vmap := map[int]int{}
		faceCount := 0
		for _, obj := range doc.Objects {
			for _, f := range obj.Faces {
				if len(f.Verts) < 3 || f.Material != mi {
					continue
				}
				var verts [3]int
				for i, v := range f.Verts[0:3] {
					if id, ok := vmap[v]; ok {
						verts[i] = id
					} else {
						verts[i] = len(dst.Vertexes)
						vmap[v] = verts[i]
						vert := &mmd.Vertex{
							Pos:         *c.convertVec3(obj.Vertexes[v]),
							Bones:       []int{-1},
							BoneWeights: []float32{0},
						}
						if len(f.UVs) > i {
							vert.UV = mmd.Vector2{X: f.UVs[i].X, Y: f.UVs[i].Y}
						}
						dst.Vertexes = append(dst.Vertexes, vert)
					}
				}
				dst.Faces = append(dst.Faces, &mmd.Face{Verts: verts})
				faceCount++
			}
		}
		texture := -1
		if m.Texture != "" {
			texture = len(dst.Textures)
			dst.Textures = append(dst.Textures, m.Texture)
		}

		dst.Materials = append(dst.Materials, &mmd.Material{
			Name:        m.Name,
			Color:       mmd.Vector4{X: m.Color.X, Y: m.Color.Y, Z: m.Color.Z, W: m.Color.W},
			Specularity: m.Power,
			TextureID:   texture,
			Count:       faceCount * 3,
		})
	}
	return dst, nil
}

func (c *mqoToMMD) convertVec3(v *mqo.Vector3) *mmd.Vector3 {
	return &mmd.Vector3{X: v.X, Y: v.Y, Z: v.Z * -1}
}
