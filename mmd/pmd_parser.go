package mmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// PMDParser is parser for .pmd model.
type PMDParser struct {
	baseParser // TODO
	header     *Header
}

// NewPMDParser returns new parser.
func NewPMDParser(r io.Reader) *PMDParser {
	return &PMDParser{baseParser: baseParser{r}}
}

func (p *PMDParser) readString(len int) string {
	b := make([]byte, len)
	_ = p.read(b)
	utf8Data, _, _ := transform.Bytes(japanese.ShiftJIS.NewDecoder(), bytes.SplitN(b, []byte{0}, 2)[0])
	return string(utf8Data)
}

func (p *PMDParser) readHeader() error {
	h := p.header
	if h == nil {
		h = &Header{}
		h.Format = make([]byte, 3)
		p.read(&h.Format)
		p.header = h
	}
	if string(h.Format) != "Pmd" {
		return fmt.Errorf("Unsupported format")
	}
	p.read(&h.Version)
	return nil
}

func (p *PMDParser) readVertex() *Vertex {
	var v Vertex
	p.read(&v.Pos)
	p.read(&v.Normal)
	p.read(&v.UV)

	v.Bones = []int{p.readVInt(2), p.readVInt(2)}
	w := float32(p.readUint8()) / 100
	v.BoneWeights = []float32{w, 1 - w}
	v.EdgeScale = float32(p.readUint8())
	return &v
}

func (p *PMDParser) readMaterial(model *Document, i int) *Material {
	var m Material
	m.Name = fmt.Sprintf("mat%d", i+1)
	p.read(&m.Color)
	p.read(&m.Specularity)
	p.read(&m.Specular)
	p.read(&m.AColor)
	m.Toon = int(p.readUint8())
	m.EdgeScale = float32(p.readUint8())
	m.Count = p.readInt()

	tex := strings.SplitN(p.readString(20), "*", 2)
	if tex[0] != "" {
		m.TextureID = len(model.Textures)
		model.Textures = append(model.Textures, tex...)
	} else {
		m.TextureID = -1
	}

	if m.Color.W < 1 {
		m.Flags = MaterialFlagDoubleSided
	}
	return &m
}

func (p *PMDParser) readBone() *Bone {
	var b Bone
	b.Name = p.readString(20)
	b.ParentID = p.readVInt(2)
	b.TailID = p.readVInt(2)
	if b.TailID == 0 {
		b.TailID = -1
	}
	p.readUint8()
	p.readUint16()
	p.read(&b.Pos)
	return &b
}

func (p *PMDParser) readMorph() *Morph {
	var m Morph
	m.Name = p.readString(20)
	vn := p.readInt()
	p.read(&m.PanelType)
	for i := 0; i < vn; i++ {
		var mv MorphVertex
		mv.Target = p.readVInt(4)
		p.read(&mv.Offset)
		m.Vertex = append(m.Vertex, &mv)
	}
	return &m
}

// Parse model data.
func (p *PMDParser) Parse() (*Document, error) {
	var model Document

	if err := p.readHeader(); err != nil {
		return nil, err
	}
	model.Header = p.header
	model.Name = p.readString(20)
	model.Comment = p.readString(256)

	// Vertexes
	n := p.readInt()
	model.Vertexes = make([]*Vertex, n)
	for i := 0; i < n; i++ {
		model.Vertexes[i] = p.readVertex()
	}

	// Faces
	n = p.readInt()
	model.Faces = make([]*Face, n/3)
	for i := 0; i < n/3; i++ {
		var f Face
		f.Verts[0] = int(p.readUint16())
		f.Verts[1] = int(p.readUint16())
		f.Verts[2] = int(p.readUint16())
		model.Faces[i] = &f
	}

	// Materials
	mn := p.readInt()
	model.Materials = make([]*Material, mn)
	for i := 0; i < mn; i++ {
		model.Materials[i] = p.readMaterial(&model, i)
	}

	// Bones
	bn := int(p.readUint16())
	model.Bones = make([]*Bone, bn)
	for i := 0; i < bn; i++ {
		model.Bones[i] = p.readBone()
	}

	// IK
	n = int(p.readUint16())
	for i := 0; i < n; i++ {
		b := model.Bones[p.readUint16()]
		b.IK.TargetID = p.readVInt(2)
		ln := int(p.readUint8())
		b.IK.Loop = int(p.readUint16())
		p.readFloat()
		for i := 0; i < ln; i++ {
			b.IK.Links = append(b.IK.Links, &Link{TargetID: p.readVInt(2)})
		}
	}

	// Morph
	n = int(p.readUint16())
	if n > 0 {
		base := p.readMorph()
		model.Morphs = make([]*Morph, n-1)
		for i := 0; i < n-1; i++ {
			model.Morphs[i] = p.readMorph()
			for _, v := range model.Morphs[i].Vertex {
				v.Target = base.Vertex[v.Target].Target
			}
		}
	}

	return &model, nil
}
