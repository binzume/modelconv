package mqo

import "github.com/binzume/modelconv/geom"

type transformable interface {
	ApplyTransform(transform *Matrix4)
}

// Transform object
func (o *Object) ApplyTransform(transform *Matrix4) {
	for i, v := range o.Vertexes {
		o.Vertexes[i] = transform.ApplyTo(v)
	}
	if o.InternalTransform != nil {
		o.InternalTransform = transform.Mul(o.InternalTransform)
	}

	rsmat := transform.Clone()
	rsmat[12], rsmat[13], rsmat[14] = 0, 0, 0 // remove translate
	normalTransform := rsmat.Inverse().Transposed()
	for _, f := range o.Faces {
		for i, n := range f.Normals {
			if n != nil {
				f.Normals[i] = normalTransform.ApplyTo(n)
			}
		}
	}
}

// Transform all objects and plugins
func (doc *Document) ApplyTransform(transform *Matrix4) {
	for _, o := range doc.Objects {
		o.SetLocalTransform(transform.Mul(o.GetLocalTransform()))
		o.ApplyTransform(transform)
	}
	for _, p := range doc.Plugins {
		if tr, ok := p.(transformable); ok {
			tr.ApplyTransform(transform)
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
		for _, tri := range geom.Triangulate(poly) {
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
