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

func NewDocument() *Document {
	globalSettings := &Obj{Node: &Node{
		Name: "GlobalSettings",
		Children: []*Node{
			NewNode("Version", 1000),
			{Name: "Properties70"},
		},
	}}
	globalSettings.SetIntProperty("UpAxis", 1)
	globalSettings.SetIntProperty("UpAxisSign", 1)
	globalSettings.SetIntProperty("FrontAxis", 2)
	globalSettings.SetIntProperty("FrontAxisSign", 1)
	globalSettings.SetIntProperty("CoordAxis", 0)
	globalSettings.SetIntProperty("CoordAxisSign", 1)
	globalSettings.SetFloatProperty("UnitScaleFactor", 1.0)
	root := &Node{Children: []*Node{
		{Name: "FBXHeaderExtension", Children: []*Node{
			NewNode("FBXHeaderVersion", 1003),
			NewNode("FBXVersion", 7500),
			NewNode("Creator", "modelconv"),
		}},
		globalSettings.Node,
		{Name: "Definitions"},
		{Name: "Objects"},
		{Name: "Connections"},
		{Name: "Takes"},
	}}

	doc, _ := BuildDocument(root)
	return doc
}

func (doc *Document) AddObject(obj Object) int64 {
	if obj.ID() == 0 || doc.Objects[obj.ID()] != nil {
		var maxId int64
		for id := range doc.Objects {
			if id > maxId {
				maxId = id
			}
		}
		id := maxId + 1
		obj.GetNode().Attributes[0] = &Attribute{Value: id}

	}
	doc.Objects[obj.ID()] = obj
	objs := doc.RawNode.FindChild("Objects")
	objs.Children = append(objs.Children, obj.GetNode())
	return obj.ID()
}

func (doc *Document) AddConnection(parent, child Object) {
	parent.AddRef(child)
	conns := doc.RawNode.FindChild("Connections")
	conns.Children = append(conns.Children, NewNode("C", "OO", child.ID(), parent.ID()))
}

type Property struct {
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
	GetNode() *Node
	NodeName() string
	ID() int64
	Name() string
	Kind() string
	GetProperty(name string) *Property
	SetProperty(name string, prop *Property) *Property
	FindRefs(name string) []Object
	AddRef(o Object)
}

type Obj struct {
	*Node
	Template   *Obj
	Refs       []Object
	properties map[string]*Property // lazty initialize
}

func newObj(typ, name, kind string, nodes []*Node) *Obj {
	children := append(nodes, &Node{Name: "Properties70"})
	obj := &Obj{Node: &Node{
		Name:       typ,
		Attributes: []*Attribute{{Value: 0}, {Value: name}, {Value: kind}},
		Children:   children,
	}}
	return obj
}

