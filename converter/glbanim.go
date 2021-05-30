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
	for i, _ := range a {
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

func AddAnimationTpGlb(doc *gltf.Document, anim *mmd.Animation, bones map[uint32]*mqo.Bone, compact bool) {

	mod := modeler.NewModeler()
	mod.Document = doc

	boneToNode := map[string]int{}
	if bones != nil {
		for n, b := range bones {
			boneToNode[b.Name] = int(n)
		}
	} else {
		bones = map[uint32]*mqo.Bone{}
		for i, node := range doc.Nodes {
			// TODO: check hierarchy
			if node.Name != "" && node.Mesh == nil && node.Skin == nil && node.Camera == nil {
				if _, exists := boneToNode[node.Name]; exists {
					log.Println("dup", node.Name)
					continue
				}
				boneToNode[node.Name] = i
			}
		}
	}

	boneByID := map[int]*mqo.Bone{}
	for _, b := range bones {
		boneByID[b.ID] = b
	}

	a := gltf.Animation{Name: anim.Name}
	var prevCh *mmd.RotationChannel
	var prevKeyAcc uint32
	for name, channel := range anim.GetRotationChannels() {
		if n, ok := boneToNode[name]; ok {
			if compact && isDefaultRotations(channel.Samples) {
				log.Println("Animation skip:", name, len(channel.Samples))
				continue
			}
			log.Println("Animation target:", name)
			var keysAcc uint32
			if prevCh != nil && keysEquals(channel.Frames, prevCh.Frames) {
				keysAcc = prevKeyAcc
			} else {
				var keys []float32
				for _, k := range channel.Frames {
					keys = append(keys, float32(k)/30)
				}
				// TODO use keys
				// keysAcc = mod.AddFloatArray(0, keys)
				keysAcc = mod.AddIndices(0, channel.Frames)
			}

			var rot mqo.Vector3
			qOffset := mqo.Vector4{X: 0, Y: 0, Z: 0, W: 1}
			if b, ok := bones[uint32(n)]; ok {
				rot = b.RotationOffset

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

			var rotations [][4]float32
			for _, ms := range channel.Samples {
				s := mqo.MultQQ(&mqo.Vector4{X: ms.X, Y: ms.Y, Z: ms.Z, W: ms.W}, &qOffset)
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

				rotations = append(rotations, r)
			}
			samplesAcc := mod.AddTangent(0, rotations)

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

			prevCh = channel
			prevKeyAcc = keysAcc
		}
	}

	doc.Animations = append(doc.Animations, &a)
}
