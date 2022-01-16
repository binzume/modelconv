package converter

import (
	"github.com/binzume/modelconv/mmd"
	"github.com/binzume/modelconv/mqo"
)

type mqoToMMD struct {
	Scale float32
}

func NewMQOToMMDConverter(options interface{}) *mqoToMMD {
	return &mqoToMMD{Scale: 1.0 / 80}
}

func (c *mqoToMMD) Convert(doc *mqo.Document) (*mmd.Document, error) {
	dst := mmd.NewDocument()
	bones := mqo.GetBonePlugin(doc).Bones()
	boneIndexByID := map[int]int{}
	for i, b := range bones {
		boneIndexByID[b.ID] = i
		dst.Bones = append(dst.Bones, c.convertBone(b))
	}
	for i, b := range bones {
		if b.Parent > 0 {
			parentID := boneIndexByID[b.Parent]
			dst.Bones[i].ParentID = parentID
			dst.Bones[parentID].TailID = i
			dst.Bones[parentID].Flags = dst.Bones[parentID].Flags | mmd.BoneFlagTailIndex
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

	doc.FixObjectID()
	for mi, m := range doc.Materials {
		faceCount := 0
		for _, obj := range doc.Objects {
			if !obj.Visible || morphTargets[obj.Name] != nil {
				continue
			}
			type vertexKey struct {
				i  int
				uv mqo.Vector2
			}
			indicesMap := map[vertexKey]int{}
			vmap := map[int][]int{}
			normals := obj.GetSmoothNormals()

			for _, f := range obj.Faces {
				if len(f.Verts) < 3 || f.Material != mi {
					continue
				}
				verts := make([]int, len(f.Verts))
				for i, v := range f.Verts {
					var uv mqo.Vector2
					if len(f.UVs) > i {
						uv = f.UVs[i]
					}

					if id, ok := indicesMap[vertexKey{v, uv}]; ok {
						verts[i] = id
					} else {
						verts[i] = len(dst.Vertexes)
						indicesMap[vertexKey{v, uv}] = verts[i]
						vmap[v] = append(vmap[v], verts[i])
						vert := &mmd.Vertex{
							Pos:       *c.convertVec3(obj.Vertexes[v]),
							UV:        mmd.Vector2{X: uv.X, Y: uv.Y},
							Normal:    mmd.Vector3{X: normals[v].X, Y: normals[v].Y, Z: normals[v].Z},
							EdgeScale: 1,
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
				// convex polygon only. TODO: triangulation.
				for n := 0; n < len(verts)-2; n++ {
					dst.Faces = append(dst.Faces, &mmd.Face{Verts: [3]int{verts[0], verts[n+1], verts[n+2]}})
					faceCount++
				}
			}
			c.setWeights(dst, obj, vmap, bones)
		}
		texture := -1
		if m.Texture != "" {
			texture = len(dst.Textures)
			dst.Textures = append(dst.Textures, m.Texture)
		}
		dst.Materials = append(dst.Materials, c.convertMaterial(m, faceCount, texture))
	}

	// Physics
	physics := mqo.GetPhysicsPlugin(doc)
	for _, b := range physics.Bodies {
		dst.Bodies = append(dst.Bodies, c.convertBody(b))
	}

	return dst, nil
}

func (c *mqoToMMD) setWeights(dst *mmd.Document, obj *mqo.Object, vmap map[int][]int, bones []*mqo.Bone) {
	for bi, b := range bones {
		for _, bw := range b.Weights {
			if bw.ObjectID != obj.UID {
				continue
			}
			for _, vw := range bw.Vertexes {
				vi := obj.GetVertexIndexByID(vw.VertexID)
				for _, v := range vmap[vi] {
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
		Pos:      *c.convertVec3(&bone.Pos.Vector3),
		ParentID: -1,
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

func (c *mqoToMMD) convertBody(b *mqo.PhysicsBody) *mmd.RigidBody {
	var shape uint8 = 0
	switch b.Shape.Type {
	case "SPHERE":
		shape = 0
		break
	case "BOX":
		shape = 1
		break
	case "CAPSULE":
		shape = 2
		break
	}
	var mode uint8 = 0
	if b.Kinematic {
		mode = 1
	}
	return &mmd.RigidBody{
		Name:  b.Name,
		Shape: shape,

		Size:           mmd.Vector3(b.Shape.Size),
		Position:       mmd.Vector3(b.Shape.Position),
		Rotation:       mmd.Vector3(b.Shape.Rotation),
		Mass:           b.Mass,
		Mode:           mode,
		Group:          b.CollisionGroup,
		GroupTarget:    b.CollisionMask,
		LinearDamping:  b.LinearDamping,
		AngularDamping: b.AngularDamping,
		Restitution:    b.Restitution,
		Friction:       b.Friction,
		Bone:           b.TargetBoneID,
	}
}

func (c *mqoToMMD) convertVec3(v *mqo.Vector3) *mmd.Vector3 {
	return &mmd.Vector3{X: v.X * c.Scale, Y: v.Y * c.Scale, Z: v.Z * c.Scale * -1}
}
