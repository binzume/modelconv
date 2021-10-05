package fbx

import (
	"fmt"
	"io"
	"strings"

	"github.com/binzume/modelconv/geom"
)

type Node struct {
	Name       string
	Attributes AttributeList
	Children   []*Node
}

func (n *Node) AddChild(node *Node) {
	n.Children = append(n.Children, node)
}

func (n *Node) RemoveChild(node *Node) bool {
	for i, c := range n.Children {
		if c == node {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			return true
		}
	}
	return false
}

func (n *Node) FindChild(name string) *Node {
	if n == nil {
		return nil
	}
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (n *Node) GetChildren() []*Node {
	if n == nil {
		return nil
	}
	return n.Children
}

func (n *Node) Attr(i int) *Attribute {
	if n == nil {
		return nil
	}
	return n.Attributes.Get(i)
}

func (n *Node) GetInt(defvalue int) int {
	return n.Attr(0).ToInt(defvalue)
}

func (n *Node) GetInt64(defvalue int64) int64 {
	return n.Attr(0).ToInt64(defvalue)
}

func (n *Node) GetFloat32(defvalue float32) float32 {
	return n.Attr(0).ToFloat32(defvalue)
}

func (n *Node) GetString() string {
	return n.Attr(0).ToString()
}

func (n *Node) GetInt32Array() []int32 {
	return n.Attr(0).ToInt32Array()
}

func (n *Node) GetFloat32Array() []float32 {
	return n.Attr(0).ToFloat32Array()
}

func (n *Node) GetVec2Array() []*geom.Vector2 {
	return n.Attr(0).ToVec2Array()
}

func (n *Node) GetVec3Array() []*geom.Vector3 {
	return n.Attr(0).ToVec3Array()
}

type Attribute struct {
	Value     interface{}
	ArraySize uint
}

type AttributeList []*Attribute

func (p AttributeList) Get(i int) *Attribute {
	if i >= len(p) {
		return nil
	}
	return p[i]
}

func (p AttributeList) ToInt(defvalue int) int {
	return p.Get(0).ToInt(defvalue)
}

func (p AttributeList) ToFloat32(defvalue float32) float32 {
	return p.Get(0).ToFloat32(defvalue)
}

func (p AttributeList) ToString() string {
	return p.Get(0).ToString()
}

func (p AttributeList) ToVector3(x, y, z float32) *geom.Vector3 {
	return &geom.Vector3{X: p.Get(0).ToFloat32(x), Y: p.Get(1).ToFloat32(y), Z: p.Get(2).ToFloat32(z)}
}

func (p *Attribute) ToInt(defvalue int) int {
	return int(p.ToInt64(int64(defvalue)))
}

func (p *Attribute) ToInt64(defvalue int64) int64 {
	if p == nil {
		return defvalue
	}
	if v, ok := p.Value.(int8); ok {
		return int64(v)
	} else if v, ok := p.Value.(int16); ok {
		return int64(v)
	} else if v, ok := p.Value.(int32); ok {
		return int64(v)
	} else if v, ok := p.Value.(int64); ok {
		return int64(v)
	}
	return defvalue
}

func (p *Attribute) ToFloat32(defvalue float32) float32 {
	return float32(p.ToFloat64(float64(defvalue)))
}

func (p *Attribute) ToFloat64(defvalue float64) float64 {
	if p == nil {
		return defvalue
	}
	if v, ok := p.Value.(float32); ok {
		return float64(v)
	} else if v, ok := p.Value.(float64); ok {
		return float64(v)
	} else if v, ok := p.Value.(int8); ok {
		return float64(v)
	} else if v, ok := p.Value.(int16); ok {
		return float64(v)
	} else if v, ok := p.Value.(int32); ok {
		return float64(v)
	} else if v, ok := p.Value.(int64); ok {
		return float64(v)
	}
	return defvalue
}

func (p *Attribute) ToString() string {
	if p == nil {
		return ""
	}
	if v, ok := p.Value.(string); ok {
		return string(v)
	} else if v, ok := p.Value.([]byte); ok {
		return string(v)
	}
	return ""
}

func (p *Attribute) ToVec3Array() []*geom.Vector3 {
	if p == nil {
		return nil
	}
	farray := p.ToFloat32Array()
	var vv []*geom.Vector3
	for i := 0; i < len(farray)/3; i++ {
		vv = append(vv, &geom.Vector3{X: farray[i*3], Y: farray[i*3+1], Z: farray[i*3+2]})
	}
	return vv
}

func (p *Attribute) ToVec2Array() []*geom.Vector2 {
	if p == nil {
		return nil
	}
	farray := p.ToFloat32Array()
	var vv []*geom.Vector2
	for i := 0; i < len(farray)/2; i++ {
		vv = append(vv, &geom.Vector2{X: farray[i*2], Y: farray[i*2+1]})
	}
	return vv
}

func (p *Attribute) ToInt32Array() []int32 {
	if p == nil {
		return nil
	}
	var r []int32
	switch v := p.Value.(type) {
	case []int32:
		r = v
	case []int8:
		for _, v := range v {
			r = append(r, int32(v))
		}
	case []int16:
		for _, v := range v {
			r = append(r, int32(v))
		}
	case []int64:
		for _, v := range v {
			r = append(r, int32(v))
		}
	}
	return r
}

func (p *Attribute) ToFloat32Array() []float32 {
	if p == nil {
		return nil
	}
	var r []float32
	switch v := p.Value.(type) {
	case []float32:
		r = v
	case []float64:
		for _, v := range v {
			r = append(r, float32(v))
		}
	case []int32:
		for _, v := range v {
			r = append(r, float32(v))
		}
	case []int64:
		for _, v := range v {
			r = append(r, float32(v))
		}
	}
	return r
}

func (p *Attribute) String() string {
	switch v := p.Value.(type) {
	case string:
		segments := strings.SplitN(v, "\x00\x01", 2)
		if len(segments) == 2 {
			segments[0], segments[1] = segments[1], segments[0]
		}
		return fmt.Sprintf("%q", strings.Join(segments, "::"))
	case []byte:
		return fmt.Sprintf("\"%v\"", v)
	default:
		return fmt.Sprint(v)
	}
}

func (n *Node) Dump(w io.Writer, d int, full bool) {
	fmt.Fprint(w, strings.Repeat("  ", d), n.Name, ":")
	var arrayReplacer = strings.NewReplacer("[", "{\n      a:", "]", "\n    }", " ", ", ")
	for i, p := range n.Attributes {
		if !full && p.ArraySize > 16 {
			fmt.Fprintf(w, " *%d { SKIPPED }", p.ArraySize)
			continue
		}
		s := p.String()
		if p.ArraySize > 0 {
			s = fmt.Sprint("*", p.ArraySize, " ", arrayReplacer.Replace(s))
		}
		if i == 0 {
			fmt.Fprint(w, " ", s)
		} else {
			fmt.Fprint(w, ", ", s)
		}
	}
	if len(n.Children) > 0 || len(n.Attributes) == 0 {
		fmt.Fprintln(w, " {")
		for _, c := range n.Children {
			c.Dump(w, d+1, full)
		}
		fmt.Fprintln(w, strings.Repeat("  ", d)+"}")
	} else {
		fmt.Fprintln(w, "")
	}
}
