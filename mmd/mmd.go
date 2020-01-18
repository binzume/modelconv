package mmd

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

type Vertex struct {
	Pos       Vector3
	Normal    Vector3
	UV        Vector2
	ExtUVs    []Vector4
	EdgeScale float32

	// TODO Matrix
	Bones       []int
	BoneWeights []float32
}

type Face struct {
	Verts [3]int
}

type Header struct {
	Format   []byte
	Version  float32
	InfoLen  uint8
	Info     []byte
	Encoding uint8
}

type Material struct {
	Name        string
	NameEn      string
	Color       Vector4
	Specular    Vector3
	Specularity float32
	AColor      Vector3
	Flags       byte
	EdgeColor   Vector4
	EdgeScale   float32
	TextureID   int
	EnvID       int
	EnvMode     byte
	ToonType    byte
	Toon        int
	Memo        string
	Count       int
}

type Link struct {
	TargetID int
	HasLimit bool
	LimitMax Vector3
	LimitMin Vector3
}

type Bone struct {
	Name     string
	NameEn   string
	Pos      Vector3
	ParentID int
	Layer    int
	Flags    uint16
	TailID   int
	TailPos  Vector3

	InheritParentID        int
	InheritParentInfluence float32

	FixedAxis Vector3

	IK struct {
		TargetID int
		Loop     int
		LimitRad float32
		Links    []*Link
	}
}

// type 0
type MorphGroup struct {
	Target int
	Weight float32
}

// type 1
type MorphVertex struct {
	Target int
	Offset Vector3
}

// type 3
type MorphUV struct {
	Target int
	Value  Vector4
}

// type 8
type MorphMaterial struct {
	Target int

	Flags           byte
	Diffuse         Vector4
	Specular        Vector3
	Specularity     float32
	Ambient         Vector3
	EdgeColor       Vector4
	EdgeSize        float32
	TextureTint     Vector4
	EnvironmentTint Vector4
	ToonTint        Vector4
}

type Morph struct {
	Name      string
	NameEn    string
	PanelType byte
	MorphType byte

	// oneof
	Group    []*MorphGroup
	Vertex   []*MorphVertex
	UV       []*MorphUV
	Material []*MorphMaterial
}

type PMXDocument struct {
	Header    *Header
	Name      string
	NameEn    string
	Comment   string
	CommentEn string
	Vertexes  []*Vertex
	Faces     []*Face
	Textures  []string
	Materials []*Material
	Bones     []*Bone
	Morphs    []*Morph
}

const (
	AttrStringEncoding int = iota
	AttrExtUV
	AttrVertIndexSz
	AttrTexIndexSz
	AttrMatIndexSz
	AttrBoneIndexSz
	AttrMorphIndexSz
	AttrRBIndexSz
)
