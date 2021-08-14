package geom

import "math"

type Vector2 struct {
	X float32
	Y float32
}

type Vector3 struct {
	X float32
	Y float32
	Z float32
}

type Vector4 struct {
	X float32
	Y float32
	Z float32
	W float32
}

type Matrix4 [16]float32

func (v *Vector2) Add(v2 *Vector2) *Vector2 {
	return &Vector2{X: v.X + v2.X, Y: v.Y + v2.Y}
}

func (v *Vector2) Sub(v2 *Vector2) *Vector2 {
	return &Vector2{X: v.X - v2.X, Y: v.Y - v2.Y}
}

func (v *Vector2) Dot(v2 *Vector2) float32 {
	return v.X*v2.X + v.Y*v2.Y
}
func (v *Vector2) Cross(v2 *Vector2) float32 {
	return v.X*v2.Y - v.Y*v2.X
}

func (v *Vector2) Len() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

func (v *Vector2) LenSqr() float32 {
	return v.X*v.X + v.Y*v.Y
}

func (v *Vector2) Normalize() *Vector2 {
	l := v.Len()
	if l > 0 {
		v.X /= l
		v.Y /= l
	} else {
		v.Y = 1
	}
	return v
}

func (v *Vector3) Add(v2 *Vector3) *Vector3 {
	return &Vector3{X: v.X + v2.X, Y: v.Y + v2.Y, Z: v.Z + v2.Z}
}

func (v *Vector3) Sub(v2 *Vector3) *Vector3 {
	return &Vector3{X: v.X - v2.X, Y: v.Y - v2.Y, Z: v.Z - v2.Z}
}

func (v *Vector3) Dot(v2 *Vector3) float32 {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z
}

func (v *Vector3) Cross(v2 *Vector3) *Vector3 {
	return &Vector3{
		X: v.Y*v2.Z - v.Z*v2.Y,
		Y: v.Z*v2.X - v.X*v2.Z,
		Z: v.X*v2.Y - v.Y*v2.X,
	}
}

func (v *Vector3) Len() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

func (v *Vector3) LenSqr() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v *Vector3) Normalize() *Vector3 {
	l := v.Len()
	if l > 0 {
		v.X /= l
		v.Y /= l
		v.Z /= l
	} else {
		v.Z = 1
	}
	return v
}

func (v *Vector3) ToArray(array []float32) {
	array[0] = v.X
	array[1] = v.Y
	array[2] = v.Z
}

func (v *Vector4) Dot(v2 *Vector4) float32 {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z + v.W*v2.W
}

func (v *Vector4) Len() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W)))
}

func (v *Vector4) LenSqr() float32 {
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

// Returns Hamilton product
func (a *Vector4) Mul(b *Vector4) *Vector4 {
	return &Vector4{
		W: a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z, // 1
		X: a.W*b.X + a.X*b.W + a.Y*b.Z - a.Z*b.Y, // i
		Y: a.W*b.Y - a.X*b.Z + a.Y*b.W + a.Z*b.X, // j
		Z: a.W*b.Z + a.X*b.Y - a.Y*b.X + a.Z*b.W, // k
	}
}
