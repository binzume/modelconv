package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/converter"
	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/mqo"
	"github.com/binzume/modelconv/vrm"
	"github.com/qmuntal/gltf"
)

func isGltf(ext string) bool {
	return ext == ".glb" || ext == ".gltf" || ext == ".vrm"
}

func isMMD(ext string) bool {
	return ext == ".pmx" || ext == ".pmd"
}

func isMQO(ext string) bool {
	return ext == ".mqo" || ext == ".mqoz"
}

func defaultOutputExt(inputExt string) string {
	if isGltf(inputExt) {
		return ".vrm"
	} else if isMQO(inputExt) {
		return ".glb"
	}
	return ".mqo"
}

func saveGltfDocument(doc *gltf.Document, output, ext, srcDir, vrmConf string) error {
	if ext == ".glb" {
		err := vrm.ToSingleFile(doc, srcDir)
		if err != nil {
			return err
		}
		return gltf.SaveBinary(doc, output)
	} else if ext == ".gltf" {
		for i, b := range doc.Buffers {
			if b.URI == "" {
				b.URI = fmt.Sprintf("%s%d.bin", strings.TrimSuffix(filepath.Base(output), filepath.Ext(output)), i)
			}
		}
		return gltf.Save(doc, output)
	} else if ext == ".vrm" {
		return saveAsVRM(doc, output, srcDir, vrmConf)
	}
	return fmt.Errorf("Unsuppored output type: %v", ext)
}

