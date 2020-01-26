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
func (doc *MQODocument) Transform(transform func(v *Vector3)) {
	for _, o := range doc.Objects {
		o.Transform(transform)
	}
	for _, p := range doc.Plugins {
		if tr, ok := p.(transformable); ok {
			tr.Transform(transform)
		}
	}
}
