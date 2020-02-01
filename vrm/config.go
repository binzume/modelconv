package vrm

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	Metadata Metadata `json:"meta"`

	BoneMappings []*struct {
		Bone
		NodeName string `json:"nodeName"`
	} `json:"boneMappings"`

	AnimationBoneGroups []*struct {
		SecondaryAnimationBoneGroup
		NodeNames []string `json:"nodeNames"`
	} `json:"animationBoneGroups"`

	MorphMappings    []*MorphMapping             `json:"morphMappings"` // EXPERIMENTAL
	MaterialSettings map[string]*MaterialSetting `json:"materialSettings"`
}

type MorphMapping struct {
	Name        string `json:"name"`
	NodeName    string `json:"nodeName"`
	TargetIndex int    `json:"targetIndex"`
}

type MaterialSetting struct {
	ForceUnlit bool `json:"forceUnlit"`
}

func ApplyConfig(doc *VRMDocument, conf *Config) {
	ext := doc.VRM()
	ext.ExporterVersion = ExporterVersion
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

	if len(conf.MaterialSettings) > 0 {
		var unlitMaterialExt = "KHR_materials_unlit"
		for _, mat := range doc.Materials {
			setting := conf.MaterialSettings[mat.Name]
			if setting == nil {
				setting = conf.MaterialSettings["*"]
			}
			if setting == nil {
				continue
			}
			if setting.ForceUnlit {
				if !doc.IsExtentionUsed(unlitMaterialExt) {
					doc.ExtensionsUsed = append(doc.ExtensionsUsed, unlitMaterialExt)
				}
				mat.Extensions = map[string]interface{}{unlitMaterialExt: map[string]string{}}
			}
		}
	}

	found := map[string]bool{}
	ext.Humanoid.Bones = []*Bone{}
	for _, mapping := range conf.BoneMappings {
		if id, ok := nodeMap[mapping.NodeName]; ok {
			var b = mapping.Bone
			found[mapping.Bone.Bone] = true
			b.Node = id
			b.UseDefaultValues = b.UseDefaultValues || b.Min == nil && b.Max == nil && b.Center == nil
			ext.Humanoid.Bones = append(ext.Humanoid.Bones, &b)
		} else {
			log.Println("Bone node not found:", mapping.NodeName)
		}
	}
	for _, name := range RequiredBones {
		if id, ok := nodeMap[name]; ok && !found[name] {
			ext.Humanoid.Bones = append(ext.Humanoid.Bones, &Bone{Bone: name, Node: id, UseDefaultValues: true})
		}
	}

	for _, boneGroup := range conf.AnimationBoneGroups {
		var b = boneGroup.SecondaryAnimationBoneGroup
		for _, nodeName := range boneGroup.NodeNames {
			if id, ok := nodeMap[nodeName]; ok {
				b.Bones = append(b.Bones, id)
			} else {
				log.Println("Bone node not found:", nodeName)
			}
		}
		if len(b.Bones) > 0 {
			if ext.SecondaryAnimation == nil {
				ext.SecondaryAnimation = &SecondaryAnimation{}
			}
			ext.SecondaryAnimation.BoneGroups = append(ext.SecondaryAnimation.BoneGroups, &b)
		}
	}

	for _, mapping := range conf.MorphMappings {
		if id, ok := nodeMap[mapping.NodeName]; ok {
			m := &BlendShapeGroup{
				Name:       mapping.Name,
				PresetName: mapping.Name,
				Binds: []interface{}{
					map[string]interface{}{
						"mesh":   doc.Nodes[id].Mesh,
						"index":  mapping.TargetIndex,
						"weight": 100,
					},
				},
			}
			ext.BlendShapeMaster.BlendShapeGroups = append(ext.BlendShapeMaster.BlendShapeGroups, m)
		} else {
			log.Println("Morph node not found:", mapping.NodeName)
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
