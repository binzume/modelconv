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

type MQODocument struct {
	Scene     *Scene
	Materials []*Material
	Objects   []*Object

	Plugins []Plugin
}

func NewDocument() *MQODocument {
	return &MQODocument{}
}

func (mqo *MQODocument) GetPlugins() []Plugin {
	return mqo.Plugins
}

type Scene struct {
	CameraPos    Vector3
	CameraLookAt Vector3
	CameraRot    Vector3
}

type Material struct {
	Name  string
	Color Vector4

	Diffuse  float32
	Ambient  float32
	Emmition float32
	Specular float32
	Power    float32
	Texture  string

	DoubleSided bool

	Ex2 *MaterialEx2
}

type MaterialEx2 struct {
	ShaderType   string
	ShaderName   string
	ShaderParams map[string]interface{}
}

type Face struct {
	Verts    []int
	Material int
	UVs      []Vector2
}

type Object struct {
	Name     string
	Vertexes []*Vector3
	Faces    []*Face
	Visible  bool
	Locked   bool
	Depth    int
}

func NewObject(name string) *Object {
	return &Object{Name: name, Visible: true}
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

type Plugin interface {
	PreSerialize(mqo *MQODocument)
	PostDeserialize(mqo *MQODocument)
}
