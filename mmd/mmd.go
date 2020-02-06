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

type Document struct {
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

func NewDocument() *Document {
	return &Document{Header: &Header{
		Format:  []byte("PMX "),
		Version: 2,
		Info:    []byte{1, 0, 2, 1, 1, 2, 1, 1},
	}}
}

type Header struct {
	Format  []byte
	Version float32
	Info    []byte
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

const (
	MaterialFlagDoubleSided uint8 = 1
	MaterialFlagCastShadow  uint8 = 2
)

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

const (
	BoneFlagTailIndex    uint16 = 1
	BoneFlagRotatable    uint16 = 2
	BoneFlagTranslatable uint16 = 4
	BoneFlagVisible      uint16 = 8
	BoneFlagEnabled      uint16 = 16
	BoneFlagEnableIK     uint16 = 32

	BoneFlagInheritRotation    uint16 = 256
	BoneFlagInheritTranslation uint16 = 512
	BoneFlagFixedAxis          uint16 = 1024
	BoneFlagLocalAxis          uint16 = 2048
	BoneFlagPhysicsMode        uint16 = 4096
	BoneFlagExternalParent     uint16 = 8192

	BoneFlagAll uint16 = (31 | 32 | 256 | 512 | 1024 | 2048 | 4096 | 8192)
)

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
