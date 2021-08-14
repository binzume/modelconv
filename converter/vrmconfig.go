package converter

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/binzume/modelconv/vrm"
	"github.com/qmuntal/gltf"
)

type Config struct {
	Metadata vrm.Metadata `json:"meta"`

	BoneMappings     []*BoneMapping              `json:"boneMappings"`
	MorphMappings    []*MorphMapping             `json:"morphMappings"`
	MaterialSettings map[string]*MaterialSetting `json:"materialSettings"`
	ExportAllMorph   bool                        `json:"exportAllMorph"`

	AnimationBoneGroups []*struct {
		vrm.SecondaryAnimationBoneGroup
		NodeNames []string `json:"nodeNames"`
	} `json:"animationBoneGroups"`
	ColliderGroups []*struct {
		vrm.SecondaryAnimationColliderGroup
		NodeName string `json:"nodeName"`
	} `json:"colliderGroups"`

	Preset string `json:"preset"`
}

type BoneMapping struct {
	vrm.Bone
	NodeName  string   `json:"nodeName"` // deprecated
	NodeNames []string `json:"nodeNames"`
}

type MorphMapping struct {
	Name        string `json:"name"`
	NodeName    string `json:"nodeName"`
	TargetName  string `json:"targetName"`
	TargetIndex int    `json:"targetIndex"`

	MaterialValues []*vrm.BlendShapeMaterialValue `json:"materialValues,omitempty"`
}

type MaterialSetting struct {
	ForceUnlit bool   `json:"forceUnlit"`
	AlphaMode  string `json:"alphaMode"`
}

func (c *Config) MergePreset(preset *Config) {
	for name, m := range preset.MaterialSettings {
		if _, exist := c.MaterialSettings[name]; !exist {
			c.MaterialSettings[name] = m
		}
	}
	c.BoneMappings = append(c.BoneMappings, preset.BoneMappings...)
	c.MorphMappings = append(c.MorphMappings, preset.MorphMappings...)
	c.AnimationBoneGroups = append(c.AnimationBoneGroups, preset.AnimationBoneGroups...)
	c.ColliderGroups = append(c.ColliderGroups, preset.ColliderGroups...)
}

func applyConfigInternal(doc *vrm.Document, conf *Config, foundBones map[string]int, nodeMap map[string]int, blendShapeMap map[[2]int]string) {
	ext := doc.VRM()
	for _, mapping := range conf.BoneMappings {
		if _, ok := foundBones[mapping.Bone.Bone]; ok {
			continue
		}
		if mapping.NodeName != "" {
			mapping.NodeNames = append(mapping.NodeNames, mapping.NodeName)
			mapping.NodeName = ""
		}
		if len(mapping.NodeNames) == 0 {
			// ignore
			foundBones[mapping.Bone.Bone] = -1
			continue
		}
		found := false
		for _, nodeName := range mapping.NodeNames {
			if id, ok := nodeMap[nodeName]; ok {
				if doc.Nodes[id].Mesh != nil {
					log.Printf("ERROR: %v is not bone node!", nodeName)
					continue
				}
				var b = mapping.Bone
				foundBones[mapping.Bone.Bone] = id
				b.Node = id
				b.UseDefaultValues = b.UseDefaultValues || b.Min == nil && b.Max == nil && b.Center == nil
				ext.Humanoid.Bones = append(ext.Humanoid.Bones, &b)
				found = true
				break
			}
		}
		if len(mapping.NodeNames) > 0 && !found {
			log.Println("Bone node not found:", mapping.NodeNames)
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
			if setting.AlphaMode == "blend" {
				mat.AlphaMode = gltf.AlphaBlend
			} else if setting.AlphaMode == "mask" {
				mat.AlphaMode = gltf.AlphaMask
			}
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
				ext.SecondaryAnimation = &vrm.SecondaryAnimation{}
			}
			ext.SecondaryAnimation.BoneGroups = append(ext.SecondaryAnimation.BoneGroups, &b)
		}
	}

	for _, colliderGroup := range conf.ColliderGroups {
		var b = colliderGroup.SecondaryAnimationColliderGroup
		if id, ok := nodeMap[colliderGroup.NodeName]; ok {
			b.Node = uint32(id)
		} else {
			log.Println("Node not found:", colliderGroup.NodeName)
			continue
		}
		if ext.SecondaryAnimation == nil {
			ext.SecondaryAnimation = &vrm.SecondaryAnimation{}
		}
		ext.SecondaryAnimation.ColliderGroups = append(ext.SecondaryAnimation.ColliderGroups, &b)
	}

	targets := map[string][2]int{}
	for mi, mesh := range doc.Meshes {
		if extras, ok := mesh.Extras.(map[string]interface{}); ok {
			if names, ok := extras["targetNames"].([]string); ok {
				for i, name := range names {
					targets[name] = [2]int{mi, i}
				}
			}
		}
	}
	for _, mapping := range conf.MorphMappings {
		if mapping.TargetName != "" {
			m := &vrm.BlendShapeGroup{
				Name:           mapping.Name,
				PresetName:     mapping.Name,
				MaterialValues: mapping.MaterialValues,
			}
			if t, ok := targets[mapping.TargetName]; ok {
				blendShapeMap[t] = mapping.Name
				m.Binds = []*vrm.BlendShapeBind{
					{
						Mesh: uint32(t[0]), Index: t[1], Weight: 100,
					},
				}
			}
			if len(m.Binds) > 0 || len(m.MaterialValues) > 0 {
				ext.BlendShapeMaster.BlendShapeGroups = append(ext.BlendShapeMaster.BlendShapeGroups, m)
			}
			continue
		}

		if id, ok := nodeMap[mapping.NodeName]; ok {
			blendShapeMap[[2]int{int(*doc.Nodes[id].Mesh), mapping.TargetIndex}] = mapping.Name
			m := &vrm.BlendShapeGroup{
				Name:       mapping.Name,
				PresetName: mapping.Name,
				Binds: []*vrm.BlendShapeBind{
					{
						Mesh: *doc.Nodes[id].Mesh, Index: mapping.TargetIndex, Weight: 100,
					},
				},
			}
			ext.BlendShapeMaster.BlendShapeGroups = append(ext.BlendShapeMaster.BlendShapeGroups, m)
		} else if mapping.NodeName == "" {
			log.Println("Morph node not found:", mapping.NodeName, ":", mapping.TargetIndex)
		}
	}
}

