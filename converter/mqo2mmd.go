package converter

import (
	"github.com/binzume/modelconv/mmd"
	"github.com/binzume/modelconv/mqo"
)

type mqoToMMD struct {
	Scale float32
}

func NewMQOToMMDConverter() *mqoToMMD {
	return &mqoToMMD{Scale: 1.0 / 80}
}

func (c *mqoToMMD) Convert(doc *mqo.MQODocument) (*mmd.Document, error) {
	dst := mmd.NewDocument()
	bones := mqo.GetBonePlugin(doc).Bones()
	for _, b := range bones {
		dst.Bones = append(dst.Bones, c.convertBone(b))
	}
	for i, b := range bones {
		if b.Parent > 0 {
			dst.Bones[b.Parent-1].TailID = i
			dst.Bones[b.Parent-1].Flags = dst.Bones[b.Parent-1].Flags | mmd.BoneFlagTailIndex
		}
	}

	objectByName := map[string]*mqo.Object{}
	morphTargets := map[string]*mmd.Morph{}
	morphBases := map[string]*mqo.MorphTargetList{}
	for _, obj := range doc.Objects {
		objectByName[obj.Name] = obj
	}

	morphs := mqo.GetMorphPlugin(doc).Morphs()
	for _, m := range morphs {
		morphBases[m.Base] = m
		for _, t := range m.Target {
			mmdMorph := &mmd.Morph{
				Name:      t.Name,
				MorphType: 1,
			}
			morphTargets[t.Name] = mmdMorph
			dst.Morphs = append(dst.Morphs, mmdMorph)
		}
	}

	for mi, m := range doc.Materials {
		faceCount := 0
		for i, obj := range doc.Objects {
			if !obj.Visible || morphTargets[obj.Name] != nil {
				continue
			}
			vmap := map[int]int{}
			for _, f := range obj.Faces {
				if len(f.Verts) < 3 || f.Material != mi {
					continue
				}
				var verts [3]int
				for i, v := range f.Verts[0:3] {
					if id, ok := vmap[v]; ok {
						// TODO: check UV
						verts[i] = id
					} else {
						verts[i] = len(dst.Vertexes)
						vmap[v] = verts[i]
						vert := &mmd.Vertex{
							Pos: *c.convertVec3(obj.Vertexes[v]),
						}
						if len(f.UVs) > i {
							vert.UV = mmd.Vector2{X: f.UVs[i].X, Y: f.UVs[i].Y}
						}
						dst.Vertexes = append(dst.Vertexes, vert)
						if morphBases[obj.Name] != nil {
							for _, t := range morphBases[obj.Name].Target {
								if target, ok := objectByName[t.Name]; ok {
									if v < len(target.Vertexes) && target.Vertexes[v] != obj.Vertexes[v] {
										p := c.convertVec3(target.Vertexes[v])
										morphTargets[t.Name].Vertex = append(morphTargets[t.Name].Vertex, &mmd.MorphVertex{
											Target: verts[i],
											Offset: mmd.Vector3{X: p.X - vert.Pos.X, Y: p.Y - vert.Pos.Y},
										})
									}
								}
							}
						}
					}
				}
				dst.Faces = append(dst.Faces, &mmd.Face{Verts: verts})
				faceCount++
			}
			objID := obj.UID
			if objID == 0 {
				objID = i + 1
			}
			c.setWeights(dst, objID, vmap, bones)
		}
		texture := -1
		if m.Texture != "" {
			texture = len(dst.Textures)
			dst.Textures = append(dst.Textures, m.Texture)
		}
		dst.Materials = append(dst.Materials, c.convertMaterial(m, faceCount, texture))
	}
	return dst, nil
}

func (c *mqoToMMD) setWeights(dst *mmd.Document, objID int, vmap map[int]int, bones []*mqo.Bone) {
	for bi, b := range bones {
		for _, bw := range b.Weights {
			if bw.ObjectID != objID {
				continue
			}
			for _, vw := range bw.Vertexes {
				if v, ok := vmap[vw.VertexID-1]; ok {
					vertex := dst.Vertexes[v]
					if len(vertex.BoneWeights) >= 4 {
						continue
					}
					vertex.Bones = append(vertex.Bones, bi)
					vertex.BoneWeights = append(vertex.BoneWeights, vw.Weight*0.01)
				}
			}
		}
	}
}

func (c *mqoToMMD) convertBone(bone *mqo.Bone) *mmd.Bone {
	return &mmd.Bone{
		Name:     bone.Name,
		Pos:      *c.convertVec3(&bone.Pos),
		ParentID: bone.Parent - 1,
		TailID:   -1,
		Flags:    mmd.BoneFlagVisible | mmd.BoneFlagEnabled | mmd.BoneFlagRotatable,
	}
}

func (c *mqoToMMD) convertMaterial(m *mqo.Material, faceCount, texture int) *mmd.Material {
	return &mmd.Material{
		Name:        m.Name,
		Color:       mmd.Vector4{X: m.Color.X, Y: m.Color.Y, Z: m.Color.Z, W: m.Color.W},
		Specularity: m.Power,
		TextureID:   texture,
		Toon:        -1,
		Count:       faceCount * 3,
	}
}

func (c *mqoToMMD) convertVec3(v *mqo.Vector3) *mmd.Vector3 {
	return &mmd.Vector3{X: v.X * c.Scale, Y: v.Y * c.Scale, Z: v.Z * c.Scale * -1}
}
