package fbx

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/binzume/modelconv/geom"
)

type Document struct {
	FileId       []byte
	Creator      string
	CreationTime string

	Objects     map[int]*Object
	Connections []*Connection

	Meshes    []*Mesh
	Materials []*Material

	RawNode *Node
}

type Connection struct {
	Type string
	To   int
	From int
}

type Object struct {
	*Node
}

func (o *Object) ID() int {
	return o.PropInt(0)
}
func (o *Object) Name() string {
	return strings.ReplaceAll(o.PropString(1), "\x00\x01", "::")
}

type Mesh struct {
	Object
	Vertices []*geom.Vector3
	Faces    [][]int
	Normals  []*geom.Vector3
}

type Material struct {
	Object
}

type Node struct {
	Name       string
	Properties []*Property
	Children   []*Node
}

func (n *Node) Child(name string) *Node {
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

func (n *Node) ChildOrEmpty(name string) *Node {
	if c := n.Child(name); c != nil {
		return c
	}
	return &Node{}
}

func (n *Node) Prop(i int) *Property {
	if n == nil || i >= len(n.Properties) {
		return nil
	}
	return n.Properties[i]
}

func (n *Node) PropValue(i int) interface{} {
	if n == nil || i >= len(n.Properties) {
		return nil
	}
	return n.Properties[i].Value
}

func (n *Node) PropString(i int) string {
	if s, ok := n.PropValue(i).(string); ok {
		return s
	}
	return ""
}

func (n *Node) PropInt(i int) int {
	if v, ok := n.PropValue(i).(byte); ok {
		return int(v)
	} else if v, ok := n.PropValue(i).(int16); ok {
		return int(v)
	} else if v, ok := n.PropValue(i).(int32); ok {
		return int(v)
	} else if v, ok := n.PropValue(i).(int64); ok {
		return int(v)
	}
	return 0
}

type Property struct {
	Type  uint8
	Value interface{}
	Count uint
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
	if len(n.Children) > 0 {
		fmt.Fprintln(w, " {")
		for _, c := range n.Children {
			c.Dump(w, d+1, full)
		}
		fmt.Fprintln(w, strings.Repeat("  ", d)+"}")
	} else {
		fmt.Fprintln(w, "")
	}
}

func (doc *Document) Dump(w io.Writer, full bool) {
	fmt.Fprintln(w, "; FBX project file")
	fmt.Fprintln(w, "; Generator: https://github.com/binzume/modelconv")
	fmt.Fprintln(w, "; -----------------------------------------------")
	for _, n := range doc.RawNode.Children {
		n.Dump(w, 0, full)
	}
}

func Load(path string) (*Document, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	p := binaryParser{r: &positionReader{r: r}}
	root, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return BuildDocument(root)
}
