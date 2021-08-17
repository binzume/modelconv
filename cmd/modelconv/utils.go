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

	if isGltf(ext) {
		doc, err := gltf.Open(input)
		if err != nil {
			return nil, err
		}
		return converter.NewGLTFToMQOConverter(nil).Convert(doc)
	}

	r, err := os.Open(input)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	if isMQO(ext) {
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

func saveAsPmx(doc *mqo.Document, path string) error {
	result, err := converter.NewMQOToMMDConverter().Convert(doc)
	if err != nil {
		return err
	}
	w, _ := os.Create(path)
	return mmd.WritePMX(result, w)
}
