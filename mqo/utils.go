package mqo

type transformable interface {
	Transform(func(v *Vector3))
}

// Transform object
func (o *Object) Transform(transform func(v *Vector3)) {
	for _, v := range o.Vertexes {
		transform(v)
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
