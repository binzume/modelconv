package converter

import (
	"embed"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/gltfutil"
	"github.com/binzume/modelconv/vrm"
	"github.com/qmuntal/gltf"
)

type Config struct {
	Metadata vrm.Metadata `json:"meta"`

	BoneMappings     []*BoneMapping              `json:"boneMappings"`
	MorphMappings    []*MorphMapping             `json:"morphMappings"`
	MaterialSettings map[string]*MaterialSetting `json:"materialSettings"`

	ExportAllMorph           bool `json:"exportAllMorph"`
	ConvertPhysicsCollider   bool `json:"convertPhysicsCollider"`
	AnimationBoneFromPhysics bool `json:"animationBoneFromPhysics"`

	AnimationBoneGroups []*AnimationBoneGroupSettings `json:"animationBoneGroups"`
	ColliderGroups      []*ColliderGroupSettings      `json:"colliderGroups"`

	Preset string `json:"preset"`
}

type AnimationBoneGroupSettings struct {
	vrm.SecondaryAnimationBoneGroup
	NodeNames []string `json:"nodeNames"`
}

type ColliderGroupSettings struct {
	vrm.SecondaryAnimationColliderGroup
	NodeName string `json:"nodeName"`
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
	ForceUnlit bool    `json:"forceUnlit"`
	AlphaMode  string  `json:"alphaMode"`
	Alpha      float32 `json:"alpha"`
}

//go:embed vrmconfig_presets/*.json
var presetsFs embed.FS

func LoadVRMConfig(name string) (*Config, error) {
	var data []byte
	var err error
	if !strings.Contains(name, ".") && !strings.Contains(name, "/") {
		execPath, err := os.Executable()
		if err != nil {
			return nil, err
		}
		presetPath := filepath.Join(filepath.Dir(execPath), "vrmconfig_presets", name+".json")
		if _, err := os.Stat(presetPath); err == nil {
			data, err = ioutil.ReadFile(presetPath)
		} else {
			data, err = presetsFs.ReadFile(path.Join("vrmconfig_presets", name+".json"))
		}
	} else {
		data, err = ioutil.ReadFile(name)
	}
	if err != nil {
		return nil, err
	}
	var presetConf Config
	err = json.Unmarshal(data, &presetConf)
	if err != nil {
		return nil, err
	}
	return &presetConf, nil
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

func matchNodeName(pattern string, nodeMap map[string]int) []int {
	var result []int
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		for name, id := range nodeMap {
			if m, _ := path.Match(pattern, name); m {
				result = append(result, id)
			}
		}
	} else if id, ok := nodeMap[pattern]; ok {
		result = append(result, id)
	}
	return result
}

func findReursive(doc *vrm.Document, target uint32, m map[uint32]string) (string, bool) {
	if name, ok := m[target]; ok {
		return name, ok
	}
	for _, c := range doc.Nodes[target].Children {
		if name, ok := findReursive(doc, c, m); ok {
			return name, ok
		}
	}
	return "", false
}

func registerReursive(doc *vrm.Document, target uint32, m map[uint32]string, name string) {
	m[target] = name
	for _, c := range doc.Nodes[target].Children {
		registerReursive(doc, c, m, name)
	}
}

