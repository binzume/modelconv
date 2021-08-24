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

	GlobalSettings Object
	Objects        map[int]Object
	Scene          Object

	Materials []*Material

	RawNode *Node
}

type Property70 struct {
	node *Node
	Type string
}

func (p *Property70) Get(i int) *Property {
	if p == nil {
		return nil
	}
	return p.node.Prop(i + 4)
}

type Connection struct {
	Type string
	To   int
	From int
	Prop string
}
type Object interface {
	NodeName() string
	ID() int
	Name() string
	Kind() string
	GetProperty70(name string) *Property70
	FindRefs(name string) []Object
	AddRef(o Object)
}

type Obj struct {
	*Node
	Template   *Obj
	Refs       []Object
	properties map[string]*Property70 // lazty initialize
}

func (o *Obj) NodeName() string {
	return o.Node.Name
}
func (o *Obj) ID() int {
	return o.PropInt(0)
}
func (o *Obj) Name() string {
	return strings.ReplaceAll(o.PropString(1), "\x00\x01", "::")
}
func (o *Obj) Kind() string {
	return o.PropString(2)
}
func (o *Obj) GetProperty70(name string) *Property70 {
	if o.properties == nil {
		o.properties = map[string]*Property70{}
		for _, node := range o.FindChild("Properties70").GetChildren() {
			p := &Property70{node: node, Type: node.PropString(1)}
			o.properties[node.PropString(0)] = p
		}
	}
	if p, ok := o.properties[name]; ok {
		return p
	} else if o.Template != nil {
		return o.Template.GetProperty70(name)
	}
	return nil
}
func (o *Obj) FindRefs(typ string) []Object {
	var refs []Object
	for _, o := range o.Refs {
		if o.NodeName() == typ {
			refs = append(refs, o)
		}
	}
	return refs
}
func (o *Obj) AddRef(ref Object) {
	o.Refs = append(o.Refs, ref)
}

type Geometry struct {
	Obj
	Vertices []*geom.Vector3
	Faces    [][]int
	Normals  []*geom.Vector3
}

type Material struct {
	Obj
}

type Model struct {
	Obj
	Translation geom.Vector3
	Rotation    geom.Vector3
	Scaling     geom.Vector3
}

func (doc *Document) Dump(w io.Writer, full bool) {
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

func Save(doc *Document, path string) error {
	w, err := os.Create(path)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "; FBX project file")
	fmt.Fprintln(w, "; Generator: https://github.com/binzume/modelconv")
	fmt.Fprintln(w, "; -----------------------------------------------")
	for _, n := range doc.RawNode.Children {
		if n.Name != "FileId" {
			n.Dump(w, 0, true)
		}
	}

	return nil
}
