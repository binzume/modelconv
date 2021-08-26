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
	if n == nil {
		return defvalue
	}
	return n.Attr(0).ToInt(defvalue)
}

func (n *Node) GetInt64(defvalue int64) int64 {
	if n == nil {
		return defvalue
	}
	return n.Attr(0).ToInt64(defvalue)
}

func (n *Node) GetFloat32(defvalue float32) float32 {
	if n == nil {
		return defvalue
	}
	return n.Attr(0).ToFloat32(defvalue)
}

func (n *Node) GetString(defvalue string) string {
	if n == nil {
		return defvalue
	}
	return n.Attr(0).ToString(defvalue)
}

func (n *Node) GetInt32Array() []int32 {
	if n == nil {
		return nil
	}
	return n.Attr(0).ToInt32Array()
}

func (n *Node) GetFloat32Array() []float32 {
	if n == nil {
		return nil
	}
	return n.Attr(0).ToFloat32Array()
}

type Attribute struct {
	Value interface{}
	Count uint
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

func (p AttributeList) ToString(defvalue string) string {
	return p.Get(0).ToString(defvalue)
}

func (p AttributeList) ToVector3(x, y, z float32) geom.Vector3 {
	return geom.Vector3{X: p.Get(0).ToFloat32(x), Y: p.Get(1).ToFloat32(y), Z: p.Get(2).ToFloat32(z)}
}

func (p *Attribute) ToInt(defvalue int) int {
	return int(p.ToInt64(int64(defvalue)))
}

func (p *Attribute) ToInt64(defvalue int64) int64 {
	if p == nil {
		return defvalue
	}
	if v, ok := p.Value.(byte); ok {
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
	} else if v, ok := p.Value.(int16); ok {
		return float64(v)
	} else if v, ok := p.Value.(int32); ok {
		return float64(v)
	} else if v, ok := p.Value.(int64); ok {
		return float64(v)
	}
	return defvalue
}

func (p *Attribute) ToString(defvalue string) string {
	if p == nil {
		return defvalue
	}
	if v, ok := p.Value.(string); ok {
		return string(v)
	} else if v, ok := p.Value.([]byte); ok {
		return string(v)
	}
	return defvalue
}

func (p *Attribute) ToVec3Array() []*geom.Vector3 {
	if p == nil {
		return nil
	}
	v := p.ToFloat32Array()
	var vv []*geom.Vector3
	for i := 0; i < len(v)/3; i++ {
		vv = append(vv, &geom.Vector3{X: v[i*3], Y: v[i*3+1], Z: v[i*3+2]})
	}
	return vv
}

func (p *Attribute) ToVec2Array() []*geom.Vector2 {
	if p == nil {
		return nil
	}
	v := p.ToFloat32Array()
	var vv []*geom.Vector2
	for i := 0; i < len(v)/2; i++ {
		vv = append(vv, &geom.Vector2{X: v[i*2], Y: v[i*2+1]})
	}
	return vv
}

func (p *Attribute) ToInt32Array() []int32 {
	if p == nil {
		return nil
	}
	var r []int32
	if vv, ok := p.Value.([]byte); ok {
		for _, v := range vv {
			r = append(r, int32(v))
		}
	} else if vv, ok := p.Value.([]int32); ok {
		return vv
	} else if vv, ok := p.Value.([]int16); ok {
		for _, v := range vv {
			r = append(r, int32(v))
		}
	} else if vv, ok := p.Value.([]int64); ok {
		for _, v := range vv {
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
	if vv, ok := p.Value.([]float32); ok {
		return vv
	} else if vv, ok := p.Value.([]float64); ok {
		for _, v := range vv {
			r = append(r, float32(v))
		}
	} else if vv, ok := p.Value.([]int32); ok {
		for _, v := range vv {
			r = append(r, float32(v))
		}
	} else if vv, ok := p.Value.([]int64); ok {
		for _, v := range vv {
			r = append(r, float32(v))
		}
	}
	return r
}

func (p *Attribute) String() string {
	switch v := p.Value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case []byte:
		return fmt.Sprintf("\"%v\"", v)
	default:
		return fmt.Sprint(v)
	}
}

func (n *Node) Dump(w io.Writer, d int, full bool) {
	fmt.Fprint(w, strings.Repeat("  ", d), n.Name, ":")
	var arrayReplacer = strings.NewReplacer("[", "{ a:", "]", "}", " ", ", ")
	for i, p := range n.Attributes {
		if !full && p.Count > 16 {
			fmt.Fprintf(w, " *%d { SKIPPED }", p.Count)
			continue
		}
		s := p.String()
		if p.Count > 0 {
			s = fmt.Sprint("*", p.Count, " ", arrayReplacer.Replace(s))
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
