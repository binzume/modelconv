package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/mmd"
	"github.com/binzume/modelconv/mqo"
)

func convertVec3(v *mmd.Vector3) *mqo.Vector3 {
	return &mqo.Vector3{X: v.X, Y: v.Y, Z: v.Z * -1}
}

func convertMaterial(mat *mmd.Material, model *mmd.PMXDocument) *mqo.Material {
	var m mqo.Material
	m.Name = mat.Name
	m.Color = mqo.Vector4{X: mat.Color.X, Y: mat.Color.Y, Z: mat.Color.Z, W: mat.Color.W}
	m.Specular = 0
	m.Diffuse = 1.0
	m.Ambient = 1.4
	m.Power = mat.Specularity
	m.DoubleSided = mat.Flags&mmd.MaterialFlagDoubleSided != 0
	if mat.TextureID >= 0 {
		m.Texture = model.Textures[mat.TextureID]
	}

	m.Ex2 = &mqo.MaterialEx2{
		ShaderType: "hlsl",
		ShaderName: "pmd",
		ShaderParams: map[string]interface{}{
			"Edge": mat.EdgeScale > 0,
		},
	}
	return &m
}

func convertBones(pmx *mmd.PMXDocument) []*mqo.Bone {
	var bones []*mqo.Bone

	for boneIdx, pmBone := range pmx.Bones {
		mqBone := &mqo.Bone{
			ID:     boneIdx + 1,
			Name:   pmBone.Name,
			Group:  pmBone.Layer,
			Pos:    *convertVec3(&pmBone.Pos),
			Parent: pmBone.ParentID + 1,
		}
		if pmBone.Flags&mmd.BoneFlagTranslatable != 0 {
			mqBone.Movable = 1
		}
		bones = append(bones, mqBone)
	}

	for _, pmBone := range pmx.Bones {
		if len(pmBone.IK.Links) > 0 {
			bones[pmBone.IK.TargetID].IK = &mqo.BoneIK{ChainCount: len(pmBone.IK.Links) + 1}
		}
	}
	return bones
}

func convertFaces(pmx *mmd.PMXDocument, faces []int, face2mat []int, o *mqo.Object, vmap map[int]int) {
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
				o.Vertexes = append(o.Vertexes, convertVec3(&v.Pos))
			}
		}
		o.Faces[i] = &mqo.Face{Verts: verts, Material: face2mat[fi], UVs: uvs}
	}
}

func setWeight(pmx *mmd.PMXDocument, bones []*mqo.Bone, objid int, vmap map[int]int) {
	for pmv, mqv := range vmap {
		v := pmx.Vertexes[pmv]
		c := map[int]*mqo.VertexWeight{}
		for bi, b := range v.Bones {
			if v.BoneWeights[bi] > 0 {
				if c[b] != nil {
					c[b].Weight += v.BoneWeights[bi]
					continue
				}
				c[b] = bones[b].SetVertexWeight(objid, mqv+1, 100*v.BoneWeights[bi])
			}
		}
	}
}

func newFg(pmx *mmd.PMXDocument, f2fg []int, v2f [][]int, fi int, fgid int, fs []int) []int {
	f2fg[fi] = fgid
	fs = append(fs, fi)
	for _, vi := range pmx.Faces[fi].Verts {
		for _, f := range v2f[vi] {
			if f2fg[f] == 0 {
				fs = newFg(pmx, f2fg, v2f, f, fgid, fs)
			}
		}
	}
	return fs
}

func newMg(pmx *mmd.PMXDocument, m2mg []int, m2fg, fg2m [][]int, mi int, mgid int, ms []int) []int {
	m2mg[mi] = mgid
	ms = append(ms, mi)
	for _, fg := range m2fg[mi] {
		for _, m := range fg2m[fg] {
			if m2mg[m] == 0 {
				ms = newMg(pmx, m2mg, m2fg, fg2m, m, mgid, ms)
			}
		}
	}
	return ms
}

