package mmd

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"unicode/utf16"
)

// PMXParser is parser for .pmx model.
type PMXParser struct {
	baseParser
	header *Header
}

// NewPMXParser returns new parser.
func NewPMXParser(r io.Reader) *PMXParser {
	return &PMXParser{
		baseParser: baseParser{r: r},
	}
}

func (p *PMXParser) readIndex(attrTyp int) int {
	return p.readVInt(p.header.Info[attrTyp])
}

func (p *PMXParser) readUIndex(attrTyp int) int {
	return p.readVUInt(p.header.Info[attrTyp])
}

func (p *PMXParser) readText() string {
	len := p.readInt()

	if p.header.Info[AttrStringEncoding] == 0 {
		utf16data := make([]uint16, len/2)
		binary.Read(p.r, binary.LittleEndian, &utf16data)
		return string(utf16.Decode(utf16data))
	}
	data := make([]byte, len)
	binary.Read(p.r, binary.LittleEndian, &data)
	return string(data)
}

func (p *PMXParser) readHeader() error {
	var h = p.header
	if h == nil {
		h = &Header{}
		h.Format = make([]byte, 4)
		p.read(&h.Format)
		p.header = h
	}
	if string(h.Format) != "PMX " {
		return fmt.Errorf("Unsupported file")
	}
	p.read(&h.Version)
	h.Info = make([]byte, p.readUint8())
	p.read(&h.Info)
	return nil
}

func (p *PMXParser) readVertex() *Vertex {
	var v Vertex
	if err := p.read(&v.Pos); err != nil {
		log.Fatal(err)
	}
	if err := p.read(&v.Normal); err != nil {
		log.Fatal(err)
	}
	if err := p.read(&v.UV); err != nil {
		log.Fatal(err)
	}
	v.ExtUVs = make([]Vector4, p.header.Info[AttrExtUV])
	if err := p.read(&v.ExtUVs); err != nil {
		log.Fatal(err)
	}
	wehghtType := p.readUint8()
	if wehghtType == 0 {
		v.Bones = []int{p.readIndex(AttrBoneIndexSz)}
		v.BoneWeights = []float32{1}
	} else if wehghtType == 1 {
		v.Bones = []int{p.readIndex(AttrBoneIndexSz), p.readIndex(AttrBoneIndexSz)}
		w := p.readFloat()
		v.BoneWeights = []float32{w, 1 - w}
	} else if wehghtType == 2 {
		v.Bones = []int{
			p.readIndex(AttrBoneIndexSz),
			p.readIndex(AttrBoneIndexSz),
			p.readIndex(AttrBoneIndexSz),
			p.readIndex(AttrBoneIndexSz),
		}
		v.BoneWeights = []float32{
			p.readFloat(),
			p.readFloat(),
			p.readFloat(),
			p.readFloat(),
		}
	} else if wehghtType == 3 {
		v.Bones = []int{p.readIndex(AttrBoneIndexSz), p.readIndex(AttrBoneIndexSz)}
		w := p.readFloat()
		weights := [11]float32{w, 1 - w}
		v.BoneWeights = weights[:]
		p.read(weights[2:])
	} else {
		log.Println(v)
		log.Fatal("unknown weight ", wehghtType)
	}
	v.EdgeScale = p.readFloat()
	return &v
}

func (p *PMXParser) readFace() *Face {
	var f Face
	f.Verts[0] = p.readUIndex(AttrVertIndexSz)
	f.Verts[1] = p.readUIndex(AttrVertIndexSz)
	f.Verts[2] = p.readUIndex(AttrVertIndexSz)
	return &f
}

func (p *PMXParser) readMaterial() *Material {
	var m Material
	m.Name = p.readText()
	m.NameEn = p.readText()
	p.read(&m.Color)
	p.read(&m.Specular)
	p.read(&m.Specularity)
	p.read(&m.AColor)
	p.read(&m.Flags)
	p.read(&m.EdgeColor)
	p.read(&m.EdgeScale)
	m.TextureID = p.readIndex(AttrTexIndexSz)
	m.EnvID = p.readIndex(AttrTexIndexSz)
	p.read(&m.EnvMode)
	p.read(&m.ToonType)
	if m.ToonType == 0 {
		m.Toon = p.readIndex(AttrTexIndexSz)
	} else {
		m.Toon = int(p.readUint8())
	}
	m.Memo = p.readText()
	m.Count = p.readInt()
	return &m
}

