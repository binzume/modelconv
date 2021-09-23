package mqo

type transformable interface {
	Transform(func(v *Vector3))
}

// Transform object
func (o *Object) Transform(transform func(v *Vector3)) {
	for _, v := range o.Vertexes {
		transform(v)
	}
	for _, f := range o.Faces {
		for _, n := range f.Normals {
			if n != nil {
				transform(n)
			}
		}
	}
}

// Transform all objects and plugins
func (doc *Document) Transform(transform func(v *Vector3)) {
	for _, o := range doc.Objects {
		o.Transform(transform)
	}
	for _, p := range doc.Plugins {
		if tr, ok := p.(transformable); ok {
			tr.Transform(transform)
		}
	}
}

func (obj *Object) GetSmoothNormals() []Vector3 {
	normal := make([]Vector3, len(obj.Vertexes))

	for _, face := range obj.Faces {
		for i, v := range face.Verts {
			if len(face.Normals) > i && face.Normals[i] != nil {
				normal[v] = *normal[v].Add(face.Normals[i])
				continue
			}
			v1 := obj.Vertexes[face.Verts[(i+len(face.Verts)-1)%len(face.Verts)]].Sub(obj.Vertexes[v])
			v2 := obj.Vertexes[face.Verts[(i+1)%len(face.Verts)]].Sub(obj.Vertexes[v])
			cross := Vector3{X: v1.Y*v2.Z - v1.Z*v2.Y, Y: v1.Z*v2.X - v1.X*v2.Z, Z: v1.X*v2.Y - v1.Y*v2.X}
			cross.Normalize()
			normal[v] = *normal[v].Add(&cross)
		}
	}
	for i := range normal {
		normal[i].Normalize()
	}
	return normal
}

func (obj *Object) FixhNormals() {
	var normals []Vector3
	for _, f := range obj.Faces {
		if len(f.Normals) == len(f.Verts) {
			continue
		}
		if normals == nil {
			normals = obj.GetSmoothNormals()
		}
		f.Normals = make([]*Vector3, len(f.Verts))
		for i, v := range f.Verts {
			f.Normals[i] = &normals[v]
		}
	}
}

func IsInTriangle(p, a, b, c *Vector3) bool {
	ab, bc, ca := b.Sub(a), c.Sub(b), a.Sub(c)
	c1, c2, c3 := ab.Cross(p.Sub(a)), bc.Cross(p.Sub(b)), ca.Cross(p.Sub(c))
	return c1.Dot(c2) > 0 && c2.Dot(c3) > 0 && c3.Dot(c1) > 0
}

func Triangulate(poly []*Vector3) [][3]int {
	var dst [][3]int
	if len(poly) < 3 {
		return dst
	}
	n := &Vector3{}
	ii := make([]int, len(poly))
	for i := range poly {
		ii[i] = i
		v0 := poly[(i+len(poly)-1)%len(poly)]
		v1 := poly[i]
		v2 := poly[(i+1)%len(poly)]
		n = n.Add(v0.Sub(v1).Cross(v2.Sub(v1)))
	}
	n = n.Normalize()

	// O(N*N)...
	count := len(ii)
	for count >= 3 {
		last_count := count
		for i := count - 1; i >= 0; i-- {
			i0 := ii[(i+count-1)%count]
			i1 := ii[i]
			i2 := ii[(i+1)%count]
			v0 := poly[i0]
			v1 := poly[i1]
			v2 := poly[i2]
			if v0.Sub(v1).Cross(v2.Sub(v1)).Dot(n) >= 0 {
				ok := true
				var tmp []int
				tmp = append(tmp, ii[:i]...)
				tmp = append(tmp, ii[i+1:]...)
				for _, i := range tmp {
					if IsInTriangle(poly[i], v0, v1, v2) {
						ok = false
						break
					}
				}
				if ok {
					dst = append(dst, [3]int{i0, i1, i2})
					ii = tmp
					count--
				}
			}
		}
		if last_count == count {
			// error: maybe self-intersecting polygon
			for i := 0; i < len(ii)-2; i++ {
				dst = append(dst, [3]int{ii[0], ii[i+1], ii[i+2]})
			}
			break
		}
	}
	return dst
}

func (o *Object) Triangulate() {
	var faces []*Face
	for _, f := range o.Faces {
		if len(f.Verts) == 3 {
			faces = append(faces, f)
			continue
		}
		var poly []*Vector3
		for _, v := range f.Verts {
			poly = append(poly, o.Vertexes[v])
		}
		for _, tri := range Triangulate(poly) {
			face := &Face{
				Verts:    []int{f.Verts[tri[0]], f.Verts[tri[1]], f.Verts[tri[2]]},
				Material: f.Material,
			}
			if len(f.UVs) > 0 {
				face.UVs = []Vector2{f.UVs[tri[0]], f.UVs[tri[1]], f.UVs[tri[2]]}
			}
			if len(f.Normals) > 0 {
				face.Normals = []*Vector3{f.Normals[tri[0]], f.Normals[tri[1]], f.Normals[tri[2]]}
			}
			faces = append(faces, face)
		}
	}
	o.Faces = faces
}
