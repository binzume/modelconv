package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/converter"
	"github.com/binzume/modelconv/fbx"
	"github.com/binzume/modelconv/mmd"
	"github.com/binzume/modelconv/mqo"
	"github.com/binzume/modelconv/unity"
	"github.com/qmuntal/gltf"
)

func loadAnimation(input string) (*mmd.Animation, error) {
	r, err := os.Open(input)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var p = mmd.NewVMDParser(r)
	return p.Parse()
}

func loadDocument(input string) (*mqo.Document, error) {
	ext := strings.ToLower(filepath.Ext(input))
	switch {
	case isMQO(ext):
		return mqo.Load(input)
	case isGltf(ext):
		doc, err := gltf.Open(input)
		if err != nil {
			return nil, err
		}
		return converter.NewGLTFToMQOConverter(nil).Convert(doc)
	case isMMD(ext):
		pmx, err := mmd.Load(input)
		if err != nil {
			return nil, err
		}
		log.Println("Name: ", pmx.Name)
		log.Println("Comment: ", pmx.Comment)
		return converter.NewMMDToMQOConverter(nil).Convert(pmx), nil
	case isUnity(ext):
		names := strings.SplitN(input, "#", 2)
		if len(names) == 1 {
			p := strings.Index(filepath.ToSlash(input), "Assets/")
			if p >= 0 && ext != ".unitypackage" {
				names = []string{input[:p+6], input[p:]}
			} else {
				names = append(names, "")
			}
		}
		var assets unity.Assets
		var err error
		if strings.HasSuffix(names[0], ".unitypackage") {
			assets, err = unity.OpenPackage(names[0])
		} else {
			assets, err = unity.OpenAssets(names[0])
		}
		if err != nil {
			return nil, err
		}
		defer assets.Close()
		scene, err := unity.LoadScene(assets, names[1])
		if err != nil {
			return nil, err
		}
		return converter.NewUnityToMQOConverter(&converter.UnityToMQOOption{ConvertPhysics: *convertPhysics}).Convert(scene)
	case ext == ".fbx":
		doc, err := fbx.Load(input)
		if err != nil {
			return nil, err
		}
		return converter.NewFBXToMQOConverter(nil).Convert(doc)
	default:
		return nil, fmt.Errorf("Unspoorted input")
	}
}

func saveAsPmx(doc *mqo.Document, path string) error {
	result, err := converter.NewMQOToMMDConverter(nil).Convert(doc)
	if err != nil {
		return err
	}
	return mmd.Save(result, path)
}
