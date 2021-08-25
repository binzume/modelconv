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
	PropertyList
	Type  string
	Label string
	Flag  string
}

func (p *Property70) ToInt(defvalue int) int {
	return p.Get(0).ToInt(defvalue)
}

func (p *Property70) ToFloat32(defvalue float32) float32 {
	return p.Get(0).ToFloat32(defvalue)
}

func (p *Property70) ToString(defvalue string) string {
	return p.Get(0).ToString(defvalue)
}

func (p *Property70) ToVector3(x, y, z float32) geom.Vector3 {
	return geom.Vector3{X: p.Get(0).ToFloat32(x), Y: p.Get(1).ToFloat32(y), Z: p.Get(2).ToFloat32(z)}
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
	return o.Prop(0).ToInt64(0)
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
			o.properties[node.PropString(0)] = &Property70{
				PropertyList: node.Properties[4:],
				Type:         node.PropString(1),
				Label:        node.PropString(2),
				Flag:         node.PropString(3)}
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

func (m *Model) GetWorldMatrix() *geom.Matrix4 {
	if m == nil {
		return geom.NewMatrix4()
	}
	rotv := m.Rotation.Scale(math.Pi / 180)
	tr := geom.NewTranslateMatrix4(m.Translation.X, m.Translation.Y, m.Translation.Z)
	rot := geom.NewEulerRotationMatrix4(rotv.X, rotv.Y, rotv.Z, 1) // XYZ order?
	sacle := geom.NewScaleMatrix4(m.Scaling.X, m.Scaling.Y, m.Scaling.Z)
	return m.Parent.GetWorldMatrix().Mul(tr).Mul(rot).Mul(sacle)
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
