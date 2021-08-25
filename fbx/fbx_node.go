package fbx

import (
	"fmt"
	"io"
	"strings"

	"github.com/binzume/modelconv/geom"
)

type Node struct {
	Name       string
	Properties PropertyList
	Children   []*Node
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

func (n *Node) Prop(i int) *Property {
	if n == nil {
		return nil
	}
	return n.Properties.Get(i)
}

func (n *Node) PropValue(i int) interface{} {
	if n == nil {
		return nil
	}
	return n.Properties.Get(i).Value
}

func (n *Node) PropInt(i int) int {
	return n.Prop(i).ToInt(0)
}

func (n *Node) PropFloat(i int) float32 {
	return n.Prop(i).ToFloat32(0)
}

func (n *Node) PropString(i int) string {
	return n.Prop(i).ToString("")
}

type Property struct {
	Value interface{}
	Count uint
}

type PropertyList []*Property

func (p PropertyList) Get(i int) *Property {
	if i >= len(p) {
		return nil
	}
	return p[i]
}

func (p *Property) ToInt(defvalue int) int {
	return int(p.ToInt64(int64(defvalue)))
}

func (p *Property) ToInt64(defvalue int64) int64 {
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

func (p *Property) ToFloat32(defvalue float32) float32 {
	if p == nil {
		return defvalue
	}
	if v, ok := p.Value.(float32); ok {
		return float32(v)
	} else if v, ok := p.Value.(float64); ok {
		return float32(v)
	} else if v, ok := p.Value.(int16); ok {
		return float32(v)
	} else if v, ok := p.Value.(int32); ok {
		return float32(v)
	} else if v, ok := p.Value.(int64); ok {
		return float32(v)
	}
	return defvalue
}

func (p *Property) ToString(defvalue string) string {
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

func (p *Property) ToVec3Array() []*geom.Vector3 {
	if p == nil {
		return nil
	}
	var vv []*geom.Vector3
	if v, ok := p.Value.([]float32); ok {
		for i := 0; i < len(v)/3; i++ {
			vv = append(vv, &geom.Vector3{X: v[i*3], Y: v[i*3+1], Z: v[i*3+2]})
		}
	} else if v, ok := p.Value.([]float64); ok {
		for i := 0; i < len(v)/3; i++ {
			vv = append(vv, &geom.Vector3{X: float32(v[i*3]), Y: float32(v[i*3+1]), Z: float32(v[i*3+2])})
		}
	}
	return vv
}

func (p *Property) ToVec2Array() []*geom.Vector2 {
	if p == nil {
		return nil
	}
	var vv []*geom.Vector2
	if v, ok := p.Value.([]float32); ok {
		for i := 0; i < len(v)/2; i++ {
			vv = append(vv, &geom.Vector2{X: v[i*2], Y: v[i*2+1]})
		}
	} else if v, ok := p.Value.([]float64); ok {
		for i := 0; i < len(v)/2; i++ {
			vv = append(vv, &geom.Vector2{X: float32(v[i*2]), Y: float32(v[i*2+1])})
		}
	}
	return vv
}

func (p *Property) ToInt32Array() []int32 {
	if p == nil {
		return nil
	}
	var r []int32
	if vv, ok := p.Value.([]byte); ok {
		for _, v := range vv {
			r = append(r, int32(v))
		}
	} else if vv, ok := p.Value.([]int32); ok {
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

func (p *Property) ToFloat32Array() []float32 {
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

func (p *Property) String() string {
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
	for i, p := range n.Properties {
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
	if len(n.Children) > 0 || len(n.Properties) == 0 {
		fmt.Fprintln(w, " {")
		for _, c := range n.Children {
			c.Dump(w, d+1, full)
		}
		fmt.Fprintln(w, strings.Repeat("  ", d)+"}")
	} else {
		fmt.Fprintln(w, "")
	}
}
