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

// TODO: move into converter package.
type Config struct {
	Metadata vrm.Metadata `json:"meta"`

	BoneMappings     []*BoneMapping              `json:"boneMappings"`
	MorphMappings    []*MorphMapping             `json:"morphMappings"` // EXPERIMENTAL
	MaterialSettings map[string]*MaterialSetting `json:"materialSettings"`

	AnimationBoneGroups []*struct {
		vrm.SecondaryAnimationBoneGroup
		NodeNames []string `json:"nodeNames"`
	} `json:"animationBoneGroups"`

	Preset string `json:"preset"`
}

type BoneMapping struct {
	vrm.Bone
	NodeName string `json:"nodeName"`
}

type MorphMapping struct {
	Name        string `json:"name"`
	NodeName    string `json:"nodeName"`
	TargetName  string `json:"targetName"`
	TargetIndex int    `json:"targetIndex"`
}

type MaterialSetting struct {
	ForceUnlit bool   `json:"forceUnlit"`
	AlphaMode  string `json:"alphaMode"`
}

func applyConfigInternal(doc *vrm.Document, conf *Config, foundBones map[string]bool, nodeMap map[string]int) {
	ext := doc.VRM()
	for _, mapping := range conf.BoneMappings {
		if foundBones[mapping.Bone.Bone] {
			continue
		}
		if mapping.NodeName == "" {
			// ignore
			foundBones[mapping.Bone.Bone] = true
			continue
		}
		if id, ok := nodeMap[mapping.NodeName]; ok {
			var b = mapping.Bone
			foundBones[mapping.Bone.Bone] = true
			b.Node = id
			b.UseDefaultValues = b.UseDefaultValues || b.Min == nil && b.Max == nil && b.Center == nil
			ext.Humanoid.Bones = append(ext.Humanoid.Bones, &b)
		} else {
			log.Println("Bone node not found:", mapping.NodeName)
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
			if t, ok := targets[mapping.TargetName]; ok {
				m := &vrm.BlendShapeGroup{
					Name:       mapping.Name,
					PresetName: mapping.Name,
					Binds: []interface{}{
						map[string]interface{}{
							"mesh":   t[0],
							"index":  t[1],
							"weight": 100,
						},
					},
				}
				ext.BlendShapeMaster.BlendShapeGroups = append(ext.BlendShapeMaster.BlendShapeGroups, m)
			}
			continue
		}

		if id, ok := nodeMap[mapping.NodeName]; ok {
			m := &vrm.BlendShapeGroup{
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
			var mp vrm.MaterialProperty
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

	nodeMap := map[string]int{}
	for id, node := range doc.Nodes {
		nodeMap[node.Name] = id
	}
	foundBones := map[string]bool{}
	ext.Humanoid.Bones = []*vrm.Bone{}

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
		applyConfigInternal(doc, &presetConf, foundBones, nodeMap)
	}

	applyConfigInternal(doc, conf, foundBones, nodeMap)

	for _, name := range vrm.RequiredBones {
		if id, ok := nodeMap[name]; ok && !foundBones[name] {
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
