package mmd

import (
	"encoding/binary"
	"io"
)

type baseParser struct {
	r io.Reader
}

func (p *baseParser) read(v interface{}) error {
	return binary.Read(p.r, binary.LittleEndian, v)
}

func (p *baseParser) readUint8() uint8 {
	var v uint8
	binary.Read(p.r, binary.LittleEndian, &v)
	return v
}

func (p *baseParser) readUint16() uint16 {
	var v uint16
	binary.Read(p.r, binary.LittleEndian, &v)
	return v
}

func (p *baseParser) readInt() int {
	var v uint32
	binary.Read(p.r, binary.LittleEndian, &v)
	return int(v)
}

func (p *baseParser) readFloat() float32 {
	var v float32
	binary.Read(p.r, binary.LittleEndian, &v)
	return v
}

func (p *baseParser) readVUInt(sz byte) int {
	if sz == 1 {
		var v uint8
		binary.Read(p.r, binary.LittleEndian, &v)
		return int(v)
	}
	if sz == 2 {
		var v uint16
		binary.Read(p.r, binary.LittleEndian, &v)
		return int(v)
	}
	if sz == 4 {
		var v uint32
		binary.Read(p.r, binary.LittleEndian, &v)
		return int(v)
	}
	return 0
}

func (p *baseParser) readVInt(sz byte) int {
	if sz == 1 {
		var v int8
		binary.Read(p.r, binary.LittleEndian, &v)
		return int(v)
	}
	if sz == 2 {
		var v int16
		binary.Read(p.r, binary.LittleEndian, &v)
		return int(v)
	}
	if sz == 4 {
		var v int32
		binary.Read(p.r, binary.LittleEndian, &v)
		return int(v)
	}
	return 0
}
