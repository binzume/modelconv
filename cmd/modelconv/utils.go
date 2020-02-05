package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/converter"
	"github.com/binzume/modelconv/mmd"
	"github.com/binzume/modelconv/mqo"
	"github.com/qmuntal/gltf"
)

func loadDocument(input string) (*mqo.MQODocument, error) {
	r, err := os.Open(input)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	if strings.ToLower(filepath.Ext(input)) == ".mqo" {
		return mqo.Parse(r, input)
	}

	pmx, err := mmd.Parse(r)
	if err != nil {
		return nil, err
	}
	log.Println("Name: ", pmx.Name)
	log.Println("Comment: ", pmx.Comment)
	return converter.NewMMDToMQOConverter().Convert(pmx), nil
}

func saveAsGlb(doc *mqo.MQODocument, path, textureDir string) error {
	gltfdoc, err := converter.NewMQOToGLTFConverter().Convert(doc, textureDir)
	if err != nil {
		return err
	}
	return gltf.SaveBinary(gltfdoc, path)
}
