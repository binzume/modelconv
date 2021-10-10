package fbx

import (
	"math"

	"github.com/binzume/modelconv/geom"
)

type Model struct {
	Obj
	Parent       *Model
	cachedMatrix *geom.Matrix4
}

func NewModel(name, kind string) *Model {
	model := &Model{
		Obj: *newObj("Model", name+"\x00\x01Model", kind, []*Node{
			NewNode("Version", 232),
		}),
	}
	model.SetStringProperty("Culling", "CullingOff")
	model.SetStringProperty("Culling", "CullingOff")
	return model
}

func (m *Model) GetTranslation() *geom.Vector3 {
	return m.GetProperty("Lcl Translation").ToVector3(0, 0, 0)
}

func (m *Model) SetTranslation(v *geom.Vector3) {
	m.SetProperty("Lcl Translation", &Property{Type: "Lcl Translation", Flag: "A+", AttributeList: []*Attribute{{Value: v.X}, {Value: v.Y}, {Value: v.Z}}})
}

func (m *Model) GetRotation() *geom.Vector3 {
	return m.GetProperty("Lcl Rotation").ToVector3(0, 0, 0) // Euler defgrees. XYZ order?
}

func (m *Model) SetRotation(v *geom.Vector3) {
	m.SetProperty("Lcl Rotation", &Property{Type: "Lcl Rotation", Flag: "A+", AttributeList: []*Attribute{{Value: v.X}, {Value: v.Y}, {Value: v.Z}}})
}

func (m *Model) GetScaling() *geom.Vector3 {
	return m.GetProperty("Lcl Scaling").ToVector3(1, 1, 1)
}

func (m *Model) SetScaling(v *geom.Vector3) {
	m.SetProperty("Lcl Scaling", &Property{Type: "Lcl Scaling", Flag: "A+", AttributeList: []*Attribute{{Value: v.X}, {Value: v.Y}, {Value: v.Z}}})
}

func (m *Model) UpdateMatrix() {
	// TODO: apply pivot
	prerotEuler := m.GetProperty("PreRotation").ToVector3(0, 0, 0).Scale(math.Pi / 180)
	prerot := geom.NewEulerRotationMatrix4(prerotEuler.X, prerotEuler.Y, prerotEuler.Z, 1)
	translation := m.GetTranslation()
	rotationEuler := m.GetRotation().Scale(math.Pi / 180)
	scale := m.GetScaling()
	tr := geom.NewTranslateMatrix4(translation.X, translation.Y, translation.Z)
	rot := geom.NewEulerRotationMatrix4(rotationEuler.X, rotationEuler.Y, rotationEuler.Z, 1)
	sacle := geom.NewScaleMatrix4(scale.X, scale.Y, scale.Z)
	m.cachedMatrix = tr.Mul(prerot).Mul(rot).Mul(sacle)
}

func (m *Model) GetMatrix() *geom.Matrix4 {
	if m.cachedMatrix == nil {
		m.UpdateMatrix()
	}
	return m.cachedMatrix
}

func (m *Model) GetWorldMatrix() *geom.Matrix4 {
	if m.Parent == nil {
		return m.GetMatrix()
	}
	return m.Parent.GetWorldMatrix().Mul(m.GetMatrix())
}

func (m *Model) GetChildModels() []*Model {
	var r []*Model
	for _, o := range m.Refs {
		if m, ok := o.(*Model); ok {
			r = append(r, m)
		}
	}
	return r
}

func (m *Model) GetGeometry() *Geometry {
	for _, o := range m.Refs {
		if g, ok := o.(*Geometry); ok {
			return g
		}
	}
	return nil
}
