package unity

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

// Load *.unity or *.prefab
func LoadScene(assets Assets, sceneAsset *Asset) (*Scene, error) {
	r, err := assets.Open(sceneAsset.Path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	scene := NewScene(sceneAsset.GUID)
	var objects []*GameObject

	for _, doc := range ParseYamlDocuments(b) {
		var element Element
		fileId, _ := strconv.ParseInt(doc.refID, 10, 64)

		var components struct {
			GameObject    *GameObject    `yaml:"GameObject"`
			Transform     *Transform     `yaml:"Transform"`
			MeshRenderer  *MeshRenderer  `yaml:"MeshRenderer"`
			MeshFilter    *MeshFilter    `yaml:"MeshFilter"`
			MonoBehaviour *MonoBehaviour `yaml:"MonoBehaviour"`
		}

		// log.Println("obj", doc.Tag, doc.refID)
		if doc.Tag == "tag:unity3d.com,2011:1" {
			err = doc.Decode(&components)
			components.GameObject.Scene = scene
			element = components.GameObject
			objects = append(objects, components.GameObject)
		} else if doc.Tag == "tag:unity3d.com,2011:4" {
			err = doc.Decode(&components)
			components.Transform.Scene = scene
			element = components.Transform
		} else if doc.Tag == "tag:unity3d.com,2011:23" {
			err = doc.Decode(&components)
			components.MeshRenderer.Scene = scene
			element = components.MeshRenderer
		} else if doc.Tag == "tag:unity3d.com,2011:33" {
			err = doc.Decode(&components)
			components.MeshFilter.Scene = scene
			element = components.MeshFilter
		} else if doc.Tag == "tag:unity3d.com,2011:114" {
			err = doc.Decode(&components)
			components.MonoBehaviour.Scene = scene
			element = components.MonoBehaviour
		} else if doc.Tag == "tag:unity3d.com,2011:1001" {
			var a struct {
				PrefabInstance *PrefabInstance `yaml:"PrefabInstance"`
			}
			err = doc.Decode(&a)
			prefab := a.PrefabInstance
			if prefab != nil && prefab.SourcePrefab != nil {
				element = prefab
				guid := prefab.SourcePrefab.GUID
				if assets.GetAsset(guid) == nil {
					continue
				}
				s, err := LoadScene(assets, assets.GetAsset(guid))
				if err == nil {
					scene.PrefabInstances[guid] = s
				}
				if len(s.Objects) > 0 {
					root := s.Objects[0]
					// TODO: apply qll modifications
					if root.GetTransform() != nil {
						root.GetTransform().Father = *prefab.Modification.TransformParent
						root.GetTransform().Father.GUID = scene.GUID
					}
					objects = append(objects, root)
				}
			}
		} else {
			err = doc.Decode(&element)
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
	for _, c := range tr.Children {
		if t, ok := tr.Scene.GetElement(c).(*Transform); ok {
			t.GetGameObject().Dump(indent+1, recursive, component)
		}
	}
}