func genMorphGroup(pmx *mmd.PMXDocument) ([][]int, [][]int) {
	// TODO: more better impl.
	v2f := make([][]int, len(pmx.Vertexes))
	for fid, f := range pmx.Faces {
		for _, vid := range f.Verts {
			v2f[vid] = append(v2f[vid], fid)
		}
	}
	f2fg := make([]int, len(pmx.Faces))
	fgs := [][]int{[]int{}}
	for fid := range pmx.Faces {
		if f2fg[fid] == 0 {
			fgs = append(fgs, newFg(pmx, f2fg, v2f, fid, len(fgs), []int{}))
		}
	}
	log.Println("face groups: ", len(fgs))

	m2fg := make([][]int, len(pmx.Morphs))
	for mi, m := range pmx.Morphs {
		fgs := map[int]bool{}
		for _, mv := range m.Vertex {
			if len(v2f[mv.Target]) == 0 {
				continue
			}
			fg := f2fg[v2f[mv.Target][0]]
			if !fgs[fg] {
				m2fg[mi] = append(m2fg[mi], fg)
				fgs[fg] = true
			}
		}
		for _, um := range m.UV {
			if len(v2f[um.Target]) == 0 {
				continue
			}
			fg := f2fg[v2f[um.Target][0]]
			if !fgs[fg] {
				m2fg[mi] = append(m2fg[mi], fg)
				fgs[fg] = true
			}
		}
	}
	fg2m := make([][]int, len(fgs))
	for mi, fgs := range m2fg {
		for _, fg := range fgs {
			fg2m[fg] = append(fg2m[fg], mi)
		}
	}

	m2mg := make([]int, len(m2fg))
	mg2m := [][]int{[]int{}}
	for mi, fg := range m2fg {
		if len(fg) > 0 && m2mg[mi] == 0 {
			mg2m = append(mg2m, newMg(pmx, m2mg, m2fg, fg2m, mi, len(mg2m), []int{}))
		}
	}

	fg2mg := make([]int, len(fgs))
	for mg, ms := range mg2m {
		for _, mi := range ms {
			for _, fg := range m2fg[mi] {
				fg2mg[fg] = mg
			}
		}
	}
	mg2fs := make([][]int, len(mg2m))
	for fi := range pmx.Faces {
		mg := fg2mg[f2fg[fi]]
		mg2fs[mg] = append(mg2fs[mg], fi)
	}
	return mg2m, mg2fs
}

func mmd2mqo(pmx *mmd.PMXDocument) *mqo.MQODocument {
	mq := mqo.NewDocument()

	bones := convertBones(pmx)
	mqo.GetBonePlugin(mq).SetBones(bones)

	mg2m, mg2fs := genMorphGroup(pmx)

	baseFaces := map[int]bool{}
	for _, f := range mg2fs[0] {
		baseFaces[f] = true
	}

	face2mat := make([]int, len(pmx.Faces))
	vpos := 0
	for matIdx, mat := range pmx.Materials {
		m := convertMaterial(mat, pmx)
		mq.Materials = append(mq.Materials, m)

		for fi := vpos; fi < vpos+mat.Count/3; fi++ {
			face2mat[fi] = matIdx
		}

		o := mqo.NewObject("M_" + mat.Name)
		faces := []int{}
		for fi := vpos; fi < vpos+mat.Count/3; fi++ {
			if baseFaces[fi] {
				faces = append(faces, fi)
			}
		}
		vpos += mat.Count / 3

		vmap := map[int]int{}
		convertFaces(pmx, faces, face2mat, o, vmap)
		if len(o.Faces) == 0 && mat.Count != 0 {
			continue
		}
		setWeight(pmx, bones, len(mq.Objects)+1, vmap)
		mq.Objects = append(mq.Objects, o)
	}

	morphPlugin := mqo.GetMorphPlugin(mq)
	for mg, faces := range mg2fs {
		if mg == 0 {
			continue
		}
		o := mqo.NewObject(fmt.Sprintf("MorphBase%d", mg))
		var morphTargets mqo.MorphTargetList
		morphPlugin.MorphSet.Targets = append(morphPlugin.MorphSet.Targets, &morphTargets)
		morphTargets.Base = o.Name
		vmap := map[int]int{}
		convertFaces(pmx, faces, face2mat, o, vmap)
		for pmv, mqv := range vmap {
			v := pmx.Vertexes[pmv]
			c := map[int]*mqo.VertexWeight{}
			for bi, b := range v.Bones {
				if v.BoneWeights[bi] > 0 {
					if c[b] != nil {
						c[b].Weight += v.BoneWeights[bi]
						continue
					}
					c[b] = bones[b].SetVertexWeight(len(mq.Objects)+1, mqv+1, 100*v.BoneWeights[bi])
				}
			}
		}

		mq.Objects = append(mq.Objects, o)
		base := o

		for _, m := range mg2m[mg] {
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
				if mu.Target < 0 {
					continue
				}
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
	return mq
}

func loadDocument(input string) (*mqo.MQODocument, error) {
	r, err := os.Open(input)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	if strings.ToLower(filepath.Ext(input)) == ".mqo" {
		return mqo.Parse(r, input)
	}

	pmx, err := mmd.Parse(r)
	if err != nil {
		return nil, err
	}
	log.Println("Name: ", pmx.Name)
	log.Println("Comment: ", pmx.Comment)
	return mmd2mqo(pmx), nil
}

func saveDocument(doc *mqo.MQODocument, output string) error {
	w, err := os.Create(output)
	if err != nil {
		return err
	}
	defer w.Close()
	return mqo.WriteMQO(doc, w, output)
}
