package unity

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// Load *.unity or *.prefab
func LoadScene(assets Assets, scenePath string) (*Scene, error) {
	sceneAsset := assets.GetAssetByPath(scenePath)
	if sceneAsset == nil {
		log.Println("Scenes:")
		for _, a := range assets.GetAllAssets() {
			if strings.HasSuffix(a.Path, ".unity") {
				log.Println("  " + assets.GetSourcePath() + "#" + a.Path)
			}
		}
		return nil, fmt.Errorf("Scene not found: %s", scenePath)
	}
	return LoadSceneAsset(assets, sceneAsset)
}

func LoadSceneAsset(assets Assets, sceneAsset *Asset) (*Scene, error) {
	r, err := assets.Open(sceneAsset.Path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var components componentDesc
	typ := reflect.TypeOf(components)
	componentTags := map[string]int{}
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Tag.Get("typeid") != "" {
			componentTags["tag:"+typ.Field(i).Tag.Get("typeid")] = i
		}
	}

	scene := NewScene(assets, sceneAsset.GUID)
	var objects []*GameObject

	addedChildren := map[Ref][]*Transform{}

	for _, doc := range ParseYamlDocuments(b) {
		var element Element
		fileId, _ := strconv.ParseInt(doc.refID, 10, 64)

		// log.Println("obj", doc.Tag, doc.refID)
		if doc.Tag == "tag:unity3d.com,2011:1" {
			var d struct {
				GameObject *GameObject `yaml:"GameObject" typeid:"unity3d.com,2011:1"`
			}
			err = doc.Decode(&d)
			d.GameObject.init(scene)
			element = d.GameObject
			objects = append(objects, d.GameObject)
		} else if doc.Tag == "tag:unity3d.com,2011:1001" {
			var a struct {
				PrefabInstance *PrefabInstance `yaml:"PrefabInstance"`
				Prefab         *PrefabInstance `yaml:"Prefab"`
			}
			err = doc.Decode(&a)
			if a.Prefab != nil {
				continue
			}
			prefab := a.PrefabInstance
			if err != nil || prefab == nil || !prefab.SourcePrefab.IsValid() {
				log.Println("invalid prefabInstance", err, scene.GUID, fileId)
				continue
			}
			prefabAsset := assets.GetAsset(prefab.SourcePrefab.GUID)
			if prefabAsset == nil {
				log.Println("prefab not found", prefab.SourcePrefab)
				continue
			}
			s, err := LoadSceneAsset(assets, prefabAsset)
			if err != nil {
				log.Println("failed to loadPrefab", prefab.SourcePrefab)
				continue
			}
			prefab.PrefabScene = s
			element = prefab
			if len(s.Objects) > 0 {
				root := s.Objects[0]
				tr := root.GetTransform()
				if tr != nil {
					tr.Father = *prefab.Modification.TransformParent
					tr.Father.GUID = scene.GUID
					for _, c := range root.Components {
						if _, ok := s.GetElement2(&c.Ref, root.PrefabInstance).(*Transform); ok {
							addedChildren[tr.Father] = append(addedChildren[tr.Father], tr)
						}
					}
				}
				objects = append(objects, root)
			}
			for _, mod := range prefab.Modification.Modifications {
				target := s.GetElement(mod.Target)
				var flaotValue float32
				if v, ok := mod.Value.(float32); ok {
					flaotValue = v
				} else if v, ok := mod.Value.(float64); ok {
					flaotValue = float32(v)
				} else if v, ok := mod.Value.(int64); ok {
					flaotValue = float32(v)
				} else if v, ok := mod.Value.(int); ok {
					flaotValue = float32(v)
				}
				// TODO: reflection
				if t, ok := target.(*Transform); ok {
					switch mod.PropertyPath {
					case "m_LocalPosition.x":
						t.LocalPosition.X = flaotValue
						break
					case "m_LocalPosition.y":
						t.LocalPosition.Y = flaotValue
						break
					case "m_LocalPosition.z":
						t.LocalPosition.Z = flaotValue
						break
					case "m_LocalScale.x":
						t.LocalScale.X = flaotValue
						break
					case "m_LocalScale.y":
						t.LocalScale.Y = flaotValue
						break
					case "m_LocalScale.z":
						t.LocalScale.Z = flaotValue
						break
					case "m_LocalRotation.x":
						t.LocalRotation.X = flaotValue
						break
					case "m_LocalRotation.y":
						t.LocalRotation.Y = flaotValue
						break
					case "m_LocalRotation.z":
						t.LocalRotation.Z = flaotValue
						break
					case "m_LocalRotation.w":
						t.LocalRotation.W = flaotValue
						break
					case "m_RootOrder":
						t.RootOrder = int(flaotValue)
						break
					case "m_LocalEulerAnglesHint.x":
					case "m_LocalEulerAnglesHint.y":
					case "m_LocalEulerAnglesHint.z":
						// ignore
						break
					default:
						log.Println("Unsupported modification property:", mod.PropertyPath)
					}
				} else if t, ok := target.(*GameObject); ok {
					switch mod.PropertyPath {
					case "m_IsActive":
						t.IsActive = int(flaotValue)
					case "m_Name":
						t.Name, _ = mod.Value.(string)
					case "m_TagString":
						t.TagString, _ = mod.Value.(string)
					}
				} else if t, ok := target.(*MeshRenderer); ok {
					if strings.HasPrefix(mod.PropertyPath, "m_Materials.Array.data[") {
						i := int(mod.PropertyPath[23] - '0')
						if i < len(t.Materials) {
							t.Materials[i] = mod.ObjectReference
						}
					}
				} else {
					// log.Printf("%T, %v, %v\n", target, mod.PropertyPath, mod.ObjectReference)
				}
			}
		} else {
			if fieldid, ok := componentTags[doc.Tag]; ok {
				var components componentDesc
				err = doc.Decode(&components)
				if component, _ := reflect.ValueOf(components).Field(fieldid).Interface().(Component); component != nil {
					component.init(scene)
					element = component
				}
			}
			if element == nil {
				err = doc.Decode(&element)
			}
		}
		if element != nil && err == nil {
			scene.Elements[fileId] = element
		}
		// log.Println("obj", a, err)
	}

	// Root objects
	for i, obj := range objects {
		tr := obj.GetTransform()
		if tr != nil && !tr.Father.IsValid() && !obj.CorrespondingSourceObject.IsValid() {
			scene.Objects = append(scene.Objects, objects[i])
		}
	}

	for pref, children := range addedChildren {
		parent := scene.GetTransform(&pref, nil)
		if parent == nil {
			continue
		}
		for _, child := range children {
			parent.AddChild(child)
		}
	}
	return scene, nil
}

func DumpScene(s *Scene, dumpComponents bool) {
	log.Println("Scene:", s.GUID)
	for _, obj := range s.Objects {
		obj.Dump(1, true, dumpComponents)
	}
}

func (o *GameObject) Dump(indent int, recursive, component bool) {
	if o == nil {
		log.Println(strings.Repeat(" ", indent*2), "Object: Missing reference")
		return
	}
	tr := o.GetTransform()
	prefix := "Object:"
	if tr != nil && tr.Father.GUID != "" && tr.Father.GUID != o.Scene.GUID {
		prefix += "(Prefab)"
	}
	log.Println(strings.Repeat(" ", indent*2), prefix, o.Name)
	if component {
		for _, c := range o.Components {
			log.Println(strings.Repeat(" ", indent*2), " -", fmt.Sprintf("%#v", o.Scene.GetElement(&c.Ref)))
		}
	}
	if tr == nil || !recursive {
		return
	}
	for _, child := range tr.GetChildren() {
		child.GetGameObject().Dump(indent+1, recursive, component)
	}
}
