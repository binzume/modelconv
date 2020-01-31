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

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s input.pmx [output.mqo]\n", os.Args[0])
		flag.PrintDefaults()
	}
	rot180 := flag.Bool("rot180", false, "rotate 180 degrees around Y (.mqo)")
	scale := flag.Float64("scale", 1.0, "scale (.mqo)")
	vrmconf := flag.String("vrmconfig", "", "config file for VRM")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}
	input := flag.Arg(0)
	output := flag.Arg(1)

	if strings.ToLower(filepath.Ext(input)) == ".glb" {
		err := glb2vrm(input, output, *vrmconf)
		if err != nil {
			log.Fatal(err)
		}
		return
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

	if output == "" {
		output = input + ".mqo"
	}
	log.Print("out: ", output)
	if err = saveDocument(doc, output); err != nil {
		log.Fatal(err)
	}
}
