package vrm

// https://vrm.dev/
// https://github.com/vrm-c/vrm-specification/blob/master/specification/0.0/README.ja.md

import (
	"encoding/json"
	"fmt"

	"github.com/qmuntal/gltf"
)

// ExporterVersion
var ExporterVersion = "modelconv-v0.0.1"

// RequiredBones for Humanoid
var RequiredBones = []string{
	"hips", "spine", "chest", "neck", "head",
	"leftUpperArm", "leftLowerArm", "leftHand",
	"rightUpperArm", "rightLowerArm", "rightHand",
	"leftUpperLeg", "leftLowerLeg", "leftFoot",
	"rightUpperLeg", "rightLowerLeg", "rightFoot",
}

const ExtensionName = "VRM"

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

type VRMExtension struct {
	Meta     Metadata `json:"meta"`
	Humanoid Humanoid `json:"humanoid"`

	FirstPerson        FirstPerson         `json:"firstPerson"`
	BlendShapeMaster   BlendShapeMaster    `json:"blendShapeMaster"`
	SecondaryAnimation *SecondaryAnimation `json:"secondaryAnimation,omitempty"`
	MaterialProperties []*MaterialProperty `json:"materialProperties"`

	ExporterVersion string `json:"exporterVersion"`
}

func NewVRMExtension() *VRMExtension {
	return &VRMExtension{MaterialProperties: []*MaterialProperty{}}
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

func Unmarshal(data []byte) (interface{}, error) {
	var vrmext VRMExtension
	if err := json.Unmarshal([]byte(data), &vrmext); err != nil {
		return nil, err
	}
	return &vrmext, nil
}

type Document gltf.Document

func (doc *Document) VRM() *VRMExtension {
	if ext, ok := doc.Extensions[ExtensionName].(*VRMExtension); ok {
		return ext
	}
	ext := NewVRMExtension()
	if doc.Extensions == nil {
		doc.Extensions = gltf.Extensions{}
	}
	doc.Extensions[ExtensionName] = ext
	if !doc.IsExtentionUsed(ExtensionName) {
		doc.ExtensionsUsed = append(doc.ExtensionsUsed, ExtensionName)
	}
	return ext
}

func (doc *Document) Title() string {
	return doc.VRM().Meta.Title
}

func (doc *Document) Author() string {
	return doc.VRM().Meta.Author
}

func (doc *Document) IsExtentionUsed(extname string) bool {
	for _, ex := range doc.ExtensionsUsed {
		if ex == extname {
			return true
		}
	}
	return false
}

func (doc *Document) ValidateBones() error {
	ext := doc.VRM()

	bones := map[string]int{}
	for _, bone := range ext.Humanoid.Bones {
		bones[bone.Bone] = bone.Node
	}

	boneErrors := []string{}
	for _, name := range RequiredBones {
		if _, ok := bones[name]; !ok {
			boneErrors = append(boneErrors, fmt.Sprintf("%v not found.", name))
		}
	}
	if len(boneErrors) > 0 {
		return fmt.Errorf("Bone error: %v", boneErrors)
	}
	return nil
}
