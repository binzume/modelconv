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
	w.WriteString("\n")

	var mqxPath string
	if path != "" && len(mqo.Plugins) > 0 {
		mqxPath = path[0:len(path)-len(filepath.Ext(path))] + ".mqx"
		fmt.Fprintf(w, "IncludeXml \"%v\"\n", filepath.Base(mqxPath))
	}

	ex2Count := 0

	fmt.Fprintf(w, "Material %v {\n", len(mqo.Materials))
	for _, mat := range mqo.Materials {
		fmt.Fprintf(w, "\t\"%v\"", mat.Name)
		if mat.DoubleSided {
			fmt.Fprintf(w, " dbls(%v)", boolToInt(mat.DoubleSided))
		}
		if mat.UID > 0 {
			fmt.Fprintf(w, " uid(%v)", mat.UID)
		}
		fmt.Fprintf(w, " col(%v %v %v %v) dif(%v) amb(%v) emi(%v) spc(%v) power(%v)",
			mat.Color.X, mat.Color.Y, mat.Color.Z, mat.Color.W,
			mat.Diffuse, mat.Ambient, mat.Emmition, mat.Specular, mat.Power)
		if mat.Texture != "" {
			fmt.Fprintf(w, " tex(\"%v\")", mat.Texture)
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
				}
				fmt.Fprintf(w, "\t\t\t%v %v %v\n", name, typ, v)
			}
			w.WriteString("\t\t}\n")
			w.WriteString("\t}\n")
		}
		w.WriteString("}\n")
		w.Flush()
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

		fmt.Fprintf(w, "\tvertex %v {\n", len(obj.Vertexes))
		for _, v := range obj.Vertexes {
			fmt.Fprintf(w, "\t\t%v %v %v\n", v.X, v.Y, v.Z)
		}
		w.WriteString("\t}\n")

		fmt.Fprintf(w, "\tface %v {\n", len(obj.Faces))
		for _, f := range obj.Faces {
			fmt.Fprintf(w, "\t\t%v V(%v) M(%v)", len(f.Verts), strings.Trim(fmt.Sprint(f.Verts), "[]"), f.Material)
			if f.UID > 0 {
				fmt.Fprintf(w, " UID(%v)", f.UID)
			}
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
