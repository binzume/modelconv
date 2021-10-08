package fbx

import (
	"strings"

	"github.com/binzume/modelconv/geom"
)

type Geometry struct {
	Obj
	Vertices           []*geom.Vector3
	Polygons           [][]int
	PolygonVertexCount int
}

type MappingType string

const (
	AllSame         MappingType = "AllSame"
	ByPolygon       MappingType = "ByPolygon"
	ByVertice       MappingType = "ByVertice"
	ByPolygonVertex MappingType = "ByPolygonVertex"
	ByControlPoint  MappingType = "ByControlPoint"
)

type LayerElement struct {
	*Node
	Array     *Node
	IndexNode *Node
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
		Obj: *newObj("Geometry", name+"\x00\x01Geometry", "Mesh", []*Node{
			NewNode("Vertices", varray),
			NewNode("PolygonVertexIndex", indices),
			NewNode("Layer", 0),
		}),
		Vertices: verts,
		Polygons: faces,
	}
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

func (g *Geometry) setLayerElementNode(name string, typeIndex int, values []*Node) {
	node := NewNode(name, 0)
	if g.AddOrReplaceChild(node) {
		layer := g.FindChild("Layer")
		el := NewNode("LayerElement")
		el.Children = []*Node{
			NewNode("Type", name),
			NewNode("TypedIndex", typeIndex),
		}
		layer.Children = append(layer.Children, el)
	}
}

func (g *Geometry) SetLayerElementMaterialIndex(mat []int32, mappingType MappingType) {
	g.setLayerElementNode("LayerElementMaterial", 0, []*Node{
		NewNode("MappingInformationType", string(mappingType)),
		NewNode("ReferenceInformationType", "IndexToDirect"),
		NewNode("Materials", mat),
	})
}

func (g *Geometry) SetLayerElementUVIndexed(uv []*geom.Vector2, indices []int32, mappingType MappingType) {
	var floatArray []float32
	for _, v := range uv {
		floatArray = append(floatArray, v.X, v.Y)
	}
	g.setLayerElementNode("LayerElementNormal", 0, []*Node{
		NewNode("MappingInformationType", string(mappingType)),
		NewNode("ReferenceInformationType", "IndexToDirect"),
		NewNode("UV", floatArray),
		NewNode("UVIndex", indices),
	})
}

func (g *Geometry) SetLayerElementNormal(normals []*geom.Vector3, mappingType MappingType) {
	var floatArray []float32
	for _, v := range normals {
		floatArray = append(floatArray, v.X, v.Y, v.Z)
	}
	g.setLayerElementNode("LayerElementNormal", 0, []*Node{
		NewNode("MappingInformationType", string(mappingType)),
		NewNode("ReferenceInformationType", "Direct"),
		NewNode("Normals", floatArray),
	})
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

func (e *LayerElement) GetMappingInformationType() MappingType {
	return MappingType(e.FindChild("MappingInformationType").GetString())
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
