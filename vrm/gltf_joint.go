package vrm

import (
	"encoding/binary"
	"math"

	"github.com/qmuntal/gltf"
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

func ResetJointMatrix(doc *gltf.Document) {
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
			if node.Rotation == [4]float32{0, 0, 0, 1} || node.Skin != nil {
				continue
			}
			fixed++
			a := Quaternion{X: node.Rotation[0], Y: node.Rotation[1], Z: node.Rotation[2], W: node.Rotation[3]}
			node.Rotation = [4]float32{0, 0, 0, 1}
			for _, c := range node.Children {
				child := doc.Nodes[c]

				pos := child.Translation
				p := Quaternion{X: pos[0], Y: pos[1], Z: pos[2], W: 1}
				p2 := a.Mul(&p).Mul(a.Inverse()) // p2 = a * p * ~a

				child.Translation[0] = p2.X
				child.Translation[1] = p2.Y
				child.Translation[2] = p2.Z

				rot := child.Rotation
				b := Quaternion{X: rot[0], Y: rot[1], Z: rot[2], W: rot[3]}
				q := a.Mul(&b)
				child.Rotation[0] = q.X
				child.Rotation[1] = q.Y
				child.Rotation[2] = q.Z
				child.Rotation[3] = q.W
			}
		}
		if fixed == 0 {
			break
		}
	}
}

func FixJointComponentType(doc *gltf.Document) {
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
