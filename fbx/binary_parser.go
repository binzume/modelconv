package fbx

import (
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
	r   *positionReader
	err error
}

func (p *binaryParser) read(v interface{}) error {
	if p.err == nil {
		p.err = binary.Read(p.r, binary.LittleEndian, v)
	}
	return p.err
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

func (p *binaryParser) readPropArray(typ uint8) *Property {
	count := uint(p.readUint32())
	encoding := p.readUint32()
	sz := p.readUint32()
	var buf interface{}
	switch typ {
	case 'b':
		buf = make([]byte, count)
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
			return &Property{typ, buf, count}
		}
		defer r.Close()
		err = binary.Read(r, binary.LittleEndian, buf)
		if p.err == nil {
			p.err = err
		}
		p.r.SkipTo(next)
	}
	return &Property{typ, buf, count}
}

func (p *binaryParser) readProp() *Property {
	typ := p.readUint8()

	switch typ {
	case 'B':
		return &Property{typ, p.readUint8(), 0}
	case 'C':
		return &Property{typ, p.readUint8(), 0}
	case 'Y':
		return &Property{typ, p.readUint16(), 0}
	case 'I':
		return &Property{typ, p.readUint32(), 0}
	case 'L':
		return &Property{typ, p.readUint64(), 0}
	case 'F':
		return &Property{typ, p.readFloat(), 0}
	case 'D':
		return &Property{typ, p.readFloat64(), 0}
	case 'S':
		return &Property{typ, p.readString(uint(p.readUint32())), 0}
	case 'R':
		buf := make([]byte, p.readUint32())
		p.read(buf)
		return &Property{typ, buf, 0}
	case 'b', 'i', 'l', 'f', 'd':
		return p.readPropArray(typ)
	}
	p.err = fmt.Errorf("unknown prop type: %v", typ)
	return nil
}

func (p *binaryParser) readNode() *Node {
	n := &Node{}
	next := p.readUint32()
	nprop := p.readInt()
	propsz := p.readUint32()
	n.Name = p.readName()

	if uint64(nprop)*2 > uint64(propsz) {
		// invalid node?
		p.err = p.r.SkipTo(int64(next))
		return nil
	}

	if next == 0 {
		return nil
	}

	for i := 0; i < nprop && p.err == nil; i++ {
		n.Properties = append(n.Properties, p.readProp())
		if p.err != nil {
			log.Println(n, p.err)

		}
	}
	if p.err != nil && p.err != io.EOF {
		return nil
	}

	for p.r.position < int64(next) && p.err == nil {
		child := p.readNode()
		if child != nil {
			n.Children = append(n.Children, child)
		}
	}

	if p.err == nil {
		p.err = p.r.SkipTo(int64(next))
	}
	if p.err != nil && p.err != io.EOF {
		return nil
	}
	return n
}

func (p *binaryParser) Parse() (*Node, error) {
	if p.readString(20) != "Kaydara FBX Binary  " {
		return nil, fmt.Errorf("unknown fbx format")

	}
	p.r.SkipTo(27)
	root := &Node{Name: "_FBX_ROOT"}

	for p.err == nil {
		node := p.readNode()
		if node != nil {
			root.Children = append(root.Children, node)
		}
	}
	if p.err != nil && p.err != io.EOF {
		return nil, p.err
	}
	return root, nil
}