func applyConfigInternal(doc *vrm.Document, conf *Config, foundBones map[string]int, nodeMap map[string]int, blendShapeMap map[[2]int]string) {
	ext := doc.VRM()
	springBones := map[uint32]string{}
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
			} else if setting.AlphaMode == "opaque" {
				mat.AlphaMode = gltf.AlphaOpaque
			}
			if setting.Alpha > 0 && mat.PBRMetallicRoughness != nil {
				mat.PBRMetallicRoughness.BaseColorFactor[3] = setting.Alpha
			}
		}
	}

	for _, boneGroup := range conf.AnimationBoneGroups {
		var b = boneGroup.SecondaryAnimationBoneGroup
		for _, nodeName := range boneGroup.NodeNames {
			matched := matchNodeName(nodeName, nodeMap)
			for _, n := range matched {
				if doc.Nodes[n].Mesh != nil {
					log.Printf("ERROR: %v is not bone node!", nodeName)
					continue
				}
				if name, ok := findReursive(doc, uint32(n), springBones); ok {
					log.Printf("ERROR: %s already added as %s", nodeName, name)
					continue
				}
				registerReursive(doc, uint32(n), springBones, nodeName)
				b.Bones = append(b.Bones, n)
			}
			if len(matched) == 0 {
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

func ApplyConfig(doc *vrm.Document, conf *Config) (*vrm.Document, error) {
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
		presetConf, err := LoadVRMConfig(conf.Preset)
		if err != nil {
			return doc, err
		}
		conf.MergePreset(presetConf)
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

	return doc, nil
}

func ToVRM(gltfDoc *gltf.Document, output, srcDir, confFile string) (*vrm.Document, error) {
	if err := gltfutil.ToSingleFile(gltfDoc, srcDir); err != nil {
		return nil, err
	}
	gltfutil.FixJointComponentType(gltfDoc)
	gltfutil.ResetJointMatrix(gltfDoc)
	doc := (*vrm.Document)(gltfDoc)
	if confFile != "" {
		conf, err := LoadVRMConfig(confFile)
		if err != nil {
			return doc, err
		}
		return ApplyConfig(doc, conf)
	}
	return doc, nil
}

func springLen(doc *gltf.Document, node *gltf.Node) int {
	if len(node.Children) == 0 {
		return 0
	}
	p, ok := node.Extensions[BlenderPhysicsName].(*BlenderPhysicsBody)
	if !ok {
		return -1
	}
	if p.Static {
		return -1
	}
	l := 0
	for _, n := range node.Children {
		d := springLen(doc, doc.Nodes[n])
		if d < 0 {
			return -1
		}
		if d >= l {
			l = d + 1
		}
	}
	return l
}

func DetectSpringBones(doc *gltf.Document, conf *Config) {
	parents := map[int]int{}
	for i, n := range doc.Nodes {
		for _, c := range n.Children {
			parents[int(c)] = i
		}
	}
	var nodeNames []string
	for i, n := range doc.Nodes {
		l := springLen(doc, n)
		if l > 0 && springLen(doc, doc.Nodes[parents[i]]) < 0 {
			log.Println("Add SpringBone:", i, n.Name)
			nodeNames = append(nodeNames, n.Name)
		}
	}
	if len(nodeNames) > 0 {
		// TODO: physics parameters.
		conf.AnimationBoneGroups = append(conf.AnimationBoneGroups,
			&AnimationBoneGroupSettings{
				SecondaryAnimationBoneGroup: vrm.SecondaryAnimationBoneGroup{
					Comment:        "auto generated",
					Stiffiness:     0.3,
					DragForce:      0.2,
					HitRadius:      0.02,
					Center:         -1,
					ColliderGroups: nil,
				}, NodeNames: nodeNames})
	}
}

func ConvertColliders(doc *gltf.Document, conf *Config) {
	for i, node := range doc.Nodes {
		p, ok := node.Extensions[BlenderPhysicsName].(*BlenderPhysicsBody)
		if !ok || !p.Static {
			continue
		}
		var colliders []*vrm.SecondaryAnimationCollider
		for _, s := range p.Shapes {
			if s["shapeType"] == "SPHERE" || s["shapeType"] == "CAPSULE" {
				if scale, ok := s["offsetScale"].([3]float32); ok && scale[0] != 0 {
					collider := &vrm.SecondaryAnimationCollider{
						Radius: float64(scale[0]),
						Offset: map[string]float64{"x": 0, "y": 0, "z": 0},
					}
					if pos, ok := s["offsetTranslation"].([3]float32); ok {
						collider.Offset = map[string]float64{"x": float64(pos[0]), "y": float64(pos[1]), "z": float64(pos[2])}
					}
					colliders = append(colliders, collider)
				}
			}
		}
		log.Println(colliders)
		if len(colliders) > 0 {
			conf.ColliderGroups = append(conf.ColliderGroups, &ColliderGroupSettings{
				SecondaryAnimationColliderGroup: vrm.SecondaryAnimationColliderGroup{
					Colliders: colliders,
					Node:      uint32(i),
				},
				NodeName: node.Name,
			})
		}
	}
}

func ApplyVRMConfig(src *gltf.Document, output, srcDir string, conf *Config) (*vrm.Document, error) {
	if err := gltfutil.ToSingleFile(src, srcDir); err != nil {
		return nil, err
	}
	if conf.AnimationBoneFromPhysics {
		DetectSpringBones(src, conf)
	}
	if conf.ConvertPhysicsCollider {
		ConvertColliders(src, conf)
	}
	gltfutil.RemoveExtension(src, BlenderPhysicsName)
	gltfutil.FixJointComponentType(src)
	gltfutil.ResetJointMatrix(src)
	return ApplyConfig((*vrm.Document)(src), conf)
}
