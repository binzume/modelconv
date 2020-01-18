package vrm

// https://vrm.dev/
// https://github.com/vrm-c/vrm-specification/blob/master/specification/0.0/README.ja.md

import (
	"encoding/json"

	"github.com/qmuntal/gltf"
)

const (
	ExtensionName = "VRM"
)

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

type Metadata struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	Author  string `json:"author"`

	// TODO

	LicenseName     string `json:"licenseName"`
	OtherLicenseUrl string `json:"otherLicenseUrl"`
}

type Bone struct {
	Bone             string  `json:"bone"`
	Node             int     `json:"node"`
	UseDefaultValues bool    `json:"useDefaultValues"`
	AxisLength       float32 `json:"axisLength"`
}

type Humanoid struct {
	Bones []*Bone `json:"humanBones"`
}

type VRMExt struct {
	Meta     Metadata `json:"meta"`
	Humanoid Humanoid `json:"humanoid"`

	// TODO
	FirstPerson        interface{} `json:"firstPerson"`
	BlendShapeMaster   interface{} `json:"blendShapeMaster"`
	SecondaryAnimation interface{} `json:"secondaryAnimation"`
	MaterialProperties interface{} `json:"materialProperties"`

	ExporterVersion string `json:"exporterVersion"`
}

func Unmarshal(data []byte) (interface{}, error) {
	var vrmext VRMExt
	if err := json.Unmarshal([]byte(data), &vrmext); err != nil {
		return nil, err
	}
	return &vrmext, nil
}
