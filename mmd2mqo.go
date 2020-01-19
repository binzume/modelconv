package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"./mmd"
	"./mqo"
)

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

		var o mqo.Object
		o.Name = mat.Name
		vmap := map[int]int{}
		for _, face := range pmx.Faces[vpos : vpos+mat.Count/3] {
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

		mq.Objects = append(mq.Objects, &o)
		vpos += mat.Count / 3
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
