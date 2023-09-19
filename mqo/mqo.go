package mqo

import (
	"fmt"
	"math"

	"github.com/binzume/modelconv/geom"
)

type Vector2 = geom.Vector2
type Vector3 = geom.Vector3
type Vector4 = geom.Vector4
type Matrix4 = geom.Matrix4

type Document struct {
	Scene     *Scene
	Materials []*Material
	Objects   []*Object

	Plugins []Plugin
}

func NewDocument() *Document {
	return &Document{}
}

func (doc *Document) GetPlugins() []Plugin {
	return doc.Plugins
}

func (doc *Document) FixObjectID() {
	for i, obj := range doc.Objects {
		if obj.UID == 0 {
			obj.UID = i + 1 // TODO: unique id
		}
	}
}

func (doc *Document) GetObjectByID(id int) *Object {
	for _, obj := range doc.Objects {
		if obj.UID == id {
			return obj
		}
	}
	return nil
}

func (doc *Document) GetWorldTransforms() map[*Object]*Matrix4 {
	transforms := map[*Object]*Matrix4{}
	dd := []int{-1}
	dt := []*Matrix4{geom.NewMatrix4()}

	for _, o := range doc.Objects {
		for len(dd) > 1 && dd[len(dd)-1] <= o.Depth {
			dd = dd[:len(dd)-1]
			dt = dt[:len(dt)-1]
		}
		transforms[o] = dt[len(dt)-1].Mul(o.GetLocalTransform())
		dd = append(dd, o.Depth)
		dt = append(dt, transforms[o])
	}
	return transforms
}

func (doc *Document) FixNames() {
	objNames := map[string]bool{}
	for _, obj := range doc.Objects {
		if objNames[obj.Name] {
			n := 2
			for objNames[fmt.Sprintf("%s_%d", obj.Name, n)] {
				n++
			}
			obj.Name = fmt.Sprintf("%s_%d", obj.Name, n)
		}
		objNames[obj.Name] = true
	}
	matNames := map[string]bool{}
	for _, mat := range doc.Materials {
		if matNames[mat.Name] {
			n := 2
			for matNames[fmt.Sprintf("%s_%d", mat.Name, n)] {
				n++
			}
			mat.Name = fmt.Sprintf("%s_%d", mat.Name, n)
		}
		matNames[mat.Name] = true
	}
}

type Scene struct {
	CameraPos    Vector3
	CameraLookAt Vector3
	CameraRot    Vector3
	Zoom2        float32
	FrontClip    float32
	BackClip     float32
	Ortho        bool

	AmbientLight *Vector3
}

type Material struct {
	Name  string
	UID   int
	Color Vector4

	Diffuse  float32
	Ambient  float32
	Emission float32
	Specular float32
	Power    float32

	Texture      string
	BumpTexture  string
	AlphaTexture string

	EmissionColor *Vector3

	DoubleSided bool

	Shader int
	Ex2    *MaterialEx2
}

const (
	ShaderClassic  = 0
	ShaderConstant = 1
	ShaderLambert  = 2
	ShaderPhong    = 3
	ShaderBlinn    = 4
	ShaderTypeHLSL = "hlsl"
	ShaderNameGlTF = "glTF"
)

func (m *Material) GetShaderName() string {
	if m.Ex2 != nil {
		return m.Ex2.ShaderName
	}
	switch m.Shader {
	case ShaderClassic:
		return "Classic"
	case ShaderConstant:
		return "Constant"
	case ShaderLambert:
		return "Lambert"
	case ShaderPhong:
		return "Phong"
	case ShaderBlinn:
		return "Blinn"
	}
	return ""
}

const (
	AlphaModeOpaque = 1
	AlphaModeMask   = 2
	AlphaModeBlend  = 3
)

func (m *Material) SetGltfAlphaMode(mode int) {
	if m.Ex2 == nil || m.Ex2.ShaderName != ShaderNameGlTF {
		m.Ex2 = NewMaterialEx2(ShaderNameGlTF)
	}
	m.Ex2.ShaderParams["AlphaMode"] = mode
}

type MaterialEx2 struct {
	ShaderType          string
	ShaderName          string
	ShaderParams        map[string]interface{}
	ShaderMapping       map[string]string
	ShaderMappingParams map[string]map[string]interface{}
}

func NewMaterialEx2(name string) *MaterialEx2 {
	return &MaterialEx2{
		ShaderName:          name,
		ShaderParams:        map[string]interface{}{},
		ShaderMapping:       map[string]string{},
		ShaderMappingParams: map[string]map[string]interface{}{},
	}
}

