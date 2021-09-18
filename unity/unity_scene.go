package unity

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

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

	scene := &Scene{GUID: sceneAsset.GUID, Elements: map[int64]interface{}{}, PrefabInstances: map[string]*Scene{}}

	var objects []*GameObject

	docs := parseDocuments(b)
	for _, doc := range docs {
		var component interface{}
		fileId, _ := strconv.ParseInt(doc.refID, 10, 64)

		// log.Println("obj", doc.Tag, doc.refID)
		if doc.Tag == "tag:unity3d.com,2011:1" {
			var a map[string]*GameObject
			err = doc.Decode(&a)
			obj := a["GameObject"]
			obj.Scene = scene
			objects = append(objects, obj)
			component = obj
		} else if doc.Tag == "tag:unity3d.com,2011:4" {
			var a map[string]*Transform
			err = doc.Decode(&a)
			a["Transform"].Scene = scene
			component = a["Transform"]
		} else if doc.Tag == "tag:unity3d.com,2011:23" {
			var a map[string]*MeshRenderer
			err = doc.Decode(&a)
			a["MeshRenderer"].Scene = scene
			component = a["MeshRenderer"]
		} else if doc.Tag == "tag:unity3d.com,2011:33" {
			var a map[string]*MeshFilter
			err = doc.Decode(&a)
			a["MeshFilter"].Scene = scene
			component = a["MeshFilter"]
		} else if doc.Tag == "tag:unity3d.com,2011:1001" {
			var a map[string]*PrefabInstance
			err = doc.Decode(&a)
			component = a["PrefabInstance"]
			if a["PrefabInstance"] != nil && a["PrefabInstance"].SourcePrefab != nil {
				guid := a["PrefabInstance"].SourcePrefab.GUID
				if assets.GetAsset(guid) == nil {
					continue
				}
				s, err := LoadScene(assets, assets.GetAsset(guid))
				if err == nil {
					scene.PrefabInstances[guid] = s
				}
			}
		} else {
			err = doc.Decode(&component)
		}
		if component != nil && err == nil {
			scene.Elements[fileId] = component
		}
		// log.Println("obj", a, err)
	}

	// TODO
	for i, obj := range objects {
		tr := obj.GetTransform()
		if tr == nil || !tr.Father.IsValid() {
			prefab := scene.GetElement(obj.PrefabInstance)

			if prefab != nil && obj.CorrespondingSourceObject != nil {
				prefabRef := obj.CorrespondingSourceObject
				objects[i] = scene.GetGameObject(prefabRef)
			}
			scene.Objects = append(scene.Objects, objects[i])
		}
	}
	return scene, nil
}

func DumpScene(s *Scene) {
	log.Println("Scene", s.GUID)
	for _, obj := range s.Objects {
		obj.Dump(1, true)
	}
}

func (o *GameObject) Dump(indent int, recursive bool) {
	if o == nil {
		log.Println(strings.Repeat(" ", indent*2), "Missing object")
		return
	}
	log.Println(strings.Repeat(" ", indent*2), o.Name)
	for _, c := range o.Components {
		log.Println(strings.Repeat(" ", indent*2), " -", fmt.Sprintf("%#v", o.GetComponent(&c.Ref)))
	}
	tr := o.GetTransform()
	if tr == nil || !recursive {
		return
	}
	for _, c := range tr.Children {
		t := tr.Scene.GetTransform(c)
		if t != nil {
			t.GetGameObject().Dump(indent+1, recursive)
		}
	}
}
