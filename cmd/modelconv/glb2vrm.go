package main

import (
	"log"
	"os"

	"github.com/binzume/modelconv/converter"
	"github.com/binzume/modelconv/vrm"
	"github.com/qmuntal/gltf"
)

func saveAsVRM(gltfDoc *gltf.Document, output, confFile string) error {
	vrm.FixJointComponentType(gltfDoc)
	vrm.FixJointMatrix(gltfDoc)
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
	if err := doc.ValidateBones(); err != nil {
		log.Print(err)
	}

	return gltf.SaveBinary(gltfDoc, output)
}