func (m *MaterialEx2) StringParam(name string) string {
	if v, ok := m.ShaderParams[name].(string); ok {
		return v
	}
	return ""
}

func (m *MaterialEx2) IntParam(name string) int {
	if v, ok := m.ShaderParams[name].(int); ok {
		return v
	}
	return 0
}

func (m *MaterialEx2) FloatParam(name string) float64 {
	if v, ok := m.ShaderParams[name].(float64); ok {
		return v
	}
	if v, ok := m.ShaderParams[name].(float32); ok {
		return float64(v)
	}
	return 0
}

func (m *MaterialEx2) BoolParam(name string) bool {
	if v, ok := m.ShaderParams[name].(bool); ok {
		return v
	}
	return false
}

func (m *MaterialEx2) ColorParam(name string) []float32 {
	if v, ok := m.ShaderParams[name].([]float32); ok {
		return v
	}
	return nil
}

func (m *MaterialEx2) Mapping(name string) string {
	if v, ok := m.ShaderMapping[name]; ok {
		return v
	}
	return ""
}

type Face struct {
	UID      int
	Verts    []int
	Material int
	UVs      []Vector2
	Normals  []*Vector3
}

func (f *Face) Flip() {
	for i, j := 0, len(f.Verts)-1; i < j; i, j = i+1, j-1 {
		f.Verts[i], f.Verts[j] = f.Verts[j], f.Verts[i]
	}
	for i, j := 0, len(f.UVs)-1; i < j; i, j = i+1, j-1 {
		f.UVs[i], f.UVs[j] = f.UVs[j], f.UVs[i]
	}
	for i, j := 0, len(f.Normals)-1; i < j; i, j = i+1, j-1 {
		f.Normals[i], f.Normals[j] = f.Normals[j], f.Normals[i]
	}
}

type Object struct {
	UID          int
	Name         string
	Vertexes     []*Vector3
	Faces        []*Face
	Visible      bool
	Locked       bool
	Depth        int
	Folding      bool
	Shading      int
	Facet        float32
	Patch        int
	PatchSegment int
	Mirror       int
	MirrorDis    float32

	Scale       *Vector3
	Rotation    *Vector3 // TODO: EulerAngles
	Translation *Vector3

	Color *Vector3

	VertexByUID map[int]int

	// Internal use
	SharedGeometryHint *SharedGeometryHint
	Extra              map[string]interface{}
}

type SharedGeometryHint struct {
	Key       string
	Transform *geom.Matrix4
}

func NewObject(name string) *Object {
	return &Object{
		Name: name, Visible: true, VertexByUID: map[int]int{}, Shading: 1, Facet: 59.5,
		Rotation:    geom.NewVector3(0, 0, 0),
		Scale:       geom.NewVector3(1, 1, 1),
		Translation: geom.NewVector3(0, 0, 0),
		Extra:       map[string]interface{}{},
	}
}

func (o *Object) SetRotation(r *geom.Quaternion) {
	// RotationOrderYXZ?
	o.Rotation = geom.NewEulerFromQuaternion(r, geom.RotationOrderZXY).Vector3.Scale(180 / math.Pi)
}

func (o *Object) GetRotation() *geom.Quaternion {
	return (&geom.EulerAngles{Vector3: *o.Rotation.Scale(math.Pi / 180), Order: geom.RotationOrderYXZ}).ToQuaternion()
}

func (o *Object) GetLocalTransform() *geom.Matrix4 {
	return geom.NewTRSMatrix4(o.Translation, o.GetRotation(), o.Scale)
}

func (o *Object) SetLocalTransform(mat *geom.Matrix4) {
	t, r, s := mat.Decompose()
	o.Translation = t
	o.Scale = s
	o.SetRotation(r)
}

func (o *Object) Clone() *Object {
	var cp = *o
	cp.Vertexes = make([]*Vector3, len(o.Vertexes))
	for i, v := range o.Vertexes {
		vv := *v
		cp.Vertexes[i] = &vv
	}
	cp.Faces = make([]*Face, len(o.Faces))
	for i, v := range o.Faces {
		d := &Face{Material: v.Material, Verts: make([]int, len(v.Verts)), UVs: make([]Vector2, len(v.UVs))}
		copy(d.Verts, v.Verts)
		copy(d.UVs, v.UVs)
		cp.Faces[i] = d
	}
	return &cp
}

func (o *Object) GetVertexIndexByID(uid int) int {
	if v, ok := o.VertexByUID[uid]; ok {
		return v
	}
	if len(o.Vertexes) >= uid {
		return uid - 1
	}
	return -1
}

type Plugin interface {
	PreSerialize(mqo *Document)
	PostDeserialize(mqo *Document)
}
