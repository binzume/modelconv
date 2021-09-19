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

	if isMQO(ext) {
		return mqo.Load(input)
	} else if isGltf(ext) {
		doc, err := gltf.Open(input)
		if err != nil {
			return nil, err
		}
		return converter.NewGLTFToMQOConverter(nil).Convert(doc)
	} else if isMMD(ext) {
		pmx, err := mmd.Load(input)
		if err != nil {
			return nil, err
		}
		log.Println("Name: ", pmx.Name)
		log.Println("Comment: ", pmx.Comment)
		return converter.NewMMDToMQOConverter(nil).Convert(pmx), nil
	} else if ext == ".unity" || ext == ".prefab" {
		names := strings.Split(input, "#")
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
		return converter.NewUnityToMQOConverter(nil).Convert(scene)
	}

	if ext == ".fbx" {
		doc, err := fbx.Load(input)
		if err != nil {
			return nil, err
		}
		return converter.NewFBXToMQOConverter(nil).Convert(doc)
	}

	return nil, fmt.Errorf("Unspoorted input")
}

func saveAsPmx(doc *mqo.Document, path string) error {
	result, err := converter.NewMQOToMMDConverter(nil).Convert(doc)
	if err != nil {
		return err
	}
	return mmd.Save(result, path)
}
