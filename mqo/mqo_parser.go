package mqo

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/scanner"
)

// Parser for mqo file.
type Parser struct {
	name string
	s    scanner.Scanner
}

// NewParser returns new parser.
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
		log.Printf("  Invalid token  %s != %s\n", p.s.TokenText(), t)
	}
}

func (p *Parser) procAttrs(handlers map[string]func(), name string) {
	line := p.s.Pos().Line
	for tok := p.s.Scan(); line == p.s.Pos().Line && tok != scanner.EOF; tok = p.s.Scan() {
		if handler, ok := handlers[p.s.TokenText()]; ok {
			p.skip("(")
			handler()
			p.skip(")")
		} else {
			log.Printf("  %s %s\n", name, p.s.TokenText())
		}
		if p.s.Peek() == 0x0d || p.s.Peek() == 0x0a {
			break
		}
	}
}

func (p *Parser) skipBlock() {
	for tok := p.s.Scan(); tok != scanner.EOF; tok = p.s.Scan() {
		if p.s.TokenText() == "}" {
			return
		}
		if p.s.TokenText() == "{" {
			p.skipBlock()
		}
	}
}

func (p *Parser) procArray(init, elem func(n int), name string) {
	n := p.readInt()
	p.skip("{")
	init(n)
	for i := 0; i < n; i++ {
		elem(i)
	}
	p.skip("}")
}

func (p *Parser) procObj(handlers map[string]func(), name string) {
	p.skip("{")
	for tok := p.s.Scan(); tok != scanner.EOF; tok = p.s.Scan() {
		if p.s.TokenText() == "}" {
			break
		}
		if p.s.TokenText() == "{" {
			p.skipBlock()
		}
		if handler, ok := handlers[p.s.TokenText()]; ok {
			handler()
		}
	}
}

func (p *Parser) readMaterial() *Material {
	var m Material
	m.Name = p.readStr()
	p.procAttrs(map[string]func(){
		"col":   func() { m.Color = Vector4{p.readFloat(), p.readFloat(), p.readFloat(), p.readFloat()} },
		"dif":   func() { m.Diffuse = p.readFloat() },
		"amb":   func() { m.Ambient = p.readFloat() },
		"emi":   func() { m.Emmition = p.readFloat() },
		"spc":   func() { m.Specular = p.readFloat() },
		"power": func() { m.Power = p.readFloat() },
		"tex":   func() { m.Texture = p.readStr() },
		"uid":   func() { p.readInt() },
	}, fmt.Sprintf("Material %s\n", m.Name))
	return &m
}

func (p *Parser) readMaterialEx() (int, *MaterialEx2) {
	var ex MaterialEx2
	p.skip("material")
	mid := p.readInt()
	p.procObj(map[string]func(){
		"shadertype": func() { ex.ShaderType = p.readStr() },
		"shadername": func() { ex.ShaderName = p.readStr() },
		"shaderparam": func() {
			ex.ShaderParams = map[string]interface{}{}
			p.procArray(func(n int) {
			}, func(i int) {
				p.s.Scan()
				n := p.s.TokenText()
				p.s.Scan()
				t := p.s.TokenText()
				p.s.Scan()
				v := p.s.TokenText()
				if t == "int" {
					ex.ShaderParams[n], _ = strconv.Atoi(v)
				} else if t == "bool" {
					ex.ShaderParams[n] = v == "true"
				}
			}, "shaderparam")
		},
	}, "MaterialEx2")
	return mid, &ex
}

func (p *Parser) readObject() *Object {
	o := NewObject(p.readStr())
	log.Println("Read object:", o.Name)

	p.procObj(map[string]func(){
		"depth":   func() { o.Depth = p.readInt() },
		"visible": func() { o.Visible = p.readInt() > 0 },
		"vertex": func() {
			p.procArray(func(n int) {
				o.Vertexes = make([]*Vector3, n)
			}, func(i int) {
				o.Vertexes[i] = &Vector3{X: p.readFloat(), Y: p.readFloat(), Z: p.readFloat()}
			}, "vertex")
		},
		"face": func() {
			p.procArray(func(n int) {
				o.Faces = make([]*Face, n)
			}, func(i int) {
				var f Face
				o.Faces[i] = &f
				vn := p.readInt()
				p.procAttrs(map[string]func(){
					"V": func() {
						f.Verts = make([]int, vn)
						for i := 0; i < vn; i++ {
							f.Verts[i] = p.readInt()
						}
					},
					"M": func() { f.Material = p.readInt() },
					"UV": func() {
						f.UVs = make([]Vector2, vn)
						for i := 0; i < vn; i++ {
							f.UVs[i] = Vector2{p.readFloat(), p.readFloat()}
						}
					},
					"UID": func() { p.readInt() },
				}, fmt.Sprintf("Object %v F%v\n", o.Name, i))
			}, "face")
		},
	}, fmt.Sprintf("Object %v\n", o.Name))
	return o
}

func (p *Parser) Parse(fname string) (*MQODocument, error) {
	var doc MQODocument
	var mqxFile string
	for tok := p.s.Scan(); tok != scanner.EOF; tok = p.s.Scan() {
		if tok == scanner.Ident && p.s.TokenText() == "Material" {
			p.procArray(func(n int) {}, func(i int) {
				doc.Materials = append(doc.Materials, p.readMaterial())
			}, "Material")
		} else if tok == scanner.Ident && p.s.TokenText() == "Object" {
			doc.Objects = append(doc.Objects, p.readObject())
		} else if tok == scanner.Ident && p.s.TokenText() == "MaterialEx2" {
			p.procArray(func(n int) {}, func(i int) {
				mid, ex := p.readMaterialEx()
				if mid >= 0 && mid < len(doc.Materials) {
					doc.Materials[mid].Ex2 = ex
				}
			}, "MaterialEx2")
		} else if tok == scanner.Ident && p.s.TokenText() == "IncludeXml" {
			mqxFile = p.readStr()
		} else if tok == scanner.Ident && p.s.TokenText() == "Eof" {
			break
		} else {
			// log.Println(" > ", p.s.TokenText())
		}
	}
	if p.s.ErrorCount > 0 {
		return &doc, fmt.Errorf("Parse error (count:%d)", p.s.ErrorCount)
	}
	if mqxFile != "" && fname != "" {
		mqxPath := fname[0:len(fname)-len(filepath.Ext(fname))] + ".mqx"
		r, _ := os.Open(mqxPath)
		if r != nil {
			defer r.Close()
			if mqx, err := ReadMQX(r); err == nil {
				doc.Plugins = mqx.Plugins
			}
		}
	}
	return &doc, nil
}

// Parse mqo file.
func Parse(r io.Reader, fname string) (*MQODocument, error) {
	p := NewParser(r, fname)
	return p.Parse(fname)
}
