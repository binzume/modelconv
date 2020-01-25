package mmd

import (
	"encoding/binary"
	"io"
	"log"
)

type baseWriter struct {
	w io.Writer
}

func (p *baseWriter) write(v interface{}) error {
	return binary.Write(p.w, binary.LittleEndian, v)
}

func (p *baseWriter) writeUint8(v uint8) uint8 {
	binary.Write(p.w, binary.LittleEndian, &v)
	return v
}

func (p *baseWriter) writeUint16(v uint16) uint16 {
	binary.Write(p.w, binary.LittleEndian, &v)
	return v
}

func (p *baseWriter) writeInt(v int) {
	vv := int32(v)
	binary.Write(p.w, binary.LittleEndian, &vv)
}

func (p *baseWriter) writeFloat(v float32) float32 {
	binary.Write(p.w, binary.LittleEndian, &v)
	return v
}

func (p *baseWriter) writeVUInt(sz byte, vv int) int {
	if sz == 1 {
		var v = uint8(vv)
		binary.Write(p.w, binary.LittleEndian, &v)
		return int(v)
	}
	if sz == 2 {
		var v = uint16(vv)
		binary.Write(p.w, binary.LittleEndian, &v)
		return int(v)
	}
	if sz == 4 {
		var v = uint32(vv)
		binary.Write(p.w, binary.LittleEndian, &v)
		return int(v)
	}
	return 0
}

func (p *baseWriter) writeVInt(sz byte, vv int) int {
	if sz == 1 {
		var v = int8(vv)
		binary.Write(p.w, binary.LittleEndian, &v)
		return int(v)
	}
	if sz == 2 {
		var v = int16(vv)
		binary.Write(p.w, binary.LittleEndian, &v)
		return int(v)
	}
	if sz == 4 {
		var v = int32(vv)
		binary.Write(p.w, binary.LittleEndian, &v)
		return int(v)
	}
	return 0
}

// PMXWriter is writer for .pmx data
type PMXWriter struct {
	baseWriter
	header *Header
}

func (w *PMXWriter) Write(doc *PMXDocument) error {
	// header
	w.writeHeader(doc.Header)
	w.writeText(doc.Name)
	w.writeText(doc.NameEn)
	w.writeText(doc.Comment)
	w.writeText(doc.CommentEn)

	// vertexes
	w.writeInt(len(doc.Vertexes))
	for _, v := range doc.Vertexes {
		w.writeVertex(v)
	}

	// faces
	w.writeInt(len(doc.Faces) * 3)
	for _, f := range doc.Faces {
		w.writeFace(f)
	}

	// textures
	w.writeInt(len(doc.Textures))
	for _, t := range doc.Textures {
		w.writeText(t)
	}

	// materials
	w.writeInt(len(doc.Materials))
	for _, m := range doc.Materials {
		w.writeMaterial(m)
	}

	// bones
	w.writeInt(len(doc.Bones))
	for _, b := range doc.Bones {
		w.writeBone(b)
	}

	// morphs
	w.writeInt(len(doc.Morphs))
	for _, m := range doc.Morphs {
		w.writeMorph(m)
	}

	// TODO
	w.writeInt(0)
	w.writeInt(0)
	w.writeInt(0)
	w.writeInt(0)

	return nil
}

func (w *PMXWriter) writeText(v string) {
	w.writeInt(len(v))
	binary.Write(w.w, binary.LittleEndian, []byte(v))
}

func (w *PMXWriter) writeIndex(attrTyp int, v int) {
	w.writeVInt(w.header.Info[attrTyp], v)
}

func (w *PMXWriter) writeUIndex(attrTyp int, v int) {
	w.writeVUInt(w.header.Info[attrTyp], v)
}

func (w *PMXWriter) writeHeader(h *Header) {
	w.header = h

	// TODO
	h.Format = []byte("PMX ")
	h.Info[AttrStringEncoding] = 1

	w.write(&h.Format)
	w.write(&h.Version)
	w.writeUint8(uint8(len(h.Info)))
	w.write(&h.Info)
}

func (w *PMXWriter) writeVertex(v *Vertex) {
	w.write(&v.Pos)
	w.write(&v.Normal)
	w.write(&v.UV)
	w.write(&v.ExtUVs)

	var wehghtType uint8
	switch len(v.BoneWeights) {
	case 1:
		wehghtType = 0
		break
	case 2:
		wehghtType = 1
		break
	case 4:
		wehghtType = 2
		break
	case 11:
		wehghtType = 3
		break
	}
	if len(v.BoneWeights) == 1 {
		wehghtType = 0
	} else if len(v.BoneWeights) == 2 {
		wehghtType = 1
	}

	w.writeUint8(wehghtType)
	for _, b := range v.Bones {
		w.writeIndex(AttrBoneIndexSz, b)
	}
	if wehghtType == 0 {
	} else if wehghtType == 1 {
		w.writeFloat(v.BoneWeights[0])
	} else if wehghtType == 3 {
		w.writeFloat(v.BoneWeights[0])
		w.write(v.BoneWeights[2:])
	} else {
		w.write(v.BoneWeights)
	}
	w.write(&v.EdgeScale)
}

