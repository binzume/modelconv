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

type Scene struct {
}

type MQODocument struct {
	Scene     Scene
	Materials []*Material
	Objects   []*Object

	// Plugins
	Bones  []*Bone
	Morphs []*MorphTargetList
}
