package mqo

type Vector2 struct {
	X float32
	Y float32
}

type Vector3 struct {
	X float32
	Y float32
	Z float32
}

type Vector4 struct {
	X float32
	Y float32
	Z float32
	W float32
}

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

type Scene struct {
	CameraPos    Vector3
	CameraLookAt Vector3
	CameraRot    Vector3
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
	Texture  string

	EmissionColor *Vector3

	DoubleSided bool

	Ex2 *MaterialEx2
}

type MaterialEx2 struct {
	ShaderType   string
	ShaderName   string
	ShaderParams map[string]interface{}
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
	return 0
}

type Face struct {
	UID      int
	Verts    []int
	Material int
	UVs      []Vector2
}

type Object struct {
	UID         int
	Name        string
	Vertexes    []*Vector3
	Faces       []*Face
	Visible     bool
	Locked      bool
	Depth       int
	VertexByUID map[int]int
}

func NewObject(name string) *Object {
	return &Object{Name: name, Visible: true, VertexByUID: map[int]int{}}
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
