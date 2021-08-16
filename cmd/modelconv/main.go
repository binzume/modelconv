package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/binzume/modelconv/converter"
	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/mqo"
	"github.com/binzume/modelconv/vrm"
	"github.com/qmuntal/gltf"
)

func defaultOutputFile(input string) string {
	ext := strings.ToLower(filepath.Ext(input))
	base := input[0 : len(input)-len(ext)]
	if ext == ".glb" || ext == ".gltf" {
		return base + ".vrm"
	} else if ext == ".mqo" {
		return base + ".glb"
	} else if ext == ".pmx" || ext == ".pmd" {
		return base + ".mqo"
	}
	return input + ".mqo"
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

func saveDocument(doc *mqo.Document, output, srcDir, vrmConf string, forceUnlit bool, inputs []string) error {
	ext := strings.ToLower(filepath.Ext(output))
	if ext == ".glb" || ext == ".gltf" || ext == ".vrm" {
		conv := converter.NewMQOToGLTFConverter(&converter.MQOToGLTFOption{ForceUnlit: forceUnlit})
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
	} else if ext == ".mqo" {
		w, err := os.Create(output)
		if err != nil {
			return err
		}
		defer w.Close()
		return mqo.WriteMQO(doc, w, output)
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
	rot180 := flag.Bool("rot180", false, "rotate 180 degrees around Y (.mqo)")
	autoTpose := flag.String("autotpose", "", "Arm bone names(.mqo)")
	forceUnlit := flag.Bool("gltfunlit", false, "unlit all materials")
	scale := flag.Float64("scale", 0, "0:auto")
	scaleX := flag.Float64("scaleX", 1, "scale-x")
	scaleY := flag.Float64("scaleY", 1, "scale-y")
	scaleZ := flag.Float64("scaleZ", 1, "scale-z")
	vrmconf := flag.String("vrmconfig", "", "config file for VRM")
	hides := flag.String("hide", "", "hide objects (OBJ1,OBJ2,...)")
	hidemats := flag.String("hidemat", "", "hide materials (MAT1,MAT2,...)")
	chparent := flag.String("chparent", "", "ch bone parent (BONE1:PARENT1,BONE2:PARENT2,...)")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}
	input := flag.Arg(0)
	output := ""
	inputN := flag.NArg() - 1
	if inputN < 1 {
		inputN = 1
		output = defaultOutputFile(input)
	} else {
		output = flag.Arg(inputN)
	}
	confFile := *vrmconf
	if confFile == "" {
		confFile = input[0:len(input)-len(filepath.Ext(input))] + ".vrmconfig.json"
		if _, err := os.Stat(confFile); err != nil {
			confFile = ""
		}
	}
	inputExt := strings.ToLower(filepath.Ext(input))
	outputExt := strings.ToLower(filepath.Ext(output))

	// mmd to vrm
	if (inputExt == ".pmx" || inputExt == ".pmd") && outputExt == ".vrm" {
		*rot180 = true
		if confFile == "" {
			confFile = "mmd" // preset
		}
		if *autoTpose == "" {
			*autoTpose = "右腕,左腕"
		}
	}

	if *scale == 0 {
		if inputExt == ".pmx" || inputExt == ".pmd" {
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

	if inputExt == ".glb" || inputExt == ".gltf" || inputExt == ".vrm" {
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

	if *hides != "" {
		objectByName := map[string]int{}
		for idx, obj := range doc.Objects {
			objectByName[obj.Name] = idx
		}
		for _, n := range strings.Split(*hides, ",") {
			if idx, ok := objectByName[n]; ok {
				d := doc.Objects[idx].Depth
				doc.Objects[idx].Visible = false
				idx++
				for idx < len(doc.Objects) && doc.Objects[idx].Depth > d {
					doc.Objects[idx].Visible = false
					idx++
				}
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

	if *hidemats != "" {
		materialsByName := map[string]int{}
		for idx, mat := range doc.Materials {
			materialsByName[mat.Name] = idx
		}
		for _, n := range strings.Split(*hidemats, ",") {
			if idx, ok := materialsByName[n]; ok {
				doc.Materials[idx].Name = "$IGNORE"
			}
		}
	}

	log.Print("out: ", output)
	if err = saveDocument(doc, output, filepath.Dir(input), confFile, *forceUnlit, flag.Args()[0:inputN]); err != nil {
		log.Fatal(err)
	}
}
