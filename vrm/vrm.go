package vrm

// https://vrm.dev/
// https://github.com/vrm-c/vrm-specification/blob/master/specification/0.0/README.ja.md

// ExporterVersion
var ExporterVersion = "modelconv-v0.2"

// RequiredBones for Humanoid
var RequiredBones = []string{
	"hips", "spine", "chest", "neck", "head",
	"leftUpperArm", "leftLowerArm", "leftHand",
	"rightUpperArm", "rightLowerArm", "rightHand",
	"leftUpperLeg", "leftLowerLeg", "leftFoot",
	"rightUpperLeg", "rightLowerLeg", "rightFoot",
}

type VRM struct {
	Meta     Metadata `json:"meta"`
	Humanoid Humanoid `json:"humanoid"`

	FirstPerson        FirstPerson         `json:"firstPerson"`
	BlendShapeMaster   BlendShapeMaster    `json:"blendShapeMaster"`
	SecondaryAnimation *SecondaryAnimation `json:"secondaryAnimation,omitempty"`
	MaterialProperties []*MaterialProperty `json:"materialProperties"`

	ExporterVersion string `json:"exporterVersion"`
}

func NewVRM() *VRM {
	return &VRM{MaterialProperties: []*MaterialProperty{}}
}

func (v *VRM) Title() string {
	return v.Meta.Title
}

func (v *VRM) Author() string {
	return v.Meta.Author
}

func (v *VRM) CheckRequiredBones() []string {
	bones := map[string]int{}
	for _, bone := range v.Humanoid.Bones {
		bones[bone.Bone] = bone.Node
	}
	var errorBones []string
	for _, name := range RequiredBones {
		if _, ok := bones[name]; !ok {
			errorBones = append(errorBones, name)
		}
	}
	return errorBones
}

type Metadata struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	Author  string `json:"author"`
	Contact string `json:"contactInformation"`
	Texture *int   `json:"texture,omitempty"`

	// TODO

	LicenseName     string `json:"licenseName"`
	OtherLicenseURL string `json:"otherLicenseUrl"`
}

type Humanoid struct {
	Bones []*Bone `json:"humanBones"`
}

type Bone struct {
	Bone             string `json:"bone"`
	Node             int    `json:"node"`
	UseDefaultValues bool   `json:"useDefaultValues"`

	Min        *[3]float32 `json:"min,omitempty"`
	Max        *[3]float32 `json:"max,omitempty"`
	Center     *[3]float32 `json:"center,omitempty"`
	AxisLength float32     `json:"axisLength,omitempty"`
}

type FirstPerson interface{} // TODO

type BlendShapeMaster struct {
	BlendShapeGroups []*BlendShapeGroup `json:"blendShapeGroups"`
}

type BlendShapeGroup struct {
	Name           string        `json:"name"`
	PresetName     string        `json:"presetName"`
	Binds          []interface{} `json:"binds"`
	MaterialValues []interface{} `json:"materialValues"`
}

type SecondaryAnimation struct {
	BoneGroups     []*SecondaryAnimationBoneGroup `json:"boneGroups"`
	ColliderGroups []interface{}                  `json:"colliderGroups"`
}

type SecondaryAnimationBoneGroup struct {
	Comment        string             `json:"comment"`
	Stiffiness     float64            `json:"stiffiness"`
	GravityPower   float64            `json:"gravityPower"`
	GravityDir     map[string]float64 `json:"gravityDir"`
	DragForce      float64            `json:"dragForce"`
	HitRadius      float64            `json:"hitRadius"`
	Center         int                `json:"center"`
	Bones          []int              `json:"bones"`
	ColliderGroups []int              `json:"colliderGroups"`
}

type MaterialProperty struct {
	Name              string                 `json:"name"`
	Shader            string                 `json:"shader"`
	RenderQueue       int                    `json:"renderQueue"`
	FloatProperties   map[string]float64     `json:"floatProperties"`
	VectorProperties  map[string]interface{} `json:"vectorProperties"`
	TextureProperties map[string]interface{} `json:"textureProperties"`
	KeywordMap        map[string]interface{} `json:"keywordMap"`
	TagMap            map[string]interface{} `json:"tagMap"`
}
