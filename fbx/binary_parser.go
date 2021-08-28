package fbx

import (
	"bufio"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

type positionReader struct {
	r        io.Reader
	position int64
}

func (r *positionReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.position += int64(n)
	return n, err
}

func (r *positionReader) SkipTo(pos int64) error {
	offset := pos - r.position
	if offset < 0 {
		return fmt.Errorf("cannot rewind")
	}
	r.position = pos
	if s, ok := r.r.(io.Seeker); ok {
		_, err := s.Seek(pos, 0)
		return err
	}
	_, err := io.CopyN(ioutil.Discard, r, offset)
	return err
}

type binaryParser struct {
	r       *positionReader
	version uint32
	err     error
}

func (p *binaryParser) read(v interface{}) interface{} {
	if p.err == nil {
		p.err = binary.Read(p.r, binary.LittleEndian, v)
	}
	return v
}

func (p *binaryParser) readUint8() uint8 {
	var v uint8
	p.read(&v)
	return v
}

func (p *binaryParser) readUint16() uint16 {
	var v uint16
	p.read(&v)
	return v
}

func (p *binaryParser) readUint64() uint64 {
	var v uint64
	p.read(&v)
	return v
}

func (p *binaryParser) readInt() int {
	var v uint32
	p.read(&v)
	return int(v)
}

func (p *binaryParser) readUint32() uint32 {
	var v uint32
	p.read(&v)
	return v
}

func (p *binaryParser) readFloat() float32 {
	var v float32
	p.read(&v)
	return v
}

func (p *binaryParser) readFloat64() float64 {
	var v float64
	p.read(&v)
	return v
}

func (p *binaryParser) readString(len uint) string {
	bytes := make([]byte, len)
	p.read(bytes)
	return string(bytes)
}

func (p *binaryParser) readName() string {
	return p.readString(uint(p.readUint8()))
}

func (p *binaryParser) readPropArray(typ uint8) *Attribute {
	count := uint(p.readUint32())
	encoding := p.readUint32()
	sz := p.readUint32()
	var buf interface{}
	switch typ {
	case 'b':
		buf = make([]int8, count)
	case 'y':
		buf = make([]int16, count)
	case 'i':
		buf = make([]int32, count)
	case 'l':
		buf = make([]int64, count)
	case 'f':
		buf = make([]float32, count)
	case 'd':
		buf = make([]float64, count)
	default:
		return nil
	}
	if encoding == 0 {
		p.read(buf)
	} else {
		next := p.r.position + int64(sz)
		r, err := zlib.NewReader(io.LimitReader(p.r, int64(sz)))
		if err != nil {
			p.err = err
			return &Attribute{buf, count}
		}
		defer r.Close()
		err = binary.Read(r, binary.LittleEndian, buf)
		if p.err == nil {
			p.err = err
		}
		p.r.SkipTo(next)
	}
	return &Attribute{buf, count}
}

func (p *binaryParser) readProp() *Attribute {
	typ := p.readUint8()

	switch typ {
	case 'B':
		var v int8
		p.read(&v)
		return &Attribute{v, 0}
	case 'C':
		var v byte
		p.read(&v)
		return &Attribute{v, 0}
	case 'Y':
		var v int16
		p.read(&v)
		return &Attribute{v, 0}
	case 'I':
		var v int32
		p.read(&v)
		return &Attribute{v, 0}
	case 'L':
		var v int64
		p.read(&v)
		return &Attribute{v, 0}
	case 'F':
		return &Attribute{p.readFloat(), 0}
	case 'D':
		return &Attribute{p.readFloat64(), 0}
	case 'S':
		return &Attribute{p.readString(uint(p.readUint32())), 0}
	case 'R':
		buf := make([]byte, p.readUint32())
		p.read(buf)
		return &Attribute{buf, 0}
	case 'b', 'y', 'i', 'l', 'f', 'd':
		return p.readPropArray(typ)
	}
	p.err = fmt.Errorf("unknown prop type: %v", typ)
	return nil
}

func (p *binaryParser) readNode() *Node {
	n := &Node{}
	var next int64
	var nprop int64
	var propsz int64
	if p.version >= 7500 {
		p.read(&next)
		p.read(&nprop)
		p.read(&propsz)
	} else {
		next = int64(p.readUint32())
		nprop = int64(p.readUint32())
		propsz = int64(p.readUint32())
	}
	n.Name = p.readName()

	if uint64(nprop)*2 > uint64(propsz) {
		// invalid node?
		p.err = p.r.SkipTo(int64(next))
		return nil
	}

	if next == 0 {
		return nil
	}

	for i := int64(0); i < nprop && p.err == nil; i++ {
		n.Attributes = append(n.Attributes, p.readProp())
		if p.err != nil {
			log.Println(n, p.err)

		}
	}
	if p.err != nil && p.err != io.EOF {
		return nil
	}

	for p.r.position < next && p.err == nil {
		child := p.readNode()
		if child != nil {
			n.AddChild(child)
		}
	}

	if p.err == nil {
		p.err = p.r.SkipTo(next)
	}
	if p.err != nil && p.err != io.EOF {
		return nil
	}
	return n
}

func (p *binaryParser) Parse() (*Node, error) {
	magic := p.readString(21)
	if magic != "Kaydara FBX Binary  \x00" {
		// try parse texy. TODO
		if magic[0] == ';' {
			p2 := &textParser{r: bufio.NewReader(p.r.r), buf: []byte(magic)}
			return p2.Parse()
		}
		return nil, fmt.Errorf("unknown fbx format")
	}
	p.readUint16()
	p.read(&p.version)
	root := &Node{Name: "_FBX_ROOT"}

	for p.err == nil {
		node := p.readNode()
		if node != nil {
			root.AddChild(node)
		}
	}
	if p.err != nil && p.err != io.EOF {
		return nil, p.err
	}
	return root, nil
}
