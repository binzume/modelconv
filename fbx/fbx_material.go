package fbx

import "github.com/binzume/modelconv/geom"

type Material struct {
	Obj
}

func NewMaterial(name string) *Material {
	mat := &Material{
		Obj: *newObj("Material", name+"\x00\x01Material", "", []*Node{
			NewNode("Version", 102),
		}),
	}
	mat.SetStringProperty("ShadingModel", "phong")
	return mat
}

func (m *Material) GetColor(name string, def *geom.Vector3) *geom.Vector3 {
	if def == nil {
		def = &geom.Vector3{}
	}
	return m.GetProperty(name).ToVector3(def.X, def.Y, def.Z)
}

func (m *Material) SetColor(name string, c *geom.Vector3) {
	m.SetColorProperty(name, c.X, c.Y, c.Z)
}

func (m *Material) GetFactor(name string, def float32) float32 {
	return m.GetProperty(name).ToFloat32(def)
}

func (m *Material) GetTexture(name string) *Obj {
	// TOOD: m.GetPropertyRef(name)...
	textures := m.FindRefs("Texture")
	if len(textures) > 0 {
		return textures[0].(*Obj)
	}
	return nil
}
