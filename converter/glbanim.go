package converter

import (
	"log"
	"math"

	"github.com/binzume/modelconv/mmd"
	"github.com/binzume/modelconv/mqo"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

func keysEquals(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isDefaultRotations(samples []*mmd.Vector4) bool {
	ident := mmd.Vector4{X: 0, Y: 0, Z: 0, W: 1}
	for _, q := range samples {
		if *q != ident {
			return false
		}
	}
	return true
}

func addBoneChannels(doc *gltf.Document, a *gltf.Animation, bones map[uint32]*mqo.Bone, bb map[string]*mmd.BoneChannel) {
	var scale float32 = 80 * 0.001

	boneToNode := map[string]int{}
	for n, b := range bones {
		boneToNode[b.Name] = int(n)
	}

	boneByID := map[int]*mqo.Bone{}
	for _, b := range bones {
		boneByID[b.ID] = b
	}

	ikNodes := map[uint32]uint32{}
	for ni, b := range bones {
		if b.IK != nil && boneToNode[b.IK.Name] > 0 {
			ikNodes[uint32(boneToNode[b.IK.Name])] = ni
		}
	}

	var prevCh *mmd.BoneChannel
	var prevKeysAcc uint32
	for _, channel := range bb {
		if n, ok := boneToNode[channel.Target]; ok {
			var keysAcc uint32
			if prevCh != nil && keysEquals(channel.Frames, prevCh.Frames) {
				keysAcc = prevKeysAcc
			} else {
				var keys []float32
				for _, k := range channel.Frames {
					keys = append(keys, float32(k)/30)
				}
				keysAcc = modeler.WriteAccessor(doc, gltf.TargetArrayBuffer, keys)
			}

			var rot mqo.Vector3
			var pos mqo.Vector3
			qOffset := mqo.Vector4{X: 0, Y: 0, Z: 0, W: 1}
			if b, ok := bones[uint32(n)]; ok {
				rot = b.RotationOffset
				pos = b.Pos.Vector3

				if boneByID[b.Parent] != nil && b.RotationOffset.X != boneByID[b.Parent].RotationOffset.X {
					// TODO rotation matrix
					d := (boneByID[b.Parent].RotationOffset.X - b.RotationOffset.X) * float32(math.Cos(float64(boneByID[b.Parent].RotationOffset.Y)))
					qOffset = mqo.Vector4{X: 0, Y: 0, Z: float32(math.Sin(float64(d / 2))), W: float32(math.Cos(float64(d / 2)))}
					log.Println("rotation offset", qOffset)
				}
			}

			// TODO rotation matrix
			cosRotY := float32(math.Cos(float64(rot.Y)))
			sinRotY := float32(math.Sin(float64(rot.Y)))
			cosRotX := float32(math.Cos(float64(rot.X)))
			sinRotX := float32(math.Sin(float64(rot.X)))

			rotate := false
			var rotations [][4]float32
			for _, ms := range channel.Rotations {
				s := (&mqo.Vector4{X: ms.X, Y: ms.Y, Z: ms.Z, W: ms.W}).Mul(&qOffset)
				r := [4]float32{-s.X, -s.Y, s.Z, s.W}

				r = [4]float32{
					r[0]*cosRotY - r[2]*sinRotY,
					r[1],
					r[0]*sinRotY + r[2]*cosRotY,
					r[3],
				}

				r = [4]float32{
					r[0]*cosRotX - r[1]*sinRotX,
					r[0]*sinRotX + r[1]*cosRotX,
					r[2],
					r[3],
				}

				if r != [4]float32{0, 0, 0, 1} {
					rotate = true
				}

				rotations = append(rotations, r)
			}
			if rotate {
				log.Println("Rotation channel:", channel.Target)
				samplesAcc := modeler.WriteTangent(doc, rotations)
				a.Samplers = append(a.Samplers, &gltf.AnimationSampler{
					Input:         gltf.Index(uint32(keysAcc)),
					Output:        gltf.Index(uint32(samplesAcc)),
					Interpolation: gltf.InterpolationLinear,
				})

				a.Channels = append(a.Channels, &gltf.Channel{
					Sampler: gltf.Index(uint32(len(a.Samplers) - 1)),
					Target: gltf.ChannelTarget{
						Node: gltf.Index(uint32(n)),
						Path: gltf.TRSRotation,
					},
				})
			}

			translate := false
			var translations [][3]float32
			for _, s := range channel.Positions {
				r := [3]float32{s.X * scale, s.Y * scale, -s.Z * scale}

				r = [3]float32{
					r[0]*cosRotY - r[2]*sinRotY,
					r[1],
					r[0]*sinRotY + r[2]*cosRotY,
				}

				r = [3]float32{
					r[0]*cosRotX - r[1]*sinRotX,
					r[0]*sinRotX + r[1]*cosRotX,
					r[2],
				}
				if *s != (mmd.Vector3{X: 0, Y: 0, Z: 0}) {
					translate = true
				}
				translations = append(translations, [3]float32{pos.X*0.001 + r[0], pos.Y*0.001 + r[1], pos.Z*0.001 + r[2]})
			}
			if translate {
				log.Println("Translate channel:", channel.Target)
				samplesAcc := modeler.WritePosition(doc, translations)
				a.Samplers = append(a.Samplers, &gltf.AnimationSampler{
					Input:         gltf.Index(uint32(keysAcc)),
					Output:        gltf.Index(uint32(samplesAcc)),
					Interpolation: gltf.InterpolationLinear,
				})

				a.Channels = append(a.Channels, &gltf.Channel{
					Sampler: gltf.Index(uint32(len(a.Samplers) - 1)),
					Target: gltf.ChannelTarget{
						Node: gltf.Index(uint32(n)),
						Path: gltf.TRSTranslation,
					},
				})
			}

			if b, ok := ikNodes[uint32(n)]; ok {
				log.Println("TODO IK:", b, translate, bones[b].IK.ChainCount)
			}

			prevCh = channel
			prevKeysAcc = keysAcc
		}
	}
}

func addMorphChannels(doc *gltf.Document, a *gltf.Animation, morphs map[string]*mmd.MorphChannel) {
	targets := map[string][3]int{} // targetName => [node, targetIndex]
	for i, n := range doc.Nodes {
		if n.Mesh == nil {
			continue
		}
		if len(doc.Meshes[*n.Mesh].Primitives) == 0 {
			continue
		}

		if extras, ok := doc.Meshes[*n.Mesh].Extras.(map[string]interface{}); ok {
			if names, ok := extras["targetNames"].([]string); ok {
				for ti, name := range names {
					// TODO multiple mesh
					targets[name] = [3]int{i, ti, len(names)}
				}
			}
		}
	}

	nodeChannels := map[int][]*mmd.MorphChannel{}
	for _, m := range morphs {
		if _, exits := targets[m.Target]; !exits {
			continue
		}
		nodeChannels[targets[m.Target][0]] = append(nodeChannels[targets[m.Target][0]], m)
	}

	for ni, mm := range nodeChannels {
		var keys []float32
		for _, k := range mm[0].Frames {
			keys = append(keys, float32(k)/30)
		}

		sampleCount := len(keys)
		sampleSize := targets[mm[0].Target][2]
		weights := make([]uint8, sampleSize*sampleCount)

		for _, m := range mm {
			if len(m.Frames) != sampleCount {
				log.Print("TODO: resample", m.Target, len(m.Frames), sampleCount)
				continue
			}
			for f, v := range m.Weights {
				weights[f*sampleSize+targets[m.Target][1]] = uint8(v * 255)
			}
		}
		log.Print(doc.Nodes[ni], weights)

		keysAcc := modeler.WriteAccessor(doc, gltf.TargetArrayBuffer, keys)
		weightsAcc := modeler.WriteAccessor(doc, gltf.TargetArrayBuffer, weights)
		doc.Accessors[weightsAcc].Normalized = true

		a.Samplers = append(a.Samplers, &gltf.AnimationSampler{
			Input:         gltf.Index(uint32(keysAcc)),
			Output:        gltf.Index(uint32(weightsAcc)),
			Interpolation: gltf.InterpolationLinear,
		})

		a.Channels = append(a.Channels, &gltf.Channel{
			Sampler: gltf.Index(uint32(len(a.Samplers) - 1)),
			Target: gltf.ChannelTarget{
				Node: gltf.Index(uint32(ni)),
				Path: gltf.TRSWeights,
			},
		})
	}
}

func AddAnimationTpGlb(doc *gltf.Document, anim *mmd.Animation, bones map[uint32]*mqo.Bone, compact bool) {
	a := gltf.Animation{Name: anim.Name}

	addBoneChannels(doc, &a, bones, anim.GetBoneChannels())
	addMorphChannels(doc, &a, anim.GetMorphChannels())

	if len(a.Channels) > 0 {
		doc.Animations = append(doc.Animations, &a)
	}
}
