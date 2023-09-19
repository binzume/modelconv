package unity

import (
	"math"

	"github.com/binzume/modelconv/geom"
)

var UnityMeshes = map[Ref]string{
	{FileID: 10202, GUID: "0000000000000000e000000000000000"}: "Cube",
	{FileID: 10206, GUID: "0000000000000000e000000000000000"}: "Cylinder",
	{FileID: 10207, GUID: "0000000000000000e000000000000000"}: "Sphere",
	{FileID: 10208, GUID: "0000000000000000e000000000000000"}: "Capsule",
	{FileID: 10209, GUID: "0000000000000000e000000000000000"}: "Plane",
	{FileID: 10210, GUID: "0000000000000000e000000000000000"}: "Quad",
}

var UnityShaders = map[Ref]string{
	{FileID: 45, GUID: "0000000000000000f000000000000000"}:      "StandardSpecularSetup",
	{FileID: 46, GUID: "0000000000000000f000000000000000"}:      "Standard",
	{FileID: 47, GUID: "0000000000000000f000000000000000"}:      "AutodeskInteractive ",
	{FileID: 4800000, GUID: "933532a4fcc9baf4fa0491de14d08ed7"}: "URP/Lit ",
	{FileID: 4800000, GUID: "650dd9526735d5b46b79224bc6e94025"}: "URP/Unlit ",
	{FileID: 4800000, GUID: "8d2bb70cbf9db8d4da26e15b26e74248"}: "URP/SimpleLit ",
}

func GetBuiltinMesh(ref *Ref) (vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2, name string) {
	if name, ok := UnityMeshes[*ref]; ok {
		if name == "Cube" {
			vs, faces, uvs = Cube()
		} else if name == "Plane" {
			vs, faces, uvs = Plane()
		} else if name == "Quad" {
			vs, faces, uvs = Quad()
		} else if name == "Sphere" {
			vs, faces, uvs = Sphere(32, 16)
		} else if name == "Cylinder" {
			vs, faces, uvs = Cylinder(32)
		} else if name == "Capsule" {
			vs, faces, uvs = Capsule(32)
		}
		return vs, faces, uvs, name
	}
	return nil, nil, nil, ""
}

func Cube() (vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2) {
	vs = []*geom.Vector3{
		{X: -0.5, Y: -0.5, Z: -0.5},
		{X: 0.5, Y: -0.5, Z: -0.5},
		{X: 0.5, Y: 0.5, Z: -0.5},
		{X: -0.5, Y: 0.5, Z: -0.5},
		{X: -0.5, Y: -0.5, Z: 0.5},
		{X: 0.5, Y: -0.5, Z: 0.5},
		{X: 0.5, Y: 0.5, Z: 0.5},
		{X: -0.5, Y: 0.5, Z: 0.5},
	}
	faces = [][]int{
		{0, 1, 2, 3}, {7, 6, 5, 4},
		{4, 5, 1, 0}, {3, 2, 6, 7},
		{2, 1, 5, 6}, {0, 3, 7, 4},
	}
	uvs = [][]geom.Vector2{
		{{X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}, {X: 1, Y: 0}},
		{{X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}, {X: 1, Y: 0}},
		{{X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}, {X: 1, Y: 0}},
		{{X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}, {X: 1, Y: 0}},
		{{X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}},
		{{X: 0, Y: 1}, {X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}},
	}
	return
}

func Sphere(sh, sv int) (vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2) {
	return sphereInternal(sh, sv, 0, sv, 0)
}

