package geom

import "math"

type Vector4 struct {
	X Element
	Y Element
	Z Element
	W Element
}

type Quaternion = Vector4

func NewQuaternion(x, y, z, w float32) *Vector4 {
	return &Vector4{X: x, Y: y, Z: z, W: w}
}

func NewQuaternionFromArray(arr [4]Element) *Vector4 {
	return &Vector4{X: arr[0], Y: arr[1], Z: arr[2], W: arr[3]}
}

func (v *Vector4) Add(v2 *Vector4) *Vector4 {
	return &Vector4{X: v.X + v2.X, Y: v.Y + v2.Y, Z: v.Z + v2.Z, W: v.W + v2.W}
}

func (v *Vector4) Sub(v2 *Vector4) *Vector4 {
	return &Vector4{X: v.X - v2.X, Y: v.Y - v2.Y, Z: v.Z - v2.Z, W: v.W - v2.W}
}

func (v *Vector4) Dot(v2 *Vector4) Element {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z + v.W*v2.W
}

func (v *Vector4) Len() Element {
	return Element(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W)))
}

func (v *Vector4) LenSqr() Element {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W
}

func (v *Vector4) Normalize() *Vector4 {
	l := v.Len()
	if l > 0 {
		v.X /= l
		v.Y /= l
		v.Z /= l
		v.W /= l
	} else {
		v.W = 1
	}
	return v
}

func (v *Vector4) Inverse() *Vector4 {
	return &Vector4{X: -v.X, Y: -v.Y, Z: -v.Z, W: v.W}
}
