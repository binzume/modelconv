package fbx

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/binzume/modelconv/geom"
)

type Document struct {
	FileId       []byte
	Creator      string
	CreationTime string

	GlobalSettings Object
	Objects        map[int64]Object
	Scene          Object

	Materials []*Material

	RawNode *Node
}

type Property70 struct {
	AttributeList
	Type  string
	Label string
	Flag  string
}

type Connection struct {
	Type string
	To   int64
	From int64
	Prop string
}
type Object interface {
	NodeName() string
	ID() int64
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
func (o *Obj) ID() int64 {
	return o.Attr(0).ToInt64(0)
}
func (o *Obj) Name() string {
	return strings.ReplaceAll(o.Attr(1).ToString(""), "\x00\x01", "::")
}
func (o *Obj) Kind() string {
	return o.Attr(2).ToString("")
}
func (o *Obj) GetProperty70(name string) *Property70 {
	if o.properties == nil {
		o.properties = map[string]*Property70{}
		for _, node := range o.FindChild("Properties70").GetChildren() {
			o.properties[node.Attr(0).ToString("")] = &Property70{
				AttributeList: node.Attributes[4:],
				Type:          node.Attr(1).ToString(""),
				Label:         node.Attr(2).ToString(""),
				Flag:          node.Attr(3).ToString("")}
		}
	}
	if p, ok := o.properties[name]; ok {
		return p
	} else if o.Template != nil {
		return o.Template.GetProperty70(name)
	}
	return &Property70{}
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
	Parent      *Model
}

func (m *Model) GetMatrix() *geom.Matrix4 {
	rotv := m.Rotation.Scale(math.Pi / 180)
	tr := geom.NewTranslateMatrix4(m.Translation.X, m.Translation.Y, m.Translation.Z)
	rot := geom.NewEulerRotationMatrix4(rotv.X, rotv.Y, rotv.Z, 1) // XYZ order?
	sacle := geom.NewScaleMatrix4(m.Scaling.X, m.Scaling.Y, m.Scaling.Z)
	return tr.Mul(rot).Mul(sacle)
}

func (m *Model) GetWorldMatrix() *geom.Matrix4 {
	if m.Parent == nil {
		return m.GetMatrix()
	}
	return m.Parent.GetWorldMatrix().Mul(m.GetMatrix())
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