func sphereInternal(sh, sv, t, b, voffset int) (vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2) {
	const r = 0.5
	ofs := voffset
	if t > 2 {
		ofs -= (t - 1) * sh
	}
	for i := t; i <= b; i++ {
		if i == 0 {
			vs = append(vs, &geom.Vector3{X: 0, Y: r, Z: 0})
			ofs += 1
			continue
		} else if i == sv {
			vs = append(vs, &geom.Vector3{X: 0, Y: -r, Z: 0})
			continue
		}
		t := float64(i) / float64(sv) * math.Pi
		y := math.Cos(t) * r
		r2 := math.Sin(t) * r
		for j := 0; j < sh; j++ {
			t2 := float64(j) / float64(sh) * 2 * math.Pi
			vs = append(vs, &geom.Vector3{X: float32(math.Cos(t2) * r2), Y: float32(y), Z: float32(math.Sin(t2) * r2)})
		}
	}
	for i := t; i < b; i++ {
		i1 := (i - 1) * sh
		i2 := (i) * sh
		for j := 0; j < sh; j++ {
			j2 := (j + 1) % sh
			if i == 0 {
				faces = append(faces, []int{ofs - 1, i2 + j + ofs, i2 + j2 + ofs})
				uvs = append(uvs, []geom.Vector2{
					{X: float32(j) / float32(sh), Y: float32(i) / float32(sv)},
					{X: float32(j) / float32(sh), Y: float32(i+1) / float32(sv)},
					{X: float32(j+1) / float32(sh), Y: float32(i+1) / float32(sv)},
				})
			} else if i == sv-1 {
				faces = append(faces, []int{i1 + j + ofs, i2 + ofs, i1 + j2 + ofs})
				uvs = append(uvs, []geom.Vector2{
					{X: float32(j) / float32(sh), Y: float32(i) / float32(sv)},
					{X: float32(j) / float32(sh), Y: float32(i+1) / float32(sv)},
					{X: float32(j+1) / float32(sh), Y: float32(i) / float32(sv)},
				})
			} else {
				faces = append(faces, []int{i1 + j + ofs, i2 + j + ofs, i2 + j2 + ofs, i1 + j2 + ofs})
				uvs = append(uvs, []geom.Vector2{
					{X: float32(j) / float32(sh), Y: float32(i) / float32(sv)},
					{X: float32(j) / float32(sh), Y: float32(i+1) / float32(sv)},
					{X: float32(j+1) / float32(sh), Y: float32(i+1) / float32(sv)},
					{X: float32(j+1) / float32(sh), Y: float32(i) / float32(sv)},
				})
			}
		}
	}
	return
}

func Cylinder(s int) (vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2) {
	const r = 0.5
	var top []int
	var bottom []int
	var topuv []geom.Vector2

	for i := 0; i < s; i++ {
		t := float64(i) / float64(s) * math.Pi * 2
		vs = append(vs,
			&geom.Vector3{X: float32(math.Cos(t) * r), Y: 1, Z: float32(math.Sin(t) * r)},
			&geom.Vector3{X: float32(math.Cos(t) * r), Y: -1, Z: float32(math.Sin(t) * r)})
		top = append(top, i*2)
		bottom = append(bottom, (s-i-1)*2+1)
		faces = append(faces, []int{i * 2, i*2 + 1, ((i+1)%s)*2 + 1, ((i + 1) % s) * 2})
		uvs = append(uvs, []geom.Vector2{
			{X: 1 - float32(i)/float32(s), Y: 0},
			{X: 1 - float32(i)/float32(s), Y: 1},
			{X: 1 - float32(i+1)/float32(s), Y: 1},
			{X: 1 - float32(i+1)/float32(s), Y: 0},
		})
		topuv = append(topuv, geom.Vector2{X: float32(i) / float32(s), Y: 1})
	}
	faces = append(faces, top, bottom)
	uvs = append(uvs, topuv, topuv)
	return
}

func Capsule(s int) (vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2) {
	const r = 0.5
	const h = 1.0

	// cap
	vs1, faces1, uvs1 := sphereInternal(s, 8, 0, 4, len(vs))
	for _, v := range vs1 {
		v.Y += h / 2
	}
	vs = append(vs, vs1...)
	faces = append(faces, faces1...)
	uvs = append(uvs, uvs1...)

	st := len(vs) - s
	for i := 0; i < s; i++ {
		faces = append(faces, []int{st + i, st + i + s, st + (i+1)%s + s, st + (i+1)%s})
		uvs = append(uvs, []geom.Vector2{
			{X: 1 - float32(i)/float32(s), Y: 0},
			{X: 1 - float32(i)/float32(s), Y: 1},
			{X: 1 - float32(i+1)/float32(s), Y: 1},
			{X: 1 - float32(i+1)/float32(s), Y: 0},
		})
	}
	vs1, faces1, uvs1 = sphereInternal(s, 8, 4, 8, len(vs))
	for _, v := range vs1 {
		v.Y -= h / 2
	}
	vs = append(vs, vs1...)
	faces = append(faces, faces1...)
	uvs = append(uvs, uvs1...)
	return
}

func Quad() (vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2) {
	vs = []*geom.Vector3{
		{X: -0.5, Y: -0.5},
		{X: 0.5, Y: -0.5},
		{X: -0.5, Y: 0.5},
		{X: 0.5, Y: 0.5},
	}
	faces = [][]int{
		{1, 0, 2, 3},
	}
	uvs = [][]geom.Vector2{
		{{X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}, {X: 1, Y: 0}},
	}
	return
}

func Plane() (vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2) {
	vs = []*geom.Vector3{
		{X: -5, Y: 0, Z: -5},
		{X: 5, Y: 0, Z: -5},
		{X: 5, Y: 0, Z: 5},
		{X: -5, Y: 0, Z: 5},
	}
	faces = [][]int{
		{0, 1, 2, 3},
	}
	uvs = [][]geom.Vector2{
		{{X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}, {X: 1, Y: 0}},
	}
	return
}
