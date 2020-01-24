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

type Scene struct {
}

type Plugin interface {
	PreSerialize(mqo *MQODocument)
}

type MQODocument struct {
	Scene     Scene
	Materials []*Material
	Objects   []*Object

	// Plugins
	// TODO: replace with []Plugin
	Bones  []*Bone
	Morphs []*MorphTargetList
}

func (mqo *MQODocument) GetPlugins() []Plugin {
	var plugins []Plugin
	if len(mqo.Bones) > 0 {
		plugins = append(plugins, &BonePlugin{
			BoneSet2: BoneSet2{Bones: mqo.Bones, Limit: 4},
		})
	}
	if len(mqo.Morphs) > 0 {
		plugins = append(plugins, &MorphPlugin{
			MorphSet: MorphSet{mqo.Morphs},
		})
	}
	return plugins
}
