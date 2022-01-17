package geom

import "math"

// column-major matrix
type Matrix4 [16]Element

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

func NewRotationMatrix4FromQuaternion(q *Quaternion) *Matrix4 {
	var (
		x = q.X
		y = q.Y
		z = q.Z
		w = q.W
	)
	return &Matrix4{
		1 - 2*y*y - 2*z*z, 2*x*y - 2*z*w, 2*x*z + 2*y*w, 0,
		2*x*y + 2*z*w, 1 - 2*x*x - 2*z*z, 2*y*z - 2*x*w, 0,
		2*x*z - 2*y*w, 2*y*z + 2*x*w, 1 - 2*x*x - 2*y*y, 0,
		0, 0, 0, 1,
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

func (m *Matrix4) Det() float32 {
	var (
		t11 = m[9]*m[14]*m[7] - m[13]*m[10]*m[7] + m[13]*m[6]*m[11] - m[5]*m[14]*m[11] - m[9]*m[6]*m[15] + m[5]*m[10]*m[15]
		t12 = m[12]*m[10]*m[7] - m[8]*m[14]*m[7] - m[12]*m[6]*m[11] + m[4]*m[14]*m[11] + m[8]*m[6]*m[15] - m[4]*m[10]*m[15]
		t13 = m[8]*m[13]*m[7] - m[12]*m[9]*m[7] + m[12]*m[5]*m[11] - m[4]*m[13]*m[11] - m[8]*m[5]*m[15] + m[4]*m[9]*m[15]
		t14 = m[12]*m[9]*m[6] - m[8]*m[13]*m[6] - m[12]*m[5]*m[10] + m[4]*m[13]*m[10] + m[8]*m[5]*m[14] - m[4]*m[9]*m[14]
		det = m[0]*t11 + m[1]*t12 + m[2]*t13 + m[3]*t14
	)
	return det
}

func (m *Matrix4) Inverse() *Matrix4 {
	var (
		t11 = m[9]*m[14]*m[7] - m[13]*m[10]*m[7] + m[13]*m[6]*m[11] - m[5]*m[14]*m[11] - m[9]*m[6]*m[15] + m[5]*m[10]*m[15]
		t12 = m[12]*m[10]*m[7] - m[8]*m[14]*m[7] - m[12]*m[6]*m[11] + m[4]*m[14]*m[11] + m[8]*m[6]*m[15] - m[4]*m[10]*m[15]
		t13 = m[8]*m[13]*m[7] - m[12]*m[9]*m[7] + m[12]*m[5]*m[11] - m[4]*m[13]*m[11] - m[8]*m[5]*m[15] + m[4]*m[9]*m[15]
		t14 = m[12]*m[9]*m[6] - m[8]*m[13]*m[6] - m[12]*m[5]*m[10] + m[4]*m[13]*m[10] + m[8]*m[5]*m[14] - m[4]*m[9]*m[14]
		det = m[0]*t11 + m[1]*t12 + m[2]*t13 + m[3]*t14
	)

	r := &Matrix4{}
	if det == 0 {
		return r
	}

	r[0] = t11 / det
	r[1] = (m[13]*m[10]*m[3] - m[9]*m[14]*m[3] - m[13]*m[2]*m[11] + m[1]*m[14]*m[11] + m[9]*m[2]*m[15] - m[1]*m[10]*m[15]) / det
	r[2] = (m[5]*m[14]*m[3] - m[13]*m[6]*m[3] + m[13]*m[2]*m[7] - m[1]*m[14]*m[7] - m[5]*m[2]*m[15] + m[1]*m[6]*m[15]) / det
	r[3] = (m[9]*m[6]*m[3] - m[5]*m[10]*m[3] - m[9]*m[2]*m[7] + m[1]*m[10]*m[7] + m[5]*m[2]*m[11] - m[1]*m[6]*m[11]) / det
	r[4] = t12 / det
	r[5] = (m[8]*m[14]*m[3] - m[12]*m[10]*m[3] + m[12]*m[2]*m[11] - m[0]*m[14]*m[11] - m[8]*m[2]*m[15] + m[0]*m[10]*m[15]) / det
	r[6] = (m[12]*m[6]*m[3] - m[4]*m[14]*m[3] - m[12]*m[2]*m[7] + m[0]*m[14]*m[7] + m[4]*m[2]*m[15] - m[0]*m[6]*m[15]) / det
	r[7] = (m[4]*m[10]*m[3] - m[8]*m[6]*m[3] + m[8]*m[2]*m[7] - m[0]*m[10]*m[7] - m[4]*m[2]*m[11] + m[0]*m[6]*m[11]) / det
	r[8] = t13 / det
	r[9] = (m[12]*m[9]*m[3] - m[8]*m[13]*m[3] - m[12]*m[1]*m[11] + m[0]*m[13]*m[11] + m[8]*m[1]*m[15] - m[0]*m[9]*m[15]) / det
	r[10] = (m[4]*m[13]*m[3] - m[12]*m[5]*m[3] + m[12]*m[1]*m[7] - m[0]*m[13]*m[7] - m[4]*m[1]*m[15] + m[0]*m[5]*m[15]) / det
	r[11] = (m[8]*m[5]*m[3] - m[4]*m[9]*m[3] - m[8]*m[1]*m[7] + m[0]*m[9]*m[7] + m[4]*m[1]*m[11] - m[0]*m[5]*m[11]) / det
	r[12] = t14 / det
	r[13] = (m[8]*m[13]*m[2] - m[12]*m[9]*m[2] + m[12]*m[1]*m[10] - m[0]*m[13]*m[10] - m[8]*m[1]*m[14] + m[0]*m[9]*m[14]) / det
	r[14] = (m[12]*m[5]*m[2] - m[4]*m[13]*m[2] - m[12]*m[1]*m[6] + m[0]*m[13]*m[6] + m[4]*m[1]*m[14] - m[0]*m[5]*m[14]) / det
	r[15] = (m[4]*m[9]*m[2] - m[8]*m[5]*m[2] + m[8]*m[1]*m[6] - m[0]*m[9]*m[6] - m[4]*m[1]*m[10] + m[0]*m[5]*m[10]) / det

	return r
}

func (m *Matrix4) Transposed() *Matrix4 {
	return &Matrix4{
		m[0], m[4], m[8], m[12],
		m[1], m[5], m[9], m[13],
		m[2], m[6], m[10], m[14],
		m[3], m[7], m[11], m[15],
	}
}

func (m *Matrix4) Clone() *Matrix4 {
	r := *m
	return &r
}

func (mat *Matrix4) ToArray(a []Element) {
	copy(a, mat[:])
}

func (mat *Matrix4) ToEulerZXY() *Vector3 {
	m11 := float64(mat[0])
	m21 := float64(mat[1])
	m31 := float64(mat[2])
	m12 := float64(mat[4])
	m22 := float64(mat[5])
	m32 := float64(mat[6])
	//m13 := float64(mat[8])
	//m23 := float64(mat[9])
	m33 := float64(mat[10])

	ret := &Vector3{}
	ret.X = float32(math.Asin(math.Max(-1, math.Min(m32, 1))))
	if math.Abs(m32) < 0.9999999 {
		ret.Y = float32(math.Atan2(-m31, m33))
		ret.Z = float32(math.Atan2(-m12, m22))
	} else {
		ret.Z = float32(math.Atan2(m21, m11))
		ret.Y = 0
	}
	return ret
}
