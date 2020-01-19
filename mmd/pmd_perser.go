package mmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

type PMDPerser struct {
	PmxPerser // TODO
}

func NewPMDParser(r io.Reader) *PMDPerser {
	return &PMDPerser{
		PmxPerser: *NewParser(r),
	}
}

func (p *PMDPerser) readString(len int) string {
	b := make([]byte, len)
	p.read(b)
	return string(bytes.Trim(b, "\x00"))
}

func (p *PMDPerser) readHeader() error {
	var h Header
	h.Format = make([]byte, 3)
	p.read(&h.Format)
	if string(h.Format) != "Pmd" {
		return fmt.Errorf("Unsupported format")
	}
	p.read(&h.Version)
	p.header = &h
	return nil
}

func (p *PMDPerser) readVertex() *Vertex {
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

	v.Bones = []int{p.readVInt(2), p.readVInt(2)}
	w := float32(p.readUint8()) / 100
	v.BoneWeights = []float32{w, 1 - w}
	if p.readUint8() > 0 {
		v.EdgeScale = 1.0
	}
	return &v
}

func (p *PMDPerser) readMaterial(pmx *PMXDocument, i int) *Material {
	var m Material
	m.Name = fmt.Sprintf("mat%d", i)
	p.read(&m.Color)
	p.read(&m.Specularity)
	p.read(&m.Specular)
	p.read(&m.AColor)
	m.Toon = int(p.readUint8())
	if p.readUint8() > 0 {
		m.EdgeScale = 1.0
	}
	m.Count = p.readInt()

	tex := p.readString(20)
	if tex != "" {
		m.TextureID = len(pmx.Textures)
		pmx.Textures = append(pmx.Textures, tex)
	}
	return &m
}

func (p *PMDPerser) readBone() *Bone {
	var b Bone
	b.Name = p.readString(20)
	b.ParentID = p.readVInt(2)
	b.TailID = p.readVInt(2)
	p.readUint8()
	p.readUint16()
	p.read(&b.Pos)
	return &b
}

func (p *PMDPerser) Parse() (*PMXDocument, error) {
	var pmx PMXDocument

	if err := p.readHeader(); err != nil {
		return nil, err
	}
	pmx.Header = p.header
	pmx.Name = p.readString(20)
	p.readString(256)

	vn := p.readInt()
	pmx.Vertexes = make([]*Vertex, vn)
	for i := 0; i < vn; i++ {
		pmx.Vertexes[i] = p.readVertex()
	}

	in := p.readInt()
	pmx.Faces = make([]*Face, in/3)
	for i := 0; i < in/3; i++ {
		var f Face
		f.Verts[0] = int(p.readUint16())
		f.Verts[1] = int(p.readUint16())
		f.Verts[2] = int(p.readUint16())
		pmx.Faces[i] = &f
	}

	mn := p.readInt()
	pmx.Materials = make([]*Material, mn)
	for i := 0; i < mn; i++ {
		pmx.Materials[i] = p.readMaterial(&pmx, i)
	}

	bn := int(p.readUint16())
	pmx.Bones = make([]*Bone, bn)
	for i := 0; i < bn; i++ {
		pmx.Bones[i] = p.readBone()
	}

	n := int(p.readUint16())
	for i := 0; i < n; i++ {
		p.readUint16()
		p.readUint16()
		ln := int(p.readUint8())
		p.readUint16()
		p.readFloat()
		for i := 0; i < ln; i++ {
			p.readUint16()
		}
	}

	fn := int(p.readUint16())
	log.Println("face morph:", fn)

	return &pmx, nil
}