func (o *Obj) GetNode() *Node {
	return o.Node
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
func (o *Obj) GetProperty(name string) *Property {
	if o.properties == nil {
		o.properties = map[string]*Property{}
		for _, node := range o.FindChild("Properties70").GetChildren() {
			o.properties[node.Attr(0).ToString()] = &Property{
				AttributeList: node.Attributes[4:],
				Type:          node.Attr(1).ToString(),
				Label:         node.Attr(2).ToString(),
				Flag:          node.Attr(3).ToString()}
		}
	}
	if p, ok := o.properties[name]; ok {
		return p
	} else if o.Template != nil {
		return o.Template.GetProperty(name)
	}
	return &Property{}
}
func (o *Obj) SetProperty(name string, prop *Property) *Property {
	if o.properties != nil {
		o.properties[name] = prop
	}
	attrs := AttributeList{
		&Attribute{Value: name},
		&Attribute{Value: prop.Type},
		&Attribute{Value: prop.Label},
		&Attribute{Value: prop.Flag},
	}
	attrs = append(attrs, prop.AttributeList...)
	properties70 := o.FindChild("Properties70")
	for _, node := range properties70.GetChildren() {
		if node.Attr(0).ToString() == name {
			node.Attributes = attrs
			return prop
		}
	}
	properties70.Children = append(properties70.Children, &Node{Name: "P", Attributes: attrs})
	return prop
}

func (o *Obj) SetIntProperty(name string, v int) *Property {
	return o.SetProperty(name, &Property{Type: "int", Label: "Integer", AttributeList: []*Attribute{{Value: int64(v)}}})
}

func (o *Obj) SetFloatProperty(name string, v float64) *Property {
	return o.SetProperty(name, &Property{Type: "double", Label: "Number", AttributeList: []*Attribute{{Value: v}}})
}

func (o *Obj) SetStringProperty(name string, v string) *Property {
	return o.SetProperty(name, &Property{Type: "KString", Label: "", AttributeList: []*Attribute{{Value: v}}})
}

func (o *Obj) SetColorProperty(name string, r, g, b float32) *Property {
	return o.SetProperty(name, &Property{Type: "ColorRGB", Label: "Color", AttributeList: []*Attribute{{Value: r}, {Value: g}, {Value: b}}})
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
	Vertices           []*geom.Vector3
	Polygons           [][]int
	PolygonVertexCount int
}

func NewGeometry(name string, verts []*geom.Vector3, faces [][]int) *Geometry {
	var varray []float32
	for _, v := range verts {
		varray = append(varray, v.X, v.Y, v.Z)
	}
	var indices []int32
	for _, f := range faces {
		for _, i := range f {
			indices = append(indices, int32(i))
		}
		if len(f) > 0 {
			indices[len(indices)-1] = ^indices[len(indices)-1]
		}
	}

	geom := &Geometry{
		Obj:      *newObj("Geometry", name+"\x00\x01Geometry", "Mesh", nil),
		Vertices: verts,
		Polygons: faces,
	}
	geom.Children = append(geom.Children,
		NewNode("Version", 124),
		NewNode("Vertices", varray),
		NewNode("PolygonVertexIndex", indices),
	)
	return geom
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

func (g *Geometry) GetLayerElement(name string, arrayName string, indexName string) *LayerElement {
	node := g.FindChild(name)
	return &LayerElement{node, node.FindChild(arrayName), node.FindChild(indexName)}
}

func (g *Geometry) GetLayerElementUV() *LayerElement {
	return g.GetLayerElement("LayerElementUV", "UV", "UVIndex")
}

func (g *Geometry) GetLayerElementMaterial() *LayerElement {
	return g.GetLayerElement("LayerElementMaterial", "Materials", "Materials") // Materials as index?
}

func (g *Geometry) GetLayerElementNormal() *LayerElement {
	return g.GetLayerElement("LayerElementNormal", "Normals", "NormalsIndex")
}

func (g *Geometry) GetLayerElementBinormal() *LayerElement {
	return g.GetLayerElement("LayerElementBinormal", "Binormals", "BinormalsIndex")
}

func (g *Geometry) GetLayerElementTangent() *LayerElement {
	return g.GetLayerElement("LayerElementTangent", "Tangents", "TangentsIndex")
}

func (g *Geometry) GetLayerElementSmoothing() *LayerElement {
	return g.GetLayerElement("LayerElementSmoothing", "Smoothing", "SmoothingIndex")
}

type LayerElement struct {
	*Node
	Array     *Node
	IndexNode *Node
}

func (e *LayerElement) GetMappingInformationType() string {
	return e.FindChild("MappingInformationType").GetString()
}

func (e *LayerElement) GetReferenceInformationType() string {
	return e.FindChild("ReferenceInformationType").GetString()
}

func (e *LayerElement) GetIndexes() []int32 {
	return e.IndexNode.GetInt32Array()
}

type GeometryShape struct {
	*Node
}

func (s *GeometryShape) Name() string {
	if len(s.Attributes) >= 3 {
		return strings.ReplaceAll(s.Attr(1).ToString(), "\x00\x01", "::")
	}
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

func (d *Deformer) GetTransform() *geom.Matrix4 {
	m := d.FindChild("Transform").GetFloat32Array()
	if len(m) != 16 {
		return geom.NewMatrix4()
	}
	return geom.NewMatrix4FromSlice(m)
}

func (d *Deformer) GetTransformLink() *geom.Matrix4 {
	m := d.FindChild("TransformLink").GetFloat32Array()
	if len(m) != 16 {
		return geom.NewMatrix4()
	}
	return geom.NewMatrix4FromSlice(m)
}

func (d *Deformer) GetTarget() *Model {
	nodes := d.FindRefs("Model")
	if len(nodes) == 0 {
		return nil
	}
	m, _ := nodes[0].(*Model)
	return m
}

func (d *Deformer) GetShapes() []*GeometryShape {
	var shapes []*GeometryShape
	for _, node := range d.FindRefs("Geometry") {
		if node.Kind() == "Shape" {
			shapes = append(shapes, &GeometryShape{Node: node.(*Geometry).Node})
		}
	}
	return shapes
}

type Material struct {
	Obj
}

func NewMaterial(name string) *Material {
	mat := &Material{
		Obj: *newObj("Material", name+"\x00\x01Material", "", []*Node{
			NewNode("Version", 102),
			NewNode("ShadingModel", "phong"),
		}),
	}
	mat.SetStringProperty("ShadingModel", "phong")
	return mat
}

func (m *Material) GetColor(name string, def *geom.Vector3) *geom.Vector3 {
	if def == nil {
		def = &geom.Vector3{}
	}
	return m.GetProperty(name).ToVector3(def.X, def.Y, def.Z)
}

func (m *Material) GetFactor(name string, def float32) float32 {
	return m.GetProperty(name).ToFloat32(def)
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

func NewModel(name string) *Model {
	model := &Model{
		Obj: *newObj("Model", name+"\x00\x01Model", "Null", []*Node{
			NewNode("Version", 232),
		}),
	}
	model.SetStringProperty("Culling", "CullingOff")
	model.SetStringProperty("Culling", "CullingOff")
	return model
}

func (m *Model) GetTranslation() *geom.Vector3 {
	return m.GetProperty("Lcl Translation").ToVector3(0, 0, 0)
}

func (m *Model) SetTranslation(v *geom.Vector3) {
	m.SetProperty("Lcl Translation", &Property{Type: "Lcl Translation", Flag: "A+", AttributeList: []*Attribute{{Value: v.X}, {Value: v.Y}, {Value: v.Z}}})
}

func (m *Model) GetRotation() *geom.Vector3 {
	return m.GetProperty("Lcl Rotation").ToVector3(0, 0, 0) // Euler defgrees. XYZ order?
}

func (m *Model) SetRotation(v *geom.Vector3) {
	m.SetProperty("Lcl Rotation", &Property{Type: "Lcl Rotation", Flag: "A+", AttributeList: []*Attribute{{Value: v.X}, {Value: v.Y}, {Value: v.Z}}})
}

func (m *Model) GetScaling() *geom.Vector3 {
	return m.GetProperty("Lcl Scaling").ToVector3(1, 1, 1)
}

func (m *Model) SetScaling(v *geom.Vector3) {
	m.SetProperty("Lcl Scaling", &Property{Type: "Lcl Scaling", Flag: "A+", AttributeList: []*Attribute{{Value: v.X}, {Value: v.Y}, {Value: v.Z}}})
}

func (m *Model) UpdateMatrix() {
	// TODO: apply pivot
	prerotEuler := m.GetProperty("PreRotation").ToVector3(0, 0, 0).Scale(math.Pi / 180)
	prerot := geom.NewEulerRotationMatrix4(prerotEuler.X, prerotEuler.Y, prerotEuler.Z, 1)
	translation := m.GetTranslation()
	rotationEuler := m.GetRotation().Scale(math.Pi / 180)
	scale := m.GetScaling()
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