func saveDocument(doc *mqo.Document, output, srcDir, vrmConf string, inputs []string) error {
	ext := strings.ToLower(filepath.Ext(output))
	if isGltf(ext) {
		conv := converter.NewMQOToGLTFConverter(&converter.MQOToGLTFOption{})
		gltfdoc, err := conv.Convert(doc, srcDir)
		if err != nil {
			return err
		}

		for _, f := range inputs[1:] {
			if strings.ToLower(filepath.Ext(f)) == ".vmd" {
				ani, err := loadAnimation(f)
				if err != nil {
					return err
				}
				converter.AddAnimationTpGlb(gltfdoc, ani, conv.JointNodeToBone, true)
			}
		}
		return saveGltfDocument(gltfdoc, output, ext, srcDir, vrmConf)
	} else if isMQO(ext) {
		return mqo.Save(doc, output)
	} else if ext == ".pmx" {
		return saveAsPmx(doc, output)
	}
	return fmt.Errorf("Unsuppored output type: %v", ext)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s input.pmx [output.mqo]\n", os.Args[0])
		flag.PrintDefaults()
	}
	format := flag.String("format", "", "output file format")
	rot180 := flag.Bool("rot180", false, "rotate 180 degrees around Y (.mqo)")
	autoTpose := flag.String("autotpose", "", "Arm bone names(.mqo)")
	scale := flag.Float64("scale", 0, "0:auto")
	scaleX := flag.Float64("scaleX", 1, "scale-x")
	scaleY := flag.Float64("scaleY", 1, "scale-y")
	scaleZ := flag.Float64("scaleZ", 1, "scale-z")
	vrmconf := flag.String("vrmconfig", "", "config file for VRM")
	unlit := flag.String("unlit", "", "unlit materials (MAT1,MAT2,...)")
	hides := flag.String("hide", "", "hide objects (OBJ1,OBJ2,...)")
	hidemats := flag.String("hidemat", "", "hide materials (MAT1,MAT2,...)")
	chparent := flag.String("chparent", "", "ch bone parent (BONE1:PARENT1,BONE2:PARENT2,...)")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}
	input := flag.Arg(0)
	inputExt := strings.ToLower(filepath.Ext(input))
	output := ""
	outputExt := "." + *format
	inputN := flag.NArg() - 1
	if inputN < 1 {
		inputN = 1
		if outputExt == "." {
			outputExt = defaultOutputExt(inputExt)
		}
		output = input[0:len(input)-len(inputExt)] + outputExt
	} else {
		output = flag.Arg(inputN)
		if outputExt == "." {
			outputExt = strings.ToLower(filepath.Ext(output))
		}
	}
	confFile := *vrmconf
	if confFile == "" {
		confFile = input[0:len(input)-len(filepath.Ext(input))] + ".vrmconfig.json"
		if _, err := os.Stat(confFile); err != nil {
			confFile = ""
		}
	}

	// mmd to vrm
	if isMMD(inputExt) && outputExt == ".vrm" {
		*rot180 = true
		if confFile == "" {
			confFile = "mmd" // preset
		}
		if *autoTpose == "" {
			*autoTpose = "右腕,左腕"
		}
	}

	if *scale == 0 {
		if isMMD(inputExt) {
			*scale = 80
		} else {
			*scale = 1
		}
	}
	var scaleVec = &geom.Vector3{
		X: float32(*scale * *scaleX),
		Y: float32(*scale * *scaleY),
		Z: float32(*scale * *scaleZ),
	}
	if *rot180 {
		scaleVec.X *= -1
		scaleVec.Z *= -1
	}
	if scaleVec.X == 1 && scaleVec.Y == 1 && scaleVec.Z == 1 {
		scaleVec = nil
	}

	if isGltf(inputExt) && isGltf(outputExt) {
		doc, err := gltf.Open(input)
		if err != nil {
			log.Fatal(err)
		}
		vrm.Transform(doc, scaleVec, nil)
		// vrm.ResetJointMatrix(doc)
		err = saveGltfDocument(doc, output, outputExt, filepath.Dir(input), confFile)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if inputExt == ".vmd" {
		// DEBUG
		anim, err := loadAnimation(input)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(anim.Name)
		log.Println(anim.GetBoneChannels())
		return
	}

	doc, err := loadDocument(input)
	if err != nil {
		log.Fatal(err)
	}

	// transform
	if scaleVec != nil {
		doc.Transform(func(v *mqo.Vector3) {
			v.X *= scaleVec.X
			v.Y *= scaleVec.Y
			v.Z *= scaleVec.Z
		})
	}
	if *rot180 {
		for _, b := range mqo.GetBonePlugin(doc).Bones() {
			b.RotationOffset.Y = math.Pi
		}
	}

	if *autoTpose != "" {
		for _, boneName := range strings.Split(*autoTpose, ",") {
			for _, b := range mqo.GetBonePlugin(doc).Bones() {
				if b.Name == boneName {
					doc.BoneAdjustX(b)
				}
			}
		}
	}

	match := func(patterns []string, name string) bool {
		for _, pattern := range patterns {
			if m, _ := path.Match(pattern, name); m {
				return true
			}
		}
		return false
	}

	if *hides != "" {
		patterns := strings.Split(*hides, ",")
		for idx, obj := range doc.Objects {
			if match(patterns, obj.Name) {
				obj.Visible = false
				d := obj.Depth
				for i := idx + 1; i < len(doc.Objects) && doc.Objects[i].Depth > d; i++ {
					doc.Objects[i].Visible = false
				}
			}
		}
	}

	if *hidemats != "" {
		patterns := strings.Split(*hidemats, ",")
		for _, mat := range doc.Materials {
			if match(patterns, mat.Name) {
				mat.Name = "$IGNORE"
			}
		}
	}

	if *unlit != "" {
		patterns := strings.Split(*unlit, ",")
		for _, mat := range doc.Materials {
			if match(patterns, mat.Name) {
				mat.Shader = 1 // Constant shader
				mat.Ex2 = nil
			}
		}
	}

	if *chparent != "" {
		boneByName := map[string]*mqo.Bone{}
		for _, b := range mqo.GetBonePlugin(doc).Bones() {
			boneByName[b.Name] = b
		}
		for _, n := range strings.Split(*chparent, ",") {
			boneAndParent := strings.Split(n, ":")
			if len(boneAndParent) != 2 {
				log.Fatal("invalid bone setting (BONE_NAME:PARENT_NAME)", n)
				continue
			}
			if b, ok := boneByName[boneAndParent[0]]; ok {
				b.Parent = boneByName[boneAndParent[1]].ID
			}
		}
	}

	log.Print("out: ", output)
	if err = saveDocument(doc, output, filepath.Dir(input), confFile, flag.Args()[0:inputN]); err != nil {
		log.Fatal(err)
	}
}
