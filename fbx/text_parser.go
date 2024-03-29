package fbx

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type tokenType int

const (
	Ident tokenType = iota
	Number
	String
	BlockStart
	BlockEnd
	Comma
	Colon
	Asterisk
	EOL
)

type textParser struct {
	r    io.Reader
	buf  []byte
	err  error
	line int
}

func (p *textParser) errorf(f string, a ...interface{}) error {
	if p.err == nil {
		p.err = fmt.Errorf("%s (line: %d)", fmt.Sprintf(f, a...), p.line+1)
	}
	return p.err
}

func (p *textParser) read() byte {
	if len(p.buf) > 0 {
		b := p.buf[0]
		p.buf = p.buf[1:]
		return b
	}
	b := []byte{0}
	if p.err == nil {
		_, err := io.ReadFull(p.r, b)
		p.err = err
	}
	return b[0]
}

func (p *textParser) getToken() (tokenType, string) {
	var c byte
	for p.err == nil {
		c = p.read()
		if c == ';' {
			for p.err == nil && c != '\n' {
				c = p.read()
			}
			p.line++
			return EOL, ""
		} else if c == '{' {
			return BlockStart, string(c)
		} else if c == '}' {
			return BlockEnd, string(c)
		} else if c == ',' {
			return Comma, string(c)
		} else if c == ':' {
			return Colon, string(c)
		} else if c == '*' {
			return Asterisk, string(c)
		} else if c >= '0' && c <= '9' || c == '.' || c == '-' {
			buf := []byte{c}
			c = p.read()
			for (c >= '0' && c <= '9' || c == '.' || c == 'e' || c == '-') && p.err == nil {
				buf = append(buf, c)
				c = p.read()
			}
			if p.err == nil {
				p.buf = append(p.buf, c)
			}
			return Number, string(buf)
		} else if c == '\n' {
			p.line++
			return EOL, ""
		} else if c == '"' {
			buf := []byte{}
			c = p.read()
			for c != '"' && p.err == nil {
				buf = append(buf, c)
				c = p.read()
			}
			return String, string(buf)
		} else if c >= 'A' && c <= 'z' {
			buf := []byte{}
			for (c >= 'A' && c <= 'z' || c >= '0' && c <= '9' || c == '-') && p.err == nil {
				buf = append(buf, c)
				c = p.read()
			}
			if p.err == nil {
				p.buf = append(p.buf, c)
			}
			return Ident, string(buf)
		}
	}
	return EOL, ""
}
func (p *textParser) Skip(t tokenType) bool {
	typ, s := p.getToken()
	if typ != t && p.err == nil {
		p.errorf("Skip token: error %v != %v(%v)", typ, t, s)
	}
	return typ == t
}

func (p *textParser) parseArrayProp() *Attribute {
	_, s := p.getToken()
	size, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		p.errorf("failed to parse num: '%v'", s)
	}
	p.Skip(BlockStart)
	for p.err == nil {
		if _, s := p.getToken(); s == ":" {
			break
		}
	}
	var dvalues []float64
	var hasPoint bool
	for i := 0; i < int(size) && p.err == nil; {
		typ, s := p.getToken()
		if typ == EOL || typ == Comma {
			continue
		} else if typ == BlockEnd {
			break
		} else if typ == Number {
			v, _ := strconv.ParseFloat(s, 64)
			dvalues = append(dvalues, v)
			hasPoint = hasPoint || strings.Contains(s, ".")
		} else {
			p.errorf("Invalid token: %v", s)
			break
		}
	}
	if len(dvalues) != int(size) {
		p.errorf("read array:.size: %v != %v", size, len(dvalues))
	}
	var values interface{} = dvalues
	if !hasPoint {
		var i32values []int32
		for _, v := range dvalues {
			i32values = append(i32values, int32(v))
		}
		values = i32values
	}
	return &Attribute{Value: values}
}

func (p *textParser) parseNodeList() []*Node {
	var nodes []*Node
	for p.err == nil {
		typ, s := p.getToken()
		if typ == EOL {
			continue
		} else if typ == BlockEnd {
			break
		} else if typ == Ident {
			p.Skip(Colon)
			node := &Node{Name: s}
			nodes = append(nodes, node)
			prev := Colon
			for p.err == nil {
				typ, s := p.getToken()
				if typ == EOL && prev != Comma {
					break
				}
				prev = typ
				if typ == BlockStart {
					node.Children = p.parseNodeList()
					break
				} else if typ == Number {
					if strings.Contains(s, ".") {
						v, err := strconv.ParseFloat(s, 64)
						if err != nil {
							p.errorf("failed to parse float: '%v'", s)
						}
						node.Attributes = append(node.Attributes, &Attribute{Value: v})
					} else {
						v, err := strconv.ParseInt(s, 10, 64)
						if err != nil {
							p.errorf("failed to parse integer: '%v'", s)
						}
						node.Attributes = append(node.Attributes, &Attribute{Value: v})
					}
				} else if typ == String {
					node.Attributes = append(node.Attributes, &Attribute{Value: s})
				} else if typ == Asterisk {
					node.Attributes = append(node.Attributes, p.parseArrayProp())
				}
			}
			if p.err == io.EOF {
				p.err = fmt.Errorf("Unexpected EOF")
			}
		} else {
			p.errorf("Unexpected Token '%v'", s)
			break
		}
	}
	return nodes
}

func (p *textParser) Parse() (*Node, error) {
	root := &Node{Name: "_FBX_ROOT"}
	root.Children = p.parseNodeList()

	// root.Dump(os.Stdout, 0, false)

	if p.err != nil && p.err != io.EOF {
		return nil, p.err
	}
	return root, nil
}
