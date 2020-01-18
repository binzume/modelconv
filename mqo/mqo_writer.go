package mqo

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func WriteMQO(mqo *MQODocument, ww io.Writer, mxqName string) error {
	w := bufio.NewWriter(ww)
	w.WriteString("Metasequoia Document\n")
	w.WriteString("Format Text Ver 1.1\n")
	w.WriteString("CodePage utf8\n")

	if mxqName != "" {
		fmt.Fprintf(w, "IncludeXml %v\n", mxqName)
	}

	fmt.Fprintf(w, "Material %v {\n", len(mqo.Materials))
	for _, mat := range mqo.Materials {
		fmt.Fprintf(w, "\t\"%v\" col(%v %v %v %v) dif(%v) amb(%v) emi(%v) spc(%v) power(%v)",
			mat.Name,
			mat.Color.X, mat.Color.Y, mat.Color.Z, mat.Color.W,
			mat.Diffuse, mat.Ambient, mat.Emmition, mat.Specular, mat.Power)
		if mat.Texture != "" {
			fmt.Fprintf(w, " tex(\"%v\")", mat.Texture)
		}
		w.WriteString("\n")
	}
	w.WriteString("}\n")
	w.Flush()

	for _, obj := range mqo.Objects {
		fmt.Fprintf(w, "Object \"%v\" {\n", obj.Name)

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

	return nil
}
