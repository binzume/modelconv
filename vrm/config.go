package vrm

import (
	"encoding/json"
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
	ext.Meta = conf.Metadata

	nodeMap := map[string]int{}
	for id, node := range doc.Nodes {
		nodeMap[node.Name] = id
	}

	for _, mapping := range conf.BoneMappings {
		if id, ok := nodeMap[mapping.NodeName]; ok {
			b := &Bone{Bone: mapping.Bone, Node: id, UseDefaultValues: !mapping.IgnoreDefaultValue}
			ext.Humanoid.Bones = append(ext.Humanoid.Bones, b)
		} else {
			log.Println("Bone node not found:", mapping.NodeName)
		}
	}
}

func ApplyConfigFile(doc *VRMDocument, confpath string) error {
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
