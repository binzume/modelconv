package geom

import "math"

type Element = float32

type Vector3 struct {
	X Element
	Y Element
	Z Element
}

func NewVector3(x, y, z float32) *Vector3 {
	return &Vector3{X: x, Y: y, Z: z}
}

func NewVector3FromArray(arr [3]Element) *Vector3 {
	return &Vector3{X: arr[0], Y: arr[1], Z: arr[2]}
}

func NewVector3FromSlice(arr []Element) *Vector3 {
	return &Vector3{X: arr[0], Y: arr[1], Z: arr[2]}
}

func (v *Vector3) Add(v2 *Vector3) *Vector3 {
	return &Vector3{X: v.X + v2.X, Y: v.Y + v2.Y, Z: v.Z + v2.Z}
}

func (v *Vector3) Sub(v2 *Vector3) *Vector3 {
	return &Vector3{X: v.X - v2.X, Y: v.Y - v2.Y, Z: v.Z - v2.Z}
}

func (v *Vector3) Dot(v2 *Vector3) Element {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z
}

func (v *Vector3) Cross(v2 *Vector3) *Vector3 {
	return &Vector3{
		X: v.Y*v2.Z - v.Z*v2.Y,
		Y: v.Z*v2.X - v.X*v2.Z,
		Z: v.X*v2.Y - v.Y*v2.X,
	}
}

func (v *Vector3) Scale(s Element) *Vector3 {
	return &Vector3{X: v.X * s, Y: v.Y * s, Z: v.Z * s}
}

func (v *Vector3) Len() Element {
	return Element(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

func (v *Vector3) LenSqr() Element {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v *Vector3) Normalize() *Vector3 {
	l := v.Len()
	if l > 0 {
		v.X /= l
		v.Y /= l
		v.Z /= l
	} else {
		v.X = 1
	}
	return v
}

func (v *Vector3) ToArray(array []Element) {
	array[0] = v.X
	array[1] = v.Y
	array[2] = v.Z
}

func (mat *Matrix4) ApplyTo(v *Vector3) *Vector3 {
	return &Vector3{
		mat[0]*v.X + mat[4]*v.Y + mat[8]*v.Z + mat[12],
		mat[1]*v.X + mat[5]*v.Y + mat[9]*v.Z + mat[13],
		mat[2]*v.X + mat[6]*v.Y + mat[10]*v.Z + mat[14],
	}
}
