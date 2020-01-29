package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	// github.com/binzume/modelconv
	"../../mmd"
	"../../mqo"
	"../../vrm"
)

func glb2vrm(r io.Reader, input, output, confFile string) {
	doc, _ := vrm.Parse(r, input)
	if _, err := os.Stat(confFile); err != nil {
		log.Print("vrm config error: ", err)
	} else {
		if err = doc.ApplyConfigFile(confFile); err != nil {
			log.Fatal("vrm config error: ", err)
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
		log.Fatal(err)
	}
	defer w.Close()
	if err := vrm.Write(doc, w, output); err != nil {
		log.Fatal(err)
	}
}

func main() {
	rot180 := flag.Bool("rot180", false, "rotation 180 degrees around Y (.mqo)")
	scale := flag.Float64("scale", 1.0, "scale (.mqo)")
	conf := flag.String("config", "", "config file for convertion")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Usage: modelconv input.pmx [output.mqo]")
		return
	}
	input := flag.Arg(0)
	output := flag.Arg(1)

	r, err := os.Open(input)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	var doc *mqo.MQODocument

	if strings.HasSuffix(input, ".mqo") {
		if doc, err = mqo.Parse(r, input); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(input, ".glb") {
		if output == "" {
			output = input[0:len(input)-len(filepath.Ext(input))] + ".vrm"
		}
		confFile := *conf
		if confFile == "" {
			confFile = input[0:len(input)-len(filepath.Ext(input))] + ".vrmconfig.json"
		}
		glb2vrm(r, input, output, confFile)
		return
	} else {
		pmx, err := mmd.Parse(r)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Name", pmx.Name)
		log.Println("Comment", pmx.Comment)
		doc = PMX2MQO(pmx)
	}

	if *rot180 {
		doc.Transform(func(v *mqo.Vector3) {
			v.X *= -1
			v.Z *= -1
		})
	}
	if *scale != 1.0 {
		s := float32(*scale)
		doc.Transform(func(v *mqo.Vector3) {
			v.X *= s
			v.Y *= s
			v.Z *= s
		})
	}
	if output == "" {
		output = input + ".mqo"
	}

	log.Print("out: ", output)
	w, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()
	if err = mqo.WriteMQO(doc, w, output); err != nil {
		log.Fatal(err)
	}
}
