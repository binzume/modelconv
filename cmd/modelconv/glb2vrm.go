package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/binzume/modelconv/vrm"
)

func glb2vrm(input, output, confFile string) error {
	if output == "" {
		output = input[0:len(input)-len(filepath.Ext(input))] + ".vrm"
	}
	if confFile == "" {
		confFile = input[0:len(input)-len(filepath.Ext(input))] + ".vrmconfig.json"
	}

	r, err := os.Open(input)
	if err != nil {
		return err
	}
	defer r.Close()

	doc, _ := vrm.Parse(r, input)
	if _, err := os.Stat(confFile); err != nil {
		log.Print("vrm config error: ", err)
	} else {
		if err = doc.ApplyConfigFile(confFile); err != nil {
			return err
		}
	}
	log.Print("Title: ", doc.Title())
	log.Print("Author: ", doc.Author())
	doc.FixJointComponentType()
	doc.FixJointMatrix()
	if err := doc.ValidateBones(); err != nil {
		log.Print(err)
	}

	log.Print("out: ", output)
	w, err := os.Create(output)
	if err != nil {
		return err
	}
	defer w.Close()
	return vrm.Write(doc, w, output)
}
