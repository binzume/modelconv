package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"./mmd"
	"./mqo"
)

func setFg(pmx *mmd.PMXDocument, f2fg []int, v2f [][]int, fi int, fgid int, fs []int) []int {
	f2fg[fi] = fgid
	fs = append(fs, fi)
	for _, vi := range pmx.Faces[fi].Verts {
		for _, f := range v2f[vi] {
			if f2fg[f] == 0 {
				fs = setFg(pmx, f2fg, v2f, f, fgid, fs)
			}
		}
	}
	return fs
}

func setMg(pmx *mmd.PMXDocument, fg2mg []int, m2fg, fg2m [][]int, fgi int, mgid int, fgs []int) []int {
	fg2mg[fgi] = mgid
	fgs = append(fgs, fgi)
	for _, m := range fg2m[fgi] {
		for _, fg := range m2fg[m] {
			if fg2mg[fg] == 0 {
				fgs = setMg(pmx, fg2mg, m2fg, fg2m, fg, mgid, fgs)
			}
		}
	}
	return fgs
}

func PMX2MQO(pmx *mmd.PMXDocument) *mqo.MQODocument {
	var mq mqo.MQODocument

	for boneIdx, pmBone := range pmx.Bones {
		if pmBone.TailID >= 0 {
			pmBone.TailPos = mmd.Vector3{
				pmx.Bones[pmBone.TailID].Pos.X - pmBone.Pos.X,
				pmx.Bones[pmBone.TailID].Pos.Y - pmBone.Pos.Y,
				pmx.Bones[pmBone.TailID].Pos.Z - pmBone.Pos.Z,
			}
		}
		mqBone := &mqo.Bone{
			ID:      boneIdx + 1,
			Name:    pmBone.Name,
			RtX:     pmBone.Pos.X,
			RtY:     pmBone.Pos.Y,
			RtZ:     pmBone.Pos.Z * -1,
			TpX:     pmBone.Pos.X + pmBone.TailPos.X,
			TpY:     pmBone.Pos.Y + pmBone.TailPos.Y,
			TpZ:     (pmBone.Pos.Z + pmBone.TailPos.Z) * -1,
			Sc:      1.0,
			MaxAngB: 90,
			MaxAngH: 180,
			MaxAngP: 180,
			MinAngB: -90,
			MinAngH: -180,
			MinAngP: -180,
		}
		mqBone.Parent.ID = pmBone.ParentID + 1
		mq.Bones = append(mq.Bones, mqBone)
	}
	mqo.UpdateBoneRef(&mq)

	v2f := make([][]int, len(pmx.Vertexes))
	for fid, f := range pmx.Faces {
		for _, vid := range f.Verts {
			v2f[vid] = append(v2f[vid], fid)
		}
	}
	f2fg := make([]int, len(pmx.Faces))
	var fgs [][]int
	for fid := range pmx.Faces {
		if f2fg[fid] == 0 {
			fgs = append(fgs, setFg(pmx, f2fg, v2f, fid, len(fgs)+1, []int{}))
		}
	}
	log.Println("face groups: ", len(fgs))

	m2fg := make([][]int, len(pmx.Morphs))
	for mi, m := range pmx.Morphs {
		fgs := map[int]bool{}
		for _, mv := range m.Vertex {
			fg := f2fg[v2f[mv.Target][0]]
			if !fgs[fg] {
				m2fg[mi] = append(m2fg[mi], fg)
				fgs[fg] = true
			}
		}

		for _, um := range m.UV {
			fg := f2fg[v2f[um.Target][0]]
			if !fgs[fg] {
				m2fg[mi] = append(m2fg[mi], fg)
				fgs[fg] = true
			}
		}
	}
	fg2m := make([][]int, len(fgs)+1)
	for mi, fgs := range m2fg {
		for _, fg := range fgs {
			fg2m[fg] = append(fg2m[fg], mi)
		}
	}

	fg2mg := make([]int, len(fgs)+1)
	mg2fg := [][]int{[]int{}}
	for fg := range fgs {
		if len(fg2m[fg]) > 0 && fg2mg[fg] == 0 {
			mg2fg = append(mg2fg, setMg(pmx, fg2mg, m2fg, fg2m, fg, len(mg2fg), []int{}))
		}
	}

	mg2fs := make([][]int, len(mg2fg))
	for fi := range pmx.Faces {
		mg := fg2mg[f2fg[fi]]
		mg2fs[mg] = append(mg2fs[mg], fi)
	}

	log.Println("morph groups: ", len(mg2fs))

	baseFaces := map[int]bool{}
	for _, f := range mg2fs[0] {
		baseFaces[f] = true
	}

	face2mat := make([]int, len(pmx.Faces))
	vpos := 0
	for matIdx, mat := range pmx.Materials {
		var m mqo.Material

		m.Name = mat.Name
		m.Color = mqo.Vector4{mat.Color.X, mat.Color.Y, mat.Color.Z, mat.Color.W}
		m.Specular = 0
		m.Diffuse = 1.0
		m.Ambient = 1.4
		m.Power = mat.Specularity
		if mat.TextureID >= 0 {
			m.Texture = pmx.Textures[mat.TextureID]
		}
		mq.Materials = append(mq.Materials, &m)

		for fi := vpos; fi < vpos+mat.Count/3; fi++ {
			face2mat[fi] = matIdx
		}
		o := mqo.NewObject("M_" + mat.Name)
		vmap := map[int]int{}
		for fi, face := range pmx.Faces[vpos : vpos+mat.Count/3] {
			if !baseFaces[vpos+fi] {
				continue
			}
			verts := make([]int, len(face.Verts))
			uvs := make([]mqo.Vector2, len(face.Verts))

			for i, vi := range face.Verts {
				v := pmx.Vertexes[vi]
				uvs[i] = mqo.Vector2{X: v.UV.X, Y: v.UV.Y}
				if mv, ok := vmap[vi]; ok {
					verts[i] = mv
				} else {
					vmap[vi] = len(o.Vertexes)
					verts[i] = vmap[vi]
					o.Vertexes = append(o.Vertexes, &mqo.Vector3{v.Pos.X, v.Pos.Y, v.Pos.Z * -1})
					if len(v.Bones) > 0 {
						for bi, b := range v.Bones {
							if v.BoneWeights[bi] > 0 {
								mq.Bones[b].Weights = append(mq.Bones[b].Weights, &mqo.BoneWeight{len(mq.Objects) + 1, vmap[vi] + 1, 100 * v.BoneWeights[bi]})
							}
						}
					}
				}
			}
			o.Faces = append(o.Faces, &mqo.Face{Verts: verts, Material: matIdx, UVs: uvs})
		}
		if len(o.Faces) > 0 || mat.Count == 0 {
			mq.Objects = append(mq.Objects, o)
		}
		vpos += mat.Count / 3
	}

	for mg, faces := range mg2fs {
		if mg == 0 {
			continue
		}
		o := mqo.NewObject(fmt.Sprintf("MorphBase%d", mg))
		var morphTargets mqo.MorphTargetList
		mq.Morphs = append(mq.Morphs, &morphTargets)
		morphTargets.Base = o.Name
		vmap := map[int]int{}
		o.Faces = make([]*mqo.Face, len(faces))
		for i, fi := range faces {
			face := pmx.Faces[fi]
			verts := make([]int, len(face.Verts))
			uvs := make([]mqo.Vector2, len(face.Verts))

			for i, vi := range face.Verts {
				v := pmx.Vertexes[vi]
				uvs[i] = mqo.Vector2{X: v.UV.X, Y: v.UV.Y}
				if mv, ok := vmap[vi]; ok {
					verts[i] = mv
				} else {
					vmap[vi] = len(o.Vertexes)
					verts[i] = vmap[vi]
					o.Vertexes = append(o.Vertexes, &mqo.Vector3{v.Pos.X, v.Pos.Y, v.Pos.Z * -1})
					if len(v.Bones) > 0 {
						for bi, b := range v.Bones {
							if v.BoneWeights[bi] > 0 {
								mq.Bones[b].Weights = append(mq.Bones[b].Weights, &mqo.BoneWeight{len(mq.Objects) + 1, vmap[vi] + 1, 100 * v.BoneWeights[bi]})
							}
						}
					}
				}
			}
			o.Faces[i] = &mqo.Face{Verts: verts, Material: face2mat[fi], UVs: uvs}
		}
		mq.Objects = append(mq.Objects, o)
		base := o

		mf := map[int]bool{}
		for _, fg := range mg2fg[mg] {
			for _, m := range fg2m[fg] {
				if mf[m] == false {
					mf[m] = true
					morph := pmx.Morphs[m]
					o := base.Clone()
					o.Name = morph.Name
					o.Depth = 1
					o.Visible = false
					morphTargets.Target = append(morphTargets.Target, &mqo.MorphTarget{Name: o.Name})
					for _, mv := range morph.Vertex {
						v := o.Vertexes[vmap[mv.Target]]
						v.X += mv.Offset.X
						v.Y += mv.Offset.Y
						v.Z += mv.Offset.Z * -1
					}
					for _, mu := range morph.UV {
						v := vmap[mu.Target]
						// TODO vert to faces
						for _, f := range o.Faces {
							for i, fv := range f.Verts {
								if fv == v {
									f.UVs[i].X += mu.Value.X
									f.UVs[i].Y += mu.Value.Y
								}
							}
						}
					}
					mq.Objects = append(mq.Objects, o)
				}
			}
		}

	}
	return &mq
}

func main() {

	r, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	if strings.HasSuffix(os.Args[1], ".mqo") {
		mq, _ := mqo.Parse(r, os.Args[1])
		w, _ := os.Create("out.mqo")
		defer w.Close()
		err = mqo.WriteMQO(mq, w, "")
		return
	}

	pmx, err := mmd.Parse(r)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Name", pmx.Name)
	log.Println("Comment", pmx.Comment)

	mqoFile := os.Args[1] + ".mqo"
	mqxFile := os.Args[1] + ".mqx"

	mq := PMX2MQO(pmx)
	w, _ := os.Create(mqoFile)
	defer w.Close()
	err = mqo.WriteMQO(mq, w, filepath.Base(mqxFile))
	if err != nil {
		log.Fatal(err)
	}

	mqxw, _ := os.Create(mqxFile)
	defer mqxw.Close()
	mqo.WriteMQX(mq, mqxw, filepath.Base(mqoFile))

	log.Println("ok")
}