func ApplyConfig(doc *vrm.Document, conf *Config) {
	ext := doc.VRM()
	ext.ExporterVersion = vrm.ExporterVersion
	ext.Meta = conf.Metadata

	if len(ext.MaterialProperties) != len(doc.Materials) {
		ext.MaterialProperties = []*vrm.MaterialProperty{}
		for _, mat := range doc.Materials {
			mp := vrm.NewMaterialProperty(mat.Name)
			mp.Shader = "VRM_USE_GLTFSHADER"
			ext.MaterialProperties = append(ext.MaterialProperties, mp)
		}
	}

	nodeMap := map[string]int{}
	for id, node := range doc.Nodes {
		nodeMap[node.Name] = id
	}
	foundBones := map[string]int{}
	ext.Humanoid.Bones = []*vrm.Bone{}

	blendShapeMap := map[[2]int]string{}

	if conf.Preset != "" {
		execPath, err := os.Executable()
		if err != nil {
			log.Fatal("preset error:", conf.Preset, err)
		}
		presetPath := filepath.Join(filepath.Dir(execPath), "vrmconfig_presets", conf.Preset+".json")
		data, err := ioutil.ReadFile(presetPath)
		if err != nil {
			log.Fatal("preset error:", conf.Preset, err)
		}
		var presetConf Config
		err = json.Unmarshal(data, &presetConf)
		if err != nil {
			log.Fatal("preset error:", conf.Preset, err)
		}
		conf.MergePreset(&presetConf)
	}

	applyConfigInternal(doc, conf, foundBones, nodeMap, blendShapeMap)

	if conf.ExportAllMorph {
		for mi, mesh := range doc.Meshes {
			if extras, ok := mesh.Extras.(map[string]interface{}); ok {
				if names, ok := extras["targetNames"].([]string); ok {
					for i, name := range names {
						if _, used := blendShapeMap[[2]int{mi, i}]; used {
							continue
						}
						m := &vrm.BlendShapeGroup{
							Name: name,
							Binds: []*vrm.BlendShapeBind{
								{
									Mesh: uint32(mi), Index: i, Weight: 100,
								},
							},
						}
						ext.BlendShapeMaster.BlendShapeGroups = append(ext.BlendShapeMaster.BlendShapeGroups, m)
					}
				}
			}
		}
	}

	if ext.FirstPerson == nil {
		if node, ok := foundBones["head"]; ok {
			ext.FirstPerson = &vrm.FirstPerson{
				FirstPersonBone: node,
			}
		}
	}

	for _, name := range vrm.RequiredBones {
		if _, ok := foundBones[name]; ok {
			continue
		}
		if id, ok := nodeMap[name]; ok {
			ext.Humanoid.Bones = append(ext.Humanoid.Bones, &vrm.Bone{Bone: name, Node: id, UseDefaultValues: true})
		}
	}
}

func ApplyVRMConfigFile(doc *vrm.Document, confpath string) error {
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