func (p *PMXParser) readBone() *Bone {
	var b Bone
	b.Name = p.readText()
	b.NameEn = p.readText()
	p.read(&b.Pos)
	b.ParentID = p.readIndex(AttrBoneIndexSz)
	b.Layer = p.readInt()
	p.read(&b.Flags)

	// TODO
	if b.Flags & ^BoneFlagAll != 0 {
		log.Println("Unsupported flags : ", b.Flags & ^BoneFlagAll)
	}

	if b.Flags&BoneFlagTailIndex != 0 {
		b.TailID = p.readIndex(AttrBoneIndexSz)
	} else {
		b.TailID = -1
		p.read(&b.TailPos)
	}

	if b.Flags&256 != 0 || b.Flags&512 != 0 {
		b.InheritParentID = p.readIndex(AttrBoneIndexSz)
		b.InheritParentInfluence = p.readFloat()
	}

	if b.Flags&1024 != 0 {
		p.read(&b.FixedAxis)
	}

	if b.Flags&2048 != 0 {
		// local
		var dummy Vector3
		p.read(&dummy)
		p.read(&dummy)
	}

	if b.Flags&8192 != 0 {
		// ??
		p.readIndex(AttrBoneIndexSz)
	}

	if b.Flags&32 != 0 {
		b.IK.TargetID = p.readIndex(AttrBoneIndexSz)
		b.IK.Loop = p.readInt()
		b.IK.LimitRad = p.readFloat()
		links := p.readInt()
		for i := 0; i < links; i++ {
			var l Link
			l.TargetID = p.readIndex(AttrBoneIndexSz)
			l.HasLimit = p.readUint8() != 0
			if l.HasLimit {
				p.read(&l.LimitMin)
				p.read(&l.LimitMax)
			}
			b.IK.Links = append(b.IK.Links, &l)
		}
	}

	return &b
}

func (p *PMXParser) readMorph() *Morph {
	var m Morph
	m.Name = p.readText()
	m.NameEn = p.readText()
	p.read(&m.PanelType)
	p.read(&m.MorphType)

	n := p.readInt()
	for i := 0; i < n; i++ {
		switch m.MorphType {
		case 0:
			m.Group = append(m.Group, &MorphGroup{
				Target: p.readIndex(AttrMorphIndexSz),
				Weight: p.readFloat(),
			})
			break
		case 1:
			var v MorphVertex
			v.Target = p.readUIndex(AttrVertIndexSz)
			p.read(&v.Offset)
			m.Vertex = append(m.Vertex, &v)
			break
		case 2:
			p.readIndex(AttrBoneIndexSz)
			p.read(&Vector3{})
			p.read(&Vector4{})
			break
		case 3:
			var v MorphUV
			v.Target = p.readUIndex(AttrVertIndexSz)
			p.read(&v.Value)
			m.UV = append(m.UV, &v)
			break
		case 8:
			var v MorphMaterial
			v.Target = p.readIndex(AttrMatIndexSz)
			p.read(&v.Flags)
			p.read(&v.Diffuse)
			p.read(&v.Specular)
			p.read(&v.Specularity)
			p.read(&v.Ambient)
			p.read(&v.EdgeColor)
			p.read(&v.EdgeSize)
			p.read(&v.TextureTint)
			p.read(&v.EnvironmentTint)
			p.read(&v.ToonTint)
			m.Material = append(m.Material, &v)
			break
		default:
			log.Fatal("Unknown morph type ", m.MorphType)
			break
		}
	}

	return &m
}

// Parse model data.
func (p *PMXParser) Parse() (*Document, error) {
	var pmx Document

	if err := p.readHeader(); err != nil {
		return nil, err
	}

	pmx.Header = p.header
	pmx.Name = p.readText()
	pmx.NameEn = p.readText()
	pmx.Comment = p.readText()
	pmx.CommentEn = p.readText()

	vn := p.readInt()
	pmx.Vertexes = make([]*Vertex, vn)
	for i := 0; i < vn; i++ {
		pmx.Vertexes[i] = p.readVertex()
	}

	fn := p.readInt() / 3
	pmx.Faces = make([]*Face, fn)
	for i := 0; i < fn; i++ {
		pmx.Faces[i] = p.readFace()
	}

	tn := p.readInt()
	pmx.Textures = make([]string, tn)
	for i := 0; i < tn; i++ {
		pmx.Textures[i] = p.readText()
	}

	mn := p.readInt()
	pmx.Materials = make([]*Material, mn)
	for i := 0; i < mn; i++ {
		pmx.Materials[i] = p.readMaterial()
	}

	bn := p.readInt()
	pmx.Bones = make([]*Bone, bn)
	for i := 0; i < bn; i++ {
		pmx.Bones[i] = p.readBone()
	}

	pn := p.readInt()
	pmx.Morphs = make([]*Morph, pn)
	for i := 0; i < pn; i++ {
		pmx.Morphs[i] = p.readMorph()
	}

	return &pmx, nil
}

// Parse pmx/pmd file.
func Parse(r io.Reader) (*Document, error) {
	// check format
	format := make([]byte, 4)
	if _, err := r.Read(format[:3]); err != nil {
		return nil, err
	}

	if string(format[:3]) == "Pmd" {
		p := NewPMDParser(bufio.NewReader(r))
		p.header = &Header{Format: format[:3]}
		return p.Parse()
	}
	r.Read(format[3:])
	p := NewPMXParser(bufio.NewReader(r))
	p.header = &Header{Format: format}
	return p.Parse()
}
