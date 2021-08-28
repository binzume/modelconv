package fbx

import (
	"math"
	"strings"

	"github.com/binzume/modelconv/geom"
)

type Document struct {
	FileId       []byte
	Creator      string
	CreationTime string

	GlobalSettings Object
	Objects        map[int64]Object
	Scene          *Model

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
	return strings.ReplaceAll(o.Attr(1).ToString(), "\x00\x01", "::")
}
func (o *Obj) Kind() string {
	return o.Attr(2).ToString()
}
func (o *Obj) GetProperty70(name string) *Property70 {
	if o.properties == nil {
		o.properties = map[string]*Property70{}
		for _, node := range o.FindChild("Properties70").GetChildren() {
			o.properties[node.Attr(0).ToString()] = &Property70{
				AttributeList: node.Attributes[4:],
				Type:          node.Attr(1).ToString(),
				Label:         node.Attr(2).ToString(),
				Flag:          node.Attr(3).ToString()}
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
}

func (g *Geometry) GetVertices() []*geom.Vector3 {
	return g.FindChild("Vertices").GetVec3Array()
}

func (g *Geometry) GetDeformers() []*Deformer {
	var r []*Deformer
	for _, deformer := range g.FindRefs("Deformer") {
		for _, sub := range deformer.FindRefs("Deformer") {
			if d, ok := sub.(*Deformer); ok {
				r = append(r, d)
			}
		}
	}
	return r
}

func (g *Geometry) GetShapes() []*GeometryShape {
	var shapes []*GeometryShape
	for _, node := range g.GetChildren() {
		if node.Name == "Shape" {
			shapes = append(shapes, &GeometryShape{node})
		}
	}
	return shapes
}

// TODO: rename to SubDeformer
type Deformer struct {
	Obj
}

func (d *Deformer) GetWeights() []float32 {
	return d.FindChild("Weights").GetFloat32Array()
}

func (d *Deformer) GetIndexes() []int32 {
	return d.FindChild("Indexes").GetInt32Array()
}

func (d *Deformer) GetTarget() *Model {
	nodes := d.FindRefs("Model")
	if len(nodes) == 0 {
		return nil
	}
	m, _ := nodes[0].(*Model)
	return m
}

type GeometryShape struct {
	*Node
}

func (s *GeometryShape) Name() string {
	return s.GetString()
}

func (s *GeometryShape) GetVertices() []*geom.Vector3 {
	return s.FindChild("Vertices").GetVec3Array()
}

func (s *GeometryShape) GetNormals() []*geom.Vector3 {
	return s.FindChild("Normals").GetVec3Array()
}

func (s *GeometryShape) GetIndexes() []int32 {
	return s.FindChild("Indexes").GetInt32Array()
}

type Material struct {
	Obj
}

func (m *Material) GetColor(name string, def *geom.Vector3) *geom.Vector3 {
	if def == nil {
		def = &geom.Vector3{}
	}
	return m.GetProperty70(name).ToVector3(def.X, def.Y, def.Z)
}

func (m *Material) GetFactor(name string, def float32) float32 {
	return m.GetProperty70(name).ToFloat32(def)
}

func (m *Material) GetTexture(name string) *Obj {
	// TOOD: m.GetPropertyRef(name)...
	textures := m.FindRefs("Texture")
	if len(textures) > 0 {
		return textures[0].(*Obj)
	}
	return nil
}

type Model struct {
	Obj
	Parent       *Model
	cachedMatrix *geom.Matrix4
}

func (m *Model) GetTranslation() *geom.Vector3 {
	return m.GetProperty70("Lcl Translation").ToVector3(0, 0, 0)
}

func (m *Model) GeRotation() *geom.Vector3 {
	return m.GetProperty70("Lcl Rotation").ToVector3(0, 0, 0) // Euler defgrees. XYZ order?
}

func (m *Model) GeScaling() *geom.Vector3 {
	return m.GetProperty70("Lcl Scaling").ToVector3(1, 1, 1)
}

func (m *Model) UpdateMatrix() {
	// TODO: apply pivot
	prerotEuler := m.GetProperty70("PreRotation").ToVector3(0, 0, 0).Scale(math.Pi / 180)
	prerot := geom.NewEulerRotationMatrix4(prerotEuler.X, prerotEuler.Y, prerotEuler.Z, 1)
	translation := m.GetTranslation()
	rotationEuler := m.GeRotation().Scale(math.Pi / 180)
	scale := m.GeScaling()
	tr := geom.NewTranslateMatrix4(translation.X, translation.Y, translation.Z)
	rot := geom.NewEulerRotationMatrix4(rotationEuler.X, rotationEuler.Y, rotationEuler.Z, 1)
	sacle := geom.NewScaleMatrix4(scale.X, scale.Y, scale.Z)
	m.cachedMatrix = tr.Mul(prerot).Mul(rot).Mul(sacle)
}

func (m *Model) GetMatrix() *geom.Matrix4 {
	if m.cachedMatrix == nil {
		m.UpdateMatrix()
	}
	return m.cachedMatrix
}

func (m *Model) GetWorldMatrix() *geom.Matrix4 {
	if m.Parent == nil {
		return m.GetMatrix()
	}
	return m.Parent.GetWorldMatrix().Mul(m.GetMatrix())
}

func (m *Model) GetChildModels() []*Model {
	var r []*Model
	for _, o := range m.FindRefs("Model") {
		if c, ok := o.(*Model); ok {
			r = append(r, c)
		}
	}
	return r
}

func (m *Model) GetGeometry() *Geometry {
	geometries := m.FindRefs("Geometry")
	if len(geometries) == 0 {
		return nil
	}
	g, _ := geometries[0].(*Geometry)
	return g
}
