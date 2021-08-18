package mqo

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Writer struct {
	path   string
	Create func(name string) (io.WriteCloser, error)
}

func NewWriter(path string) *Writer {
	w := &Writer{path: path}
	if path != "" {
		w.Create = func(name string) (io.WriteCloser, error) {
			return os.Create(filepath.Dir(path) + "/" + name)
		}
	}
	return w
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (writer *Writer) WriteMQO(mqo *Document, ww io.Writer) error {
	w := bufio.NewWriter(ww)
	w.WriteString("Metasequoia Document\n")
	w.WriteString("Format Text Ver 1.1\n")
	w.WriteString("CodePage utf8\n")
	w.WriteString("\n")

	var mqxFile string
	if writer.path != "" && len(mqo.Plugins) > 0 {
		path := writer.path
		mqxFile = filepath.Base(path[0:len(path)-len(filepath.Ext(path))] + ".mqx")
		fmt.Fprintf(w, "IncludeXml \"%v\"\n", mqxFile)
	}

	ex2Count := 0

	fmt.Fprintf(w, "Material %v {\n", len(mqo.Materials))
	for _, mat := range mqo.Materials {
		fmt.Fprintf(w, "\t\"%v\"", mat.Name)
		if mat.Shader != 0 {
			fmt.Fprintf(w, " shader(%d)", mat.Shader)
		}
		if mat.DoubleSided {
			fmt.Fprintf(w, " dbls(%d)", boolToInt(mat.DoubleSided))
		}
		if mat.UID > 0 {
			fmt.Fprintf(w, " uid(%d)", mat.UID)
		}
		fmt.Fprintf(w, " col(%.3f %.3f %.3f %.3f) dif(%.3f) amb(%.3f) emi(%.3f) spc(%.3f) power(%.2f)",
			mat.Color.X, mat.Color.Y, mat.Color.Z, mat.Color.W,
			mat.Diffuse, mat.Ambient, mat.Emission, mat.Specular, mat.Power)
		if mat.Texture != "" {
			fmt.Fprintf(w, " tex(\"%v\")", strings.Replace(mat.Texture, "\\", "/", -1))
		}
		w.WriteString("\n")

		if mat.Ex2 != nil {
			ex2Count++
		}
	}
	w.WriteString("}\n")
	w.Flush()

	if ex2Count > 0 {
		fmt.Fprintf(w, "MaterialEx2 %v {\n", ex2Count)
		for mi, mat := range mqo.Materials {
			if mat.Ex2 == nil {
				continue
			}
			fmt.Fprintf(w, "\tmaterial %v {\n", mi)
			fmt.Fprintf(w, "\t\tshadertype \"%v\"\n", mat.Ex2.ShaderType)
			fmt.Fprintf(w, "\t\tshadername \"%v\"\n", mat.Ex2.ShaderName)
			fmt.Fprintf(w, "\t\tshaderparam %v {\n", len(mat.Ex2.ShaderParams))
			for name, v := range mat.Ex2.ShaderParams {
				typ := "int"
				if b, ok := v.(bool); ok {
					typ = "bool"
					v = boolToInt(b)
				} else if _, ok := v.(float64); ok {
					typ = "float"
				} else if _, ok := v.(float32); ok {
					typ = "float"
				}
				fmt.Fprintf(w, "\t\t\t%v %v %v\n", name, typ, v)
			}
			w.WriteString("\t\t}\n")
			w.WriteString("\t}\n")
		}
		w.WriteString("}\n")
	}

	for _, obj := range mqo.Objects {
		fmt.Fprintf(w, "Object \"%v\" {\n", obj.Name)

		if obj.UID > 0 {
			fmt.Fprintf(w, "\tuid %v\n", obj.UID)
		}
		fmt.Fprintf(w, "\tdepth %d\n", obj.Depth)
		fmt.Fprintf(w, "\tlocking %v\n", boolToInt(obj.Locked))
		if !obj.Visible {
			fmt.Fprint(w, "\tvisible 0\n")
		}
		fmt.Fprintf(w, "\tshading %v\n", obj.Shading)
		fmt.Fprintf(w, "\tfacet %v\n", obj.Facet)
		fmt.Fprintf(w, "\tmirror %d\n", obj.Mirror)
		fmt.Fprintf(w, "\tmirror_dis %f\n", obj.MirrorDis)
		if obj.Patch > 0 {
			fmt.Fprintf(w, "\tpatch %d\n", obj.Patch)
			fmt.Fprintf(w, "\tsegment %d\n", obj.Segment)
		}

		fmt.Fprintf(w, "\tvertex %v {\n", len(obj.Vertexes))
		for _, v := range obj.Vertexes {
			fmt.Fprintf(w, "\t\t%v %v %v\n", v.X, v.Y, v.Z)
		}
		w.WriteString("\t}\n")

		if len(obj.VertexByUID) > 0 {
			w.WriteString("\tvertexattr {\n")
			w.WriteString("\t\tuid {\n")
			uids := make([]int, len(obj.Vertexes))
			for uid, v := range obj.VertexByUID {
				uids[v] = uid
			}
			for i, uid := range uids {
				if uid == 0 {
					uid = i + 1
				}
				fmt.Fprintf(w, "\t\t\t%d\n", uid)
			}
			w.WriteString("\t\t}\n")
			w.WriteString("\t}\n")
		}

		fmt.Fprintf(w, "\tface %v {\n", len(obj.Faces))
		for _, f := range obj.Faces {
			fmt.Fprintf(w, "\t\t%v V(%v) M(%v)", len(f.Verts), strings.Trim(fmt.Sprint(f.Verts), "[]"), f.Material)
			if f.UID > 0 {
				fmt.Fprintf(w, " UID(%v)", f.UID)
			}
			if len(f.UVs) > 0 {
				w.WriteString(" UV(")
				for i, uv := range f.UVs {
					if i != 0 {
						fmt.Fprint(w, " ")
					}
					fmt.Fprintf(w, "%v %v", uv.X, uv.Y)
				}
				w.WriteString(")")
			}
			w.WriteString("\n")
		}
		w.WriteString("\t}\n")

		w.WriteString("}\n")
	}

	w.WriteString("Eof\n")
	w.Flush()

	if mqxFile != "" && writer.Create != nil {
		w, _ := writer.Create(mqxFile)
		defer w.Close()
		WriteMQX(mqo, w, filepath.Base(writer.path))
	}

	return nil
}

func SaveMQOZ(doc *Document, path string) error {
	filename := filepath.Base(path)
	name := filename[0 : len(filename)-len(filepath.Ext(filename))]
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()
	zw := zip.NewWriter(w)
	defer zw.Close()

	mqow, err := zw.Create(name + ".mqo")
	writer := NewWriter(path)
	writer.Create = func(name string) (io.WriteCloser, error) {
		w, err := zw.Create(name)
		return &struct {
			io.Writer
			io.Closer
		}{w, io.NopCloser(nil)}, err
	}
	return writer.WriteMQO(doc, mqow)
}

func Save(doc *Document, path string) error {
	if strings.HasSuffix(path, ".mqoz") {
		return SaveMQOZ(doc, path)
	}
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()
	return NewWriter(path).WriteMQO(doc, w)
}
