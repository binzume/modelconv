package main

import (
	"log"

	"github.com/binzume/modelconv/converter"
	"github.com/qmuntal/gltf"
)

func saveAsVRM(gltfDoc *gltf.Document, output, srcDir, confFile string) error {
	doc, err := converter.ToVRM(gltfDoc, output, srcDir, confFile)
	if err != nil {
		gltf.SaveBinary(gltfDoc, output) // for debug
		return err
	}
	log.Print("Title: ", doc.VRM().Title())
	log.Print("Author: ", doc.VRM().Author())
	if err := doc.ValidateBones(); err != nil {
		log.Print(err)
	}
	return gltf.SaveBinary(gltfDoc, output)
}
