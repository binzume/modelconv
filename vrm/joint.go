package vrm

import (
	"encoding/binary"
	"math"
)

func readMatrix(data []byte) [16]float32 {
	var mat [16]float32
	for i := 0; i < 16; i++ {
		d := binary.LittleEndian.Uint32(data[i*4 : i*4+4])
		mat[i] = math.Float32frombits(d)
	}
	return mat
}

func writeMatrix(data []byte, mat [16]float32) {
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(data[i*4:i*4+4], math.Float32bits(mat[i]))
	}
}

type Q struct {
	x, y, z, w float64
}

func (q *Q) c() *Q {
	return &Q{-q.x, -q.y, -q.z, q.w}
}

func mul(a, b *Q) *Q {
	var r Q
	r.w = a.w*b.w - a.x*b.x - a.y*b.y - a.z*b.z // w
	r.x = a.w*b.x + a.x*b.w + a.y*b.z - a.z*b.y // i
	r.y = a.w*b.y - a.x*b.z + a.y*b.w + a.z*b.x // j
	r.z = a.w*b.z + a.x*b.y - a.y*b.x + a.z*b.w // k
	return &r
}

func (doc *VRMDocument) FixJointMatrix() {
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
					// fix inverseBindMatrix
					offset := bufferView.ByteOffset + uint32(i)*64
					mat := readMatrix(data[offset : offset+64])

					x := mat[0]*mat[12] + mat[1]*mat[13] + mat[2]*mat[14]
					y := mat[4]*mat[12] + mat[5]*mat[13] + mat[6]*mat[14]
					z := mat[8]*mat[12] + mat[9]*mat[13] + mat[10]*mat[14]
					writeMatrix(data[offset:offset+64], [16]float32{
						1, 0, 0, 0,
						0, 1, 0, 0,
						0, 0, 1, 0,
						x, y, z, 1,
					})
				}
			}
		}
	}

	// TODO: check node dependency
	for i := 0; i < 10; i++ {
		fixed := 0
		for _, node := range doc.Nodes {
			if node.Rotation == [4]float64{0, 0, 0, 1} || node.Skin != nil {
				continue
			}
			fixed++
			a := Q{node.Rotation[0], node.Rotation[1], node.Rotation[2], node.Rotation[3]}
			node.Rotation = [4]float64{0, 0, 0, 1}
			for _, c := range node.Children {
				child := doc.Nodes[c]
				rb := child.Rotation
				b := Q{rb[0], rb[1], rb[2], rb[3]}

				pos := child.Translation
				p := Q{pos[0], pos[1], pos[2], 1}
				p2 := mul(&a, mul(&p, a.c()))

				child.Translation[0] = p2.x
				child.Translation[1] = p2.y
				child.Translation[2] = p2.z

				q := mul(&a, &b)
				child.Rotation[0] = q.x
				child.Rotation[1] = q.y
				child.Rotation[2] = q.z
				child.Rotation[3] = q.w
			}
		}
		if fixed == 0 {
			break
		}
	}
}

func (doc *VRMDocument) FixJointComponentType() {
	fixedbuffer := map[uint32]bool{}
	for _, mesh := range doc.Meshes {
		for _, primitiv := range mesh.Primitives {
			for k, attr := range primitiv.Attributes {
				accessor := doc.Accessors[attr]
				if k == "JOINTS_0" && doc.Accessors[attr].ComponentType == 2 { // byte
					bufferView := doc.BufferViews[*accessor.BufferView]
					buffer := doc.Buffers[bufferView.Buffer]

					accessor.ComponentType = 4
					if fixedbuffer[*accessor.BufferView] {
						continue
					}
					fixedbuffer[*accessor.BufferView] = true

					// byte to uint16
					src := buffer.Data[bufferView.ByteOffset : bufferView.ByteOffset+bufferView.ByteLength]
					dst := make([]byte, bufferView.ByteLength*2)
					for i, b := range src {
						binary.LittleEndian.PutUint16(dst[i*2:i*2+2], uint16(b))
					}

					bufferView.ByteLength *= 2
					bufferView.ByteStride *= 2
					bufferView.ByteOffset = uint32(len(buffer.Data))

					buffer.Data = append(buffer.Data, dst...)
					buffer.ByteLength += uint32(len(dst))
				}
			}
		}
	}
}
