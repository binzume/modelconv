package mqo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func WriteMQO(mqo *MQODocument, ww io.Writer, path string) error {
	w := bufio.NewWriter(ww)
	w.WriteString("Metasequoia Document\n")
	w.WriteString("Format Text Ver 1.1\n")
	w.WriteString("CodePage utf8\n")

	var mqxPath string
	if path != "" && len(mqo.Plugins) > 0 {
		mqxPath = path[0:len(path)-len(filepath.Ext(path))] + ".mqx"
		fmt.Fprintf(w, "IncludeXml \"%v\"\n", filepath.Base(mqxPath))
	}

	ex2 := map[int]*MaterialEx2{}

	fmt.Fprintf(w, "Material %v {\n", len(mqo.Materials))
	for i, mat := range mqo.Materials {
		fmt.Fprintf(w, "\t\"%v\" col(%v %v %v %v) dif(%v) amb(%v) emi(%v) spc(%v) power(%v)",
			mat.Name,
			mat.Color.X, mat.Color.Y, mat.Color.Z, mat.Color.W,
			mat.Diffuse, mat.Ambient, mat.Emmition, mat.Specular, mat.Power)
		if mat.DoubleSided {
			fmt.Fprintf(w, " dbls(%v)", boolToInt(mat.DoubleSided))
		}
		if mat.Texture != "" {
			fmt.Fprintf(w, " tex(\"%v\")", mat.Texture)
		}
		w.WriteString("\n")

		if mat.Ex2 != nil {
			ex2[i] = mat.Ex2
		}
	}
	w.WriteString("}\n")
	w.Flush()

	if len(ex2) > 0 {
		fmt.Fprintf(w, "MaterialEx2 %v {\n", len(ex2))
		for mi, mat := range ex2 {
			fmt.Fprintf(w, "material %v {\n", mi)
			fmt.Fprintf(w, "\tshadertype \"%v\"\n", mat.ShaderType)
			fmt.Fprintf(w, "\tshadername \"%v\"\n", mat.ShaderName)
			fmt.Fprintf(w, "\tshaderparam %v {\n", len(mat.ShaderParams))
			for name, v := range mat.ShaderParams {
				typ := "int"
				if _, ok := v.(bool); ok {
					typ = "bool"
				}
				fmt.Fprintf(w, "\t\t%v %v %v\n", name, typ, v)
			}
			w.WriteString("\t}\n")
			w.WriteString("}\n")
		}
		w.WriteString("}\n")
		w.Flush()
	}

	for _, obj := range mqo.Objects {
		fmt.Fprintf(w, "Object \"%v\" {\n", obj.Name)

		fmt.Fprintf(w, "\tdepth %d\n", obj.Depth)
		if !obj.Visible {
			fmt.Fprint(w, "\tvisible 0\n")
		}
		if obj.Locked {
			fmt.Fprint(w, "\tlocking 1\n")
		}

		fmt.Fprintf(w, "\tvertex %v {\n", len(obj.Vertexes))
		for _, v := range obj.Vertexes {
			fmt.Fprintf(w, "\t\t%v %v %v\n", v.X, v.Y, v.Z)
		}
		w.WriteString("\t}\n")

		fmt.Fprintf(w, "\tface %v {\n", len(obj.Faces))
		for _, f := range obj.Faces {
			fmt.Fprintf(w, "\t\t%v V(%v) M(%v)", len(f.Verts), strings.Trim(fmt.Sprint(f.Verts), "[]"), f.Material)
			if len(f.UVs) > 0 {
				w.WriteString(" UV(")
				for _, uv := range f.UVs {
					fmt.Fprintf(w, "%v %v ", uv.X, uv.Y)
				}
				w.WriteString(")")
			}
			w.WriteString("\n")
		}
		w.WriteString("\t}\n")

		w.WriteString("}\n")
		w.Flush()
	}

	w.WriteString("Eof\n")
	w.Flush()

	if mqxPath != "" && len(mqo.Plugins) > 0 {
		w, _ := os.Create(mqxPath)
		defer w.Close()
		WriteMQX(mqo, w, filepath.Base(path))
	}

	return nil
}
