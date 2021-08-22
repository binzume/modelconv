package geom

import "math"

type Element = float32

type Vector2 struct {
	X Element
	Y Element
}

type Vector3 struct {
	X Element
	Y Element
	Z Element
}

type Vector4 struct {
	X Element
	Y Element
	Z Element
	W Element
}

// column-major matrix
type Matrix4 [16]Element

func (v *Vector2) Add(v2 *Vector2) *Vector2 {
	return &Vector2{X: v.X + v2.X, Y: v.Y + v2.Y}
}

func (v *Vector2) Sub(v2 *Vector2) *Vector2 {
	return &Vector2{X: v.X - v2.X, Y: v.Y - v2.Y}
}

func (v *Vector2) Dot(v2 *Vector2) Element {
	return v.X*v2.X + v.Y*v2.Y
}
func (v *Vector2) Cross(v2 *Vector2) Element {
	return v.X*v2.Y - v.Y*v2.X
}

func (v *Vector2) Len() Element {
	return Element(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

func (v *Vector2) LenSqr() Element {
	return v.X*v.X + v.Y*v.Y
}

func (v *Vector2) Normalize() *Vector2 {
	l := v.Len()
	if l > 0 {
		v.X /= l
		v.Y /= l
	} else {
		v.X = 1
	}
	return v
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

// Returns Hamilton product
func (a *Vector4) Mul(b *Vector4) *Vector4 {
	return &Vector4{
		W: a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z, // 1
		X: a.W*b.X + a.X*b.W + a.Y*b.Z - a.Z*b.Y, // i
		Y: a.W*b.Y - a.X*b.Z + a.Y*b.W + a.Z*b.X, // j
		Z: a.W*b.Z + a.X*b.Y - a.Y*b.X + a.Z*b.W, // k
	}
}

func (mat *Matrix4) ApplyTo(v *Vector3) *Vector3 {
	return &Vector3{
		mat[0]*v.X + mat[4]*v.Y + mat[8]*v.Z + mat[12],
		mat[1]*v.X + mat[5]*v.Y + mat[9]*v.Z + mat[13],
		mat[2]*v.X + mat[6]*v.Y + mat[10]*v.Z + mat[14],
	}
}

func NewMatrix4() *Matrix4 {
	return &Matrix4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func NewMatrix4FromSlice(a []Element) *Matrix4 {
	mat := &Matrix4{}
	copy(mat[:], a[:])
	return mat
}

func NewScaleMatrix4(x, y, z Element) *Matrix4 {
	return &Matrix4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

func NewTranslateMatrix4(x, y, z Element) *Matrix4 {
	return &Matrix4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

func NewEulerRotationMatrix4(x, y, z Element, rev int) *Matrix4 {
	m := NewMatrix4()
	cx := Element(math.Cos(float64(x)))
	sx := Element(math.Sin(float64(x)))
	cy := Element(math.Cos(float64(y)))
	sy := Element(math.Sin(float64(y)))
	cz := Element(math.Cos(float64(z)))
	sz := Element(math.Sin(float64(z)))

	if rev == 0 {
		m[0] = cy * cz
		m[4] = -cy * sz
		m[8] = sy

		m[1] = cx*sz + sx*cz*sy
		m[5] = cx*cz - sx*sz*sy
		m[9] = -sx * cy

		m[2] = sx*sz - cx*cz*sy
		m[6] = sx*cz + cx*sz*sy
		m[10] = cx * cy
	} else {
		m[0] = cy * cz
		m[4] = sx*cz*sy - cx*sz
		m[8] = cx*cz*sy + sx*sz

		m[1] = cy * sz
		m[5] = sx*sz*sy + cx*cz
		m[9] = cx*sz*sy - sx*cz

		m[2] = -sy
		m[6] = sx * cy
		m[10] = cx * cy
	}
	return m
}

func (b *Matrix4) Mul(a *Matrix4) *Matrix4 {
	r := &Matrix4{}

	r[0] = a[0]*b[0] + a[1]*b[4] + a[2]*b[8] + a[3]*b[12]
	r[1] = a[0]*b[1] + a[1]*b[5] + a[2]*b[9] + a[3]*b[13]
	r[2] = a[0]*b[2] + a[1]*b[6] + a[2]*b[10] + a[3]*b[14]
	r[3] = a[0]*b[3] + a[1]*b[7] + a[2]*b[11] + a[3]*b[15]

	r[4] = a[4]*b[0] + a[5]*b[4] + a[6]*b[8] + a[7]*b[12]
	r[5] = a[4]*b[1] + a[5]*b[5] + a[6]*b[9] + a[7]*b[13]
	r[6] = a[4]*b[2] + a[5]*b[6] + a[6]*b[10] + a[7]*b[14]
	r[7] = a[4]*b[3] + a[5]*b[7] + a[6]*b[11] + a[7]*b[15]

	r[8] = a[8]*b[0] + a[9]*b[4] + a[10]*b[8] + a[11]*b[12]
	r[9] = a[8]*b[1] + a[9]*b[5] + a[10]*b[9] + a[11]*b[13]
	r[10] = a[8]*b[2] + a[9]*b[6] + a[10]*b[10] + a[11]*b[14]
	r[11] = a[8]*b[3] + a[9]*b[7] + a[10]*b[11] + a[11]*b[15]

	r[12] = a[12]*b[0] + a[13]*b[4] + a[14]*b[8] + a[15]*b[12]
	r[13] = a[12]*b[1] + a[13]*b[5] + a[14]*b[9] + a[15]*b[13]
	r[14] = a[12]*b[2] + a[13]*b[6] + a[14]*b[10] + a[15]*b[14]
	r[15] = a[12]*b[3] + a[13]*b[7] + a[14]*b[11] + a[15]*b[15]
	return r
}

func (mat *Matrix4) ToArray(a []Element) {
	copy(a, mat[:])
}
