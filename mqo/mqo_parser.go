package mqo

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Parser for mqo file.
type Parser struct {
	name string
	r    io.Reader
	s    scanner.Scanner
	Open func(name string) (io.ReadCloser, error)
}

// NewParser returns new parser.
func NewParser(r io.Reader, path string) *Parser {
	p := &Parser{
		name: path,
		r:    r,
	}
	if path != "" {
		p.s.Filename = path
		p.Open = func(name string) (io.ReadCloser, error) {
			return os.Open(filepath.Dir(path) + "/" + name)
		}
	}
	return p
}

type basckSlashReplacer struct{}

func (*basckSlashReplacer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	n := copy(dst, src)
	for i := 0; i < n; i++ {
		if dst[i] == '\\' {
			dst[i] = '/'
		}
	}
	return n, n, nil
}

func (*basckSlashReplacer) Reset() {
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

func (p *Parser) readIdent() string {
	p.s.Scan()
	return p.s.TokenText()
}

func (p *Parser) skipN(n int) {
	for i := 0; i < n; i++ {
		p.s.Scan()
	}
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
			log.Printf("  skip %s %s\n", name, p.s.TokenText())
			p.skip("(")
			for tok := p.s.Scan(); line == p.s.Pos().Line && tok != scanner.EOF; tok = p.s.Scan() {
				if p.s.TokenText() == ")" {
					break
				}
			}
		}
		if p.s.Peek() == 0x0d || p.s.Peek() == 0x0a {
			break
		}
	}
}

func (p *Parser) skipBlock() {
	p.s.Error = func(s *scanner.Scanner, msg string) {
		s.ErrorCount--
	}
	defer func() { p.s.Error = nil }()
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
		"col":     func() { m.Color = Vector4{X: p.readFloat(), Y: p.readFloat(), Z: p.readFloat(), W: p.readFloat()} },
		"emi_col": func() { m.EmissionColor = &Vector3{X: p.readFloat(), Y: p.readFloat(), Z: p.readFloat()} },
		"dif":     func() { m.Diffuse = p.readFloat() },
		"amb":     func() { m.Ambient = p.readFloat() },
		"emi":     func() { m.Emission = p.readFloat() },
		"spc":     func() { m.Specular = p.readFloat() },
		"power":   func() { m.Power = p.readFloat() },
		"tex":     func() { m.Texture = p.readStr() },
		"dbls":    func() { m.DoubleSided = p.readInt() != 0 },
		"uid":     func() { m.UID = p.readInt() },
		"shader":  func() { m.Shader = p.readInt() },
	}, "Material "+m.Name)
	return &m
}

func (p *Parser) readMaterialEx() (int, *MaterialEx2) {
	var ex MaterialEx2
	p.skip("material")
	mid := p.readInt()

	readTypedKeyValue := func() (string, interface{}) {
		n := p.readStr()
		t := p.readIdent()
		if t == "." {
			n += t + p.readStr()
			t = p.readIdent()
		}
		if t == "int" {
			return n, p.readInt()
		} else if t == "float" {
			return n, p.readFloat()
		} else if t == "bool" {
			v := p.readStr()
			return n, v == "true" || v == "1"
		} else if t == "color" {
			return n, []float32{p.readFloat(), p.readFloat(), p.readFloat(), p.readFloat()}
		}
		return n, p.readStr()
	}

	p.procObj(map[string]func(){
		"shadertype": func() { ex.ShaderType = p.readStr() },
		"shadername": func() { ex.ShaderName = p.readStr() },
		"shaderparam": func() {
			p.procArray(func(n int) {
				ex.ShaderParams = map[string]interface{}{}
			}, func(i int) {
				key, value := readTypedKeyValue()
				ex.ShaderParams[key] = value
			}, "shaderparam")
		},
		"shadermappingparam": func() {
			if ex.ShaderMappingParams == nil {
				ex.ShaderMappingParams = map[string]map[string]interface{}{}
			}
			name := p.readStr()
			p.procArray(func(n int) {
				ex.ShaderMappingParams[name] = map[string]interface{}{}
			}, func(i int) {
				key, value := readTypedKeyValue()
				ex.ShaderMappingParams[name][key] = value
			}, "shadermappingparam")
		},
	}, "MaterialEx2")
	return mid, &ex
}

