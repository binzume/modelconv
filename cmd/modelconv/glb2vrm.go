package main

import (
	"log"
	"os"

	"github.com/binzume/modelconv/converter"
	"github.com/binzume/modelconv/vrm"
	"github.com/qmuntal/gltf"
)

func glb2vrm(input, output, confFile string) error {
	doc, err := gltf.Open(input)
	if err != nil {
		return err
	}
	return saveAsVRM(doc, output, confFile)
}

func saveAsVRM(gltfDoc *gltf.Document, output, confFile string) error {
	doc := (*vrm.Document)(gltfDoc)
	if _, err := os.Stat(confFile); err != nil {
		log.Print("vrm config error: ", err)
	} else {
		if err = converter.ApplyVRMConfigFile(doc, confFile); err != nil {
			return err
		}
	}
	log.Print("Title: ", doc.VRM().Title())
	log.Print("Author: ", doc.VRM().Author())
	doc.FixJointComponentType()
	doc.FixJointMatrix()
	if err := doc.ValidateBones(); err != nil {
		log.Print(err)
	}

	return gltf.SaveBinary(gltfDoc, output)
}
