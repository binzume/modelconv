package vrm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type Config struct {
	Metadata Metadata `json:"meta"`

	BoneMappings []*BoneMapping `json:"boneMappings"`
}

type BoneMapping struct {
	Bone               string `json:"bone"`
	NodeName           string `json:"nodeName"`
	IgnoreDefaultValue bool   `json:"ignoreDefaultValue,omitempty"`
}

func ApplyConfig(doc *VRMDocument, conf *Config) {
	ext := doc.VRMExt()
	ext.ExporterVersion = "modelconv-BETA"
	ext.Meta = conf.Metadata

	nodeMap := map[string]int{}
	for id, node := range doc.Nodes {
		nodeMap[node.Name] = id
	}

	if len(ext.MaterialProperties) != len(doc.Materials) {
		ext.MaterialProperties = []*MaterialProperty{}
		for _, mat := range doc.Materials {
			var mp MaterialProperty
			mp.Name = mat.Name
			mp.Shader = "VRM_USE_GLTFSHADER"
			mp.RenderQueue = 2000

			mp.FloatProperties = map[string]float64{}
			mp.VectorProperties = map[string]interface{}{}
			mp.TextureProperties = map[string]interface{}{}
			mp.KeywordMap = map[string]interface{}{}
			mp.TagMap = map[string]interface{}{}

			ext.MaterialProperties = append(ext.MaterialProperties, &mp)
		}
	}

	ext.Humanoid.Bones = []*Bone{}
	for _, mapping := range conf.BoneMappings {
		if id, ok := nodeMap[mapping.NodeName]; ok {
			b := &Bone{Bone: mapping.Bone, Node: id, UseDefaultValues: !mapping.IgnoreDefaultValue}
			ext.Humanoid.Bones = append(ext.Humanoid.Bones, b)
		} else {
			log.Println("Bone node not found:", mapping.NodeName)
		}
	}
}

func (doc *VRMDocument) ApplyConfigFile(confpath string) error {
	data, err := ioutil.ReadFile(confpath)
	if err != nil {
		return err
	}
	var conf Config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return err
	}
	ApplyConfig(doc, &conf)
	return nil
}

func (doc *VRMDocument) ValidateBones() error {
	ext := doc.VRMExt()

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