func (w *PMXWriter) writeFace(f *Face) {
	w.writeIndex(AttrVertIndexSz, f.Verts[0])
	w.writeIndex(AttrVertIndexSz, f.Verts[1])
	w.writeIndex(AttrVertIndexSz, f.Verts[2])
}

func (w *PMXWriter) writeMaterial(m *Material) {
	w.writeText(m.Name)
	w.writeText(m.NameEn)
	w.write(&m.Color)
	w.write(&m.Specular)
	w.write(&m.Specularity)
	w.write(&m.AColor)
	w.write(&m.Flags)
	w.write(&m.EdgeColor)
	w.write(&m.EdgeScale)

	w.writeIndex(AttrTexIndexSz, m.TextureID)
	w.writeIndex(AttrTexIndexSz, m.EnvID)

	w.write(&m.EnvMode)
	w.write(&m.ToonType)
	if m.ToonType == 0 {
		w.writeIndex(AttrTexIndexSz, m.Toon)
	} else {
		w.writeUint8(uint8(m.Toon))
	}

	w.writeText(m.Memo)
	w.writeInt(m.Count)
}

func (w *PMXWriter) writeBone(b *Bone) {
	w.writeText(b.Name)
	w.writeText(b.NameEn)
	w.write(&b.Pos)

	w.writeIndex(AttrBoneIndexSz, b.ParentID)
	w.writeInt(b.Layer)

	w.write(&b.Flags)

	if b.Flags & ^BoneFlagAll != 0 {
		log.Println("Unsupported flags : ", b.Flags & ^BoneFlagAll)
	}

	if b.Flags&BoneFlagTailIndex != 0 {
		w.writeIndex(AttrBoneIndexSz, b.TailID)
	} else {
		w.write(&b.TailPos)
	}

	if b.Flags&256 != 0 || b.Flags&512 != 0 {
		w.writeIndex(AttrBoneIndexSz, b.InheritParentID)
		w.write(&b.InheritParentInfluence)
	}

	if b.Flags&1024 != 0 {
		w.write(&b.FixedAxis)
	}

	if b.Flags&2048 != 0 {
		// local
		var dummy Vector3
		w.write(&dummy)
		w.write(&dummy)
	}

	if b.Flags&8192 != 0 {
		// ??
		w.writeIndex(AttrBoneIndexSz, -1)
	}

	if b.Flags&32 != 0 {
		w.writeIndex(AttrBoneIndexSz, b.IK.TargetID)
		w.writeInt(b.IK.Loop)
		w.write(&b.IK.LimitRad)
		w.writeInt(len(b.IK.Links))
		for _, l := range b.IK.Links {
			w.writeIndex(AttrBoneIndexSz, l.TargetID)
			if l.HasLimit {
				w.writeUint8(1)
				w.write(&l.LimitMin)
				w.write(&l.LimitMax)
			} else {
				w.writeUint8(0)
			}
		}
	}
}

func (w *PMXWriter) writeMorph(m *Morph) {
	w.writeText(m.Name)
	w.writeText(m.NameEn)
	w.write(&m.PanelType)
	w.write(&m.MorphType)

	// oneof
	w.writeInt(len(m.Group) + len(m.Vertex) + len(m.UV) + len(m.Material))

	for _, m := range m.Group {
		w.writeIndex(AttrMorphIndexSz, m.Target)
		w.write(&m.Weight)
	}
	for _, m := range m.Vertex {
		w.writeIndex(AttrVertIndexSz, m.Target)
		w.write(&m.Offset)
	}
	for _, m := range m.UV {
		w.writeIndex(AttrVertIndexSz, m.Target)
		w.write(&m.Value)
	}
	for _, m := range m.Material {
		w.writeIndex(AttrMatIndexSz, m.Target)
		w.write(&m.Flags)
		w.write(&m.Diffuse)
		w.write(&m.Specular)
		w.write(&m.Specularity)
		w.write(&m.Ambient)
		w.write(&m.EdgeColor)
		w.write(&m.EdgeSize)
		w.write(&m.TextureTint)
		w.write(&m.EnvironmentTint)
		w.write(&m.ToonTint)
	}
}

// WritePMX writes .pmx data
func WritePMX(doc *PMXDocument, w io.Writer) error {
	return (&PMXWriter{baseWriter: baseWriter{w}}).Write(doc)
}
