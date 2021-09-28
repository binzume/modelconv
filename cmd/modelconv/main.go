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
	"github.com/binzume/modelconv/fbx"
	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/gltfutil"
	"github.com/binzume/modelconv/mqo"
	"github.com/qmuntal/gltf"
)

var (
	format    = flag.String("format", "", "output file format")
	rot180    = flag.Bool("rot180", false, "rotate 180 degrees around Y (.mqo)")
	autoTpose = flag.String("autotpose", "", "Arm bone names(.mqo)")
	scale     = flag.Float64("scale", 0, "0:auto")
	scaleX    = flag.Float64("scaleX", 1, "scale-x")
	scaleY    = flag.Float64("scaleY", 1, "scale-y")
	scaleZ    = flag.Float64("scaleZ", 1, "scale-z")
	vrmconf   = flag.String("vrmconfig", "", "config file for VRM")
	unlit     = flag.String("unlit", "", "unlit materials (MAT1,MAT2,...)")
	hides     = flag.String("hide", "", "hide objects (OBJ1,OBJ2,...)")
	hidemats  = flag.String("hidemat", "", "hide materials (MAT1,MAT2,...)")
	chparent  = flag.String("chparent", "", "ch bone parent (BONE1:PARENT1,BONE2:PARENT2,...)")

	texReCompress      = flag.Bool("texReCompress", false, "re-compress all textures (gltf)")
	texBytesThreshold  = flag.Int64("texBytesThreshold", 0, "resize large textures (gltf)")
	texResolutionLimit = flag.Int("texResolutionLimit", 4096, "resize large textures (gltf)")
	texResizeScale     = flag.Float64("texResizeScale", 1.0, "resize large textures (gltf)")
	reuseGeometry      = flag.Bool("reuseGeometry", false, "use shared geometry data (gltf, experimental)")
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

func isUnity(ext string) bool {
	return ext == ".unity" || ext == ".prefab" || ext == ".unitypackage"
}

func defaultOutputExt(inputExt string) string {
	if isMQO(inputExt) {
		return ".glb"
	}
	return ".mqo"
}

func saveGltfDocument(doc *gltf.Document, output, ext, srcDir, vrmConf string) error {
	if ext == ".glb" {
		err := gltfutil.ToSingleFile(doc, srcDir)
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
		vrmdoc, err := converter.ToVRM(doc, output, srcDir, vrmConf)
		if err := vrmdoc.ValidateBones(); err != nil {
			log.Print(err)
		}
		if err != nil {
			gltf.SaveBinary(doc, output) // for debug
			return err
		}
		return gltf.SaveBinary(doc, output)
	}
	return fmt.Errorf("Unsuppored output type: %v", ext)
}

func saveDocument(doc *mqo.Document, output, ext, srcDir, vrmConf string, inputs []string) error {
	if isGltf(ext) {
		opt := &converter.MQOToGLTFOption{
			TextureReCompress:      *texReCompress,
			TextureBytesThreshold:  *texBytesThreshold,
			TextureResolutionLimit: *texResolutionLimit,
			TextureScale:           float32(*texResizeScale),
			ReuseGeometry:          *reuseGeometry,
		}
		conv := converter.NewMQOToGLTFConverter(opt)
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
		fmt.Fprintf(os.Stderr, "Usage: %s [options] input.pmx [output.mqo]\n", os.Args[0])
		flag.PrintDefaults()
	}

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

	if outputExt == ".vrm" && *vrmconf == "" {
		conf := input[0:len(input)-len(filepath.Ext(input))] + ".vrmconfig.json"
		if _, err := os.Stat(conf); err == nil {
			*vrmconf = conf
		}
	}

	if isMMD(inputExt) && outputExt == ".vrm" {
		*rot180 = true
		if *vrmconf == "" {
			*vrmconf = "mmd" // preset
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

	// glTF to glTF
	if isGltf(inputExt) && isGltf(outputExt) {
		doc, err := gltfutil.Load(input)
		if err != nil {
			log.Fatal(err)
		}
		gltfutil.ApplyTransform(doc, geom.NewScaleMatrix4(scaleVec.X, scaleVec.Y, scaleVec.Z))
		err = saveGltfDocument(doc, output, outputExt, filepath.Dir(input), *vrmconf)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// FBX to FBX
	if inputExt == ".fbx" && outputExt == ".fbx" {
		doc, err := fbx.Load(input)
		if err != nil {
			log.Fatal(err)
		}
		err = fbx.Save(doc, output) // Save ASCII file
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
	if scaleVec != nil {
		doc.ApplyTransform(geom.NewScaleMatrix4(scaleVec.X, scaleVec.Y, scaleVec.Z))
		if *rot180 {
			for _, b := range mqo.GetBonePlugin(doc).Bones() {
				b.RotationOffset.Y = math.Pi
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

	if *autoTpose != "" {
		patterns := strings.Split(*autoTpose, ",")
		for _, b := range mqo.GetBonePlugin(doc).Bones() {
			if match(patterns, b.Name) {
				*reuseGeometry = false
				doc.BoneAdjustX(b)
			}
		}
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
				mat.Name += "$IGNORE"
				mat.Color.W = 0
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

	if isUnity(inputExt) {
		input = strings.Split(input, "#")[0]
	}
	baseDir := filepath.Dir(input)

	log.Print("out: ", output)
	if err = saveDocument(doc, output, outputExt, baseDir, *vrmconf, flag.Args()[0:inputN]); err != nil {
		log.Fatal(err)
	}
}