func (p *Parser) readObject() *Object {
	o := NewObject(p.readStr())

	p.procObj(map[string]func(){
		"uid":        func() { o.UID = p.readInt() },
		"depth":      func() { o.Depth = p.readInt() },
		"visible":    func() { o.Visible = p.readInt() > 0 },
		"shading":    func() { o.Shading = p.readInt() },
		"facet":      func() { o.Facet = p.readFloat() },
		"patch":      func() { o.Patch = p.readInt() },
		"segment":    func() { o.Segment = p.readInt() },
		"mirror":     func() { o.Mirror = p.readInt() },
		"mirror_dis": func() { o.MirrorDis = p.readFloat() },
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
							f.UVs[i] = Vector2{X: p.readFloat(), Y: p.readFloat()}
						}
					},
					"CRS": func() {
						for i := 0; i < vn; i++ {
							p.readFloat()
						}
					},
					"UID": func() { f.UID = p.readInt() },
				}, fmt.Sprintf("Object %v F%v\n", o.Name, i))
			}, "face")
		},
		"vertexattr": func() {
			p.procObj(map[string]func(){
				"uid": func() {
					p.skip("{")
					for i := 0; i < len(o.Vertexes); i++ {
						o.VertexByUID[p.readInt()] = i
					}
					p.skip("}")
				},
			}, "vertexattr")
		},
	}, fmt.Sprintf("Object %v\n", o.Name))
	return o
}

func (p *Parser) detectCodePage() {
	buf := make([]byte, 128)
	n, _ := p.r.Read(buf)
	p.r = io.MultiReader(bytes.NewReader(buf[:n]), p.r)
	if matched, _ := regexp.Match(`CodePage\s+utf8`, buf[:n]); !matched {
		p.r = transform.NewReader(p.r, transform.Chain(japanese.ShiftJIS.NewDecoder(), &basckSlashReplacer{}))
	}
}

func (p *Parser) Parse() (*Document, error) {
	p.detectCodePage()
	p.s.Init(p.r)

	var doc Document
	var mqxFile string
	for tok := p.s.Scan(); tok != scanner.EOF; tok = p.s.Scan() {
		if tok == scanner.Ident && p.s.TokenText() == "Material" {
			p.procArray(func(n int) {}, func(i int) {
				doc.Materials = append(doc.Materials, p.readMaterial())
			}, "Material")
		} else if tok == scanner.Ident && p.s.TokenText() == "Object" {
			doc.Objects = append(doc.Objects, p.readObject())
		} else if tok == scanner.Ident && p.s.TokenText() == "Thumbnail" {
			p.skipN(5)
			p.skip("{")
			p.skipBlock()
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
	if mqxFile != "" && p.Open != nil {
		r, _ := p.Open(mqxFile)
		if r != nil {
			defer r.Close()
			if mqx, err := ReadMQX(r); err == nil {
				doc.Plugins = mqx.Plugins
			}
		}
	}
	return &doc, nil
}

func LoadMQOZ(path string) (*Document, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	for _, f := range z.File {
		if strings.HasSuffix(f.Name, ".mqo") {
			r, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer r.Close()
			parser := NewParser(r, path)
			parser.Open = func(name string) (io.ReadCloser, error) {
				for _, f := range z.File {
					if f.Name == name {
						return f.Open()
					}
				}
				return nil, os.ErrNotExist
			}
			return parser.Parse()
		}
	}
	return nil, os.ErrNotExist
}

func Load(path string) (*Document, error) {
	if strings.HasSuffix(path, ".mqoz") {
		return LoadMQOZ(path)
	}
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	parser := NewParser(r, path)
	return parser.Parse()
}
