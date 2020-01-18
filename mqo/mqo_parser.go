package mqo

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"text/scanner"
)

type Parser struct {
	name string
	s    scanner.Scanner
}

func NewParser(r io.Reader, fname string) *Parser {
	var s scanner.Scanner
	s.Init(r)
	s.Filename = fname
	return &Parser{
		name: fname,
		s:    s,
	}
}

func (p *Parser) readFloat() float32 {
	tok := p.s.Scan()
	var s float32 = 1
	if p.s.TokenText() == "-" {
		tok = p.s.Scan()
		s = -1
	}
	if tok != scanner.Int && tok != scanner.Float {
		return 0
	}
	n, _ := strconv.ParseFloat(p.s.TokenText(), 32)
	return float32(n) * s
}

func (p *Parser) readInt() int {
	tok := p.s.Scan()
	if tok != scanner.Int {
		log.Printf("  Invalid num  %s\n", p.s.TokenText())
		return 0
	}
	n, _ := strconv.Atoi(p.s.TokenText())
	return n
}

func (p *Parser) readStr() string {
	p.s.Scan()
	return strings.Trim(p.s.TokenText(), "\"")
}

func (p *Parser) skip(t string) {
	p.s.Scan()
	if p.s.TokenText() != t {
		log.Printf("  Invalid token  %s %s\n", p.s.TokenText(), t)
	}
}

func (p *Parser) readMaterials(n int) []*Material {
	p.skip("{")
	var mm []*Material
	for i := 0; i < n; i++ {
		var m Material
		m.Name = p.readStr()
		line := p.s.Pos().Line

		for tok := p.s.Scan(); line == p.s.Pos().Line && tok != scanner.EOF; tok = p.s.Scan() {
			if p.s.TokenText() == "col" {
				p.skip("(")
				m.Color = Vector4{p.readFloat(), p.readFloat(), p.readFloat(), p.readFloat()}
				p.skip(")")
			} else if p.s.TokenText() == "dif" {
				p.skip("(")
				m.Diffuse = p.readFloat()
				p.skip(")")
			} else if p.s.TokenText() == "amb" {
				p.skip("(")
				m.Ambient = p.readFloat()
				p.skip(")")
			} else if p.s.TokenText() == "emi" {
				p.skip("(")
				m.Emmition = p.readFloat()
				p.skip(")")
			} else if p.s.TokenText() == "spc" {
				p.skip("(")
				m.Specular = p.readFloat()
				p.skip(")")
			} else if p.s.TokenText() == "power" {
				p.skip("(")
				m.Power = p.readFloat()
				p.skip(")")
			} else if p.s.TokenText() == "tex" {
				p.skip("(")
				m.Texture = p.readStr()
				p.skip(")")
			} else {
				log.Printf("  Mat %s %s\n", m.Name, p.s.TokenText())
			}
			p.s.Next()
			if p.s.Pos().Line != line {
				break
			}
		}
		mm = append(mm, &m)
	}
	p.skip("}")
	return mm
}

func (p *Parser) readObject() *Object {
	var o Object
	o.Name = p.readStr()
	log.Println("Read object:", o.Name)

	p.skip("{")

	for tok := p.s.Scan(); tok != scanner.EOF; tok = p.s.Scan() {
		if p.s.TokenText() == "}" {
			break
		}
		if tok == scanner.Ident && p.s.TokenText() == "vertex" {
			n := p.readInt()
			p.skip("{")
			for i := 0; i < n; i++ {
				v := Vector3{X: p.readFloat(), Y: p.readFloat(), Z: p.readFloat()}
				o.Vertexes = append(o.Vertexes, &v)
			}
			p.skip("}")
		} else if tok == scanner.Ident && p.s.TokenText() == "face" {
			n := p.readInt()
			p.skip("{")

			o.Faces = make([]*Face, n)
			for i := 0; i < n; i++ {
				vn := p.readInt()
				line := p.s.Pos().Line
				var f Face
				o.Faces[i] = &f

				for tok = p.s.Scan(); line == p.s.Pos().Line && tok != scanner.EOF; tok = p.s.Scan() {
					if p.s.TokenText() == "V" {
						f.Verts = make([]int, vn)
						p.skip("(")
						for i := 0; i < vn; i++ {
							f.Verts[i] = p.readInt()
						}
						p.skip(")")
					} else if p.s.TokenText() == "M" {
						p.skip("(")
						f.Material = p.readInt()
						p.skip(")")
					} else if p.s.TokenText() == "UV" {
						p.skip("(")
						f.UVs = make([]Vector2, vn)
						for i := 0; i < vn; i++ {
							f.UVs[i] = Vector2{p.readFloat(), p.readFloat()}
						}
						p.skip(")")
					} else {
						log.Printf("  Face %d: token: %s\n", i, p.s.TokenText())
					}
					p.s.Next()
					if p.s.Pos().Line != line {
						break
					}
				}
			}
			p.skip("}")
		}
	}
	return &o
}

func (p *Parser) Parse() (*MQODocument, error) {
	var doc MQODocument
	for tok := p.s.Scan(); tok != scanner.EOF; tok = p.s.Scan() {
		if tok == scanner.Ident && p.s.TokenText() == "Material" {
			doc.Materials = p.readMaterials(p.readInt())
		} else if tok == scanner.Ident && p.s.TokenText() == "Object" {
			doc.Objects = append(doc.Objects, p.readObject())
		} else if tok == scanner.Ident && p.s.TokenText() == "Eof" {
			break
		} else {
			// log.Println(" > ", p.s.TokenText())
		}
	}
	if p.s.ErrorCount > 0 {
		return &doc, fmt.Errorf("Parse error (count:%d)", p.s.ErrorCount)
	}
	return &doc, nil
}

func Parse(r io.Reader, fname string) (*MQODocument, error) {
	p := NewParser(r, fname)
	return p.Parse()
}
