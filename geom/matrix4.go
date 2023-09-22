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

func NewTRSMatrix4(t *Vector3, r *Quaternion, s *Vector3) *Matrix4 {
	return NewTranslateMatrix4(t.X, t.Y, t.Z).Mul(NewRotationMatrix4FromQuaternion(r).Mul(NewScaleMatrix4(s.X, s.Y, s.Z)))
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
		1 - 2*y*y - 2*z*z, 2*x*y + 2*z*w, 2*x*z - 2*y*w, 0,
		2*x*y - 2*z*w, 1 - 2*x*x - 2*z*z, 2*y*z + 2*x*w, 0,
		2*x*z + 2*y*w, 2*y*z - 2*x*w, 1 - 2*x*x - 2*y*y, 0,
		0, 0, 0, 1,
	}
}

func (m *Matrix4) Add(a *Matrix4) *Matrix4 {
	return &Matrix4{
		m[0] + a[0], m[1] + a[1], m[2] + a[2], m[3] + a[3],
		m[4] + a[4], m[5] + a[5], m[6] + a[6], m[7] + a[7],
		m[8] + a[8], m[9] + a[9], m[10] + a[10], m[11] + a[11],
		m[12] + a[12], m[13] + a[13], m[14] + a[14], m[15] + a[15],
	}
}

func (m *Matrix4) Sub(a *Matrix4) *Matrix4 {
	return &Matrix4{
		m[0] - a[0], m[1] - a[1], m[2] - a[2], m[3] - a[3],
		m[4] - a[4], m[5] - a[5], m[6] - a[6], m[7] - a[7],
		m[8] - a[8], m[9] - a[9], m[10] - a[10], m[11] - a[11],
		m[12] - a[12], m[13] - a[13], m[14] - a[14], m[15] - a[15],
	}
}

func (m *Matrix4) Scale(s Element) *Matrix4 {
	return &Matrix4{
		m[0] * s, m[1] * s, m[2] * s, m[3] * s,
		m[4] * s, m[5] * s, m[6] * s, m[7] * s,
		m[8] * s, m[9] * s, m[10] * s, m[11] * s,
		m[12] * s, m[13] * s, m[14] * s, m[15] * s,
	}
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

func (m *Matrix4) Det() Element {
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

func (mat *Matrix4) Decompose() (translate *Vector3, rotation *Quaternion, scale *Vector3) {
	scale = &Vector3{
		(&Vector3{mat[0], mat[1], mat[2]}).Len(),
		(&Vector3{mat[4], mat[5], mat[6]}).Len(),
		(&Vector3{mat[8], mat[9], mat[10]}).Len(),
	}
	if mat.Det() < 0 {
		scale.X = -scale.X
	}

	m11, m12, m13 := mat[0]/scale.X, mat[4]/scale.Y, mat[8]/scale.Z
	m21, m22, m23 := mat[1]/scale.X, mat[5]/scale.Y, mat[9]/scale.Z
	m31, m32, m33 := mat[2]/scale.X, mat[6]/scale.Y, mat[10]/scale.Z
	trace := m11 + m22 + m33
	if trace > 0 {
		s := 0.5 / math.Sqrt(float64(trace+1.0))
		rotation = &Quaternion{
			(m32 - m23) * Element(s),
			(m13 - m31) * Element(s),
			(m21 - m12) * Element(s),
			0.25 / Element(s),
		}

	} else if m11 > m22 && m11 > m33 {
		s := 2.0 * math.Sqrt(float64(1.0+m11-m22-m33))
		rotation = &Quaternion{
			0.25 * Element(s),
			(m12 + m21) / Element(s),
			(m13 + m31) / Element(s),
			(m32 - m23) / Element(s),
		}
	} else if m22 > m33 {
		s := 2.0 * math.Sqrt(float64(1.0+m22-m11-m33))
		rotation = &Quaternion{
			(m12 + m21) / Element(s),
			0.25 * Element(s),
			(m23 + m32) / Element(s),
			(m13 - m31) / Element(s),
		}
	} else {
		s := 2.0 * math.Sqrt(float64(1.0+m33-m11-m22))
		rotation = &Quaternion{
			(m13 + m31) / Element(s),
			(m23 + m32) / Element(s),
			0.25 * Element(s),
			(m21 - m12) / Element(s),
		}
	}

	translate = &Vector3{mat[12], mat[13], mat[14]}
	return
}

func (m *Matrix4) TranslationScale(s Element) *Matrix4 {
	// scale * m * scale-1
	return &Matrix4{
		m[0], m[1], m[2], m[3],
		m[4], m[5], m[6], m[7],
		m[8], m[9], m[10], m[11],
		m[12] * s, m[13] * s, m[14] * s, m[15],
	}
}

func (m *Matrix4) Clone() *Matrix4 {
	r := *m
	return &r
}

func (mat *Matrix4) ToArray(a []Element) {
	copy(a, mat[:])
}
