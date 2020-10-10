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

	FirstPerson        *FirstPerson        `json:"firstPerson,omitempty"`
	BlendShapeMaster   BlendShapeMaster    `json:"blendShapeMaster,omitempty"`
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

	AllowedUserName      string `json:"allowedUserName,omitempty"`
	ViolentUssageName    string `json:"violentUssageName,omitempty"`
	SexualUssageName     string `json:"sexualUssageName,omitempty"`
	CommercialUssageName string `json:"commercialUssageName,omitempty"`

	OtherPermissionURL string `json:"otherPermissionUrl,omitempty"`

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

type FirstPerson struct {
	FirstPersonBone       int                `json:"firstPersonBone,omitempty"`
	FirstPersonBoneOffset map[string]float64 `json:"firstPersonBoneOffset,omitempty"`

	MeshAnnotations []interface{} `json:"meshAnnotations,omitempty"`

	LookAtTypeName        string                 `json:"lookAtTypeName,omitempty"`
	LookAtHorizontalInner map[string]interface{} `json:"lookAtHorizontalInner,omitempty"`
	LookAtHorizontalOuter map[string]interface{} `json:"lookAtHorizontalOuter,omitempty"`
	LookAtVerticalDown    map[string]interface{} `json:"lookAtVerticalDown,omitempty"`
	LookAtVerticalUp      map[string]interface{} `json:"lookAtVerticalUp,omitempty"`
}

type BlendShapeMaster struct {
	BlendShapeGroups []*BlendShapeGroup `json:"blendShapeGroups"`
}

type BlendShapeGroup struct {
	Name           string                     `json:"name"`
	PresetName     string                     `json:"presetName"`
	Binds          []*BlendShapeBind          `json:"binds"`
	MaterialValues []*BlendShapeMaterialValue `json:"materialValues,omitempty"`
}

type BlendShapeBind struct {
	Mesh   uint32  `json:"mesh"`
	Index  int     `json:"index"`
	Weight float64 `json:"weight"`
}

type BlendShapeMaterialValue struct {
	MaterialName string    `json:"materialName"`
	PropertyName string    `json:"propertyName"`
	TargetValue  []float64 `json:"targetValue"`
}

type SecondaryAnimation struct {
	BoneGroups     []*SecondaryAnimationBoneGroup     `json:"boneGroups,omitempty"`
	ColliderGroups []*SecondaryAnimationColliderGroup `json:"colliderGroups,omitempty"`
}

type SecondaryAnimationBoneGroup struct {
	Comment        string             `json:"comment"`
	Stiffiness     float64            `json:"stiffiness"`
	GravityPower   float64            `json:"gravityPower"`
	GravityDir     map[string]float64 `json:"gravityDir,omitempty"`
	DragForce      float64            `json:"dragForce"`
	HitRadius      float64            `json:"hitRadius"`
	Center         int                `json:"center"`
	Bones          []int              `json:"bones"`
	ColliderGroups []int              `json:"colliderGroups,omitempty"`
}

type SecondaryAnimationColliderGroup struct {
	Node      uint32                        `json:"node"`
	Colliders []*SecondaryAnimationCollider `json:"colliders"`
}

type SecondaryAnimationCollider struct {
	Radius float64            `json:"radius"`
	Offset map[string]float64 `json:"offset"`
}

type MaterialProperty struct {
	Name              string               `json:"name"`
	Shader            string               `json:"shader"`
	RenderQueue       int                  `json:"renderQueue"`
	FloatProperties   map[string]float64   `json:"floatProperties"`
	VectorProperties  map[string][]float64 `json:"vectorProperties"`
	TextureProperties map[string]uint32    `json:"textureProperties"`
	KeywordMap        map[string]bool      `json:"keywordMap"`
	TagMap            map[string]string    `json:"tagMap"`
}

func NewMaterialProperty(name string) *MaterialProperty {
	return &MaterialProperty{
		Name:              name,
		RenderQueue:       2000,
		FloatProperties:   map[string]float64{},
		VectorProperties:  map[string][]float64{},
		TextureProperties: map[string]uint32{},
		KeywordMap:        map[string]bool{},
		TagMap:            map[string]string{},
	}
}
