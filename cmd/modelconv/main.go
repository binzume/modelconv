package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/mqo"
)

func defaultOutputFile(input string) string {
	ext := strings.ToLower(filepath.Ext(input))
	base := input[0 : len(input)-len(ext)]
	if ext == ".glb" {
		return base + ".vrm"
	} else if ext == ".mqo" {
		return base + ".glb"
	} else if ext == ".pmx" || ext == ".pmd" {
		return base + ".mqo"
	}
	return input + ".mqo"
}

func saveDocument(doc *mqo.MQODocument, output, srcDir, vrmConf string) error {
	ext := strings.ToLower(filepath.Ext(output))
	if ext == ".glb" {
		return saveAsGlb(doc, output, srcDir)
	} else if ext == ".vrm" {
		gltfdoc, err := NewMQO2GLTF().Convert(doc, srcDir)
		if err != nil {
			return err
		}
		return saveAsVRM(gltfdoc, output, vrmConf)
	} else if ext == ".mqo" {
		w, err := os.Create(output)
		if err != nil {
			return err
		}
		defer w.Close()
		return mqo.WriteMQO(doc, w, output)
	}
	return fmt.Errorf("Unsuppored output type: %v", ext)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s input.pmx [output.mqo]\n", os.Args[0])
		flag.PrintDefaults()
	}
	rot180 := flag.Bool("rot180", false, "rotate 180 degrees around Y (.mqo)")
	scale := flag.Float64("scale", 0, "0:auto")
	vrmconf := flag.String("vrmconfig", "", "config file for VRM")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}
	input := flag.Arg(0)
	output := flag.Arg(1)
	if output == "" {
		output = defaultOutputFile(input)
	}
	confFile := *vrmconf
	if confFile == "" {
		confFile = input[0:len(input)-len(filepath.Ext(input))] + ".vrmconfig.json"
	}

	inputExt := strings.ToLower(filepath.Ext(input))
	if inputExt == ".glb" {
		err := glb2vrm(input, output, confFile)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if *scale == 0 {
		if inputExt == ".pmx" || inputExt == ".pmd" {
			*scale = 80
		} else {
			*scale = 1
		}
	}

	doc, err := loadDocument(input)
	if err != nil {
		log.Fatal(err)
	}

	// transform
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

	log.Print("out: ", output)
	if err = saveDocument(doc, output, filepath.Dir(input), confFile); err != nil {
		log.Fatal(err)
	}
}
