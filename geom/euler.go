package geom

import "math"

type RotationOrder int

const (
	RotationOrderXYZ = iota
	RotationOrderZXY
)

type EulerAngles struct {
	Vector3
	Order RotationOrder
}

func NewEuler(x, y, z float32, order RotationOrder) *EulerAngles {
	return &EulerAngles{Vector3: Vector3{x, y, z}, Order: order}
}

func NewEulerFromQuaternion(q *Quaternion, order RotationOrder) *EulerAngles {
	return NewEulerFromMatrix4(NewRotationMatrix4FromQuaternion(q), order)
}

func NewEulerFromMatrix4(mat *Matrix4, order RotationOrder) *EulerAngles {
	const eps = 0.00000001
	m11 := float64(mat[0])
	m21 := float64(mat[1])
	m31 := float64(mat[2])
	m12 := float64(mat[4])
	m22 := float64(mat[5])
	m32 := float64(mat[6])
	m13 := float64(mat[8])
	m23 := float64(mat[9])
	m33 := float64(mat[10])

	ret := &EulerAngles{Order: order}
	switch order {
	case RotationOrderXYZ:
		ret.Y = Element(math.Asin(math.Max(-1, math.Min(m13, 1))))
		if math.Abs(m13) < 1-eps {
			ret.X = Element(math.Atan2(-m23, m33))
			ret.Z = Element(math.Atan2(-m12, m11))
		} else {
			ret.X = Element(math.Atan2(m32, m22))
			ret.Z = 0
		}
		break
	case RotationOrderZXY:
		ret.X = Element(math.Asin(math.Max(-1, math.Min(m32, 1))))
		if math.Abs(m32) < 1-eps {
			ret.Y = Element(math.Atan2(-m31, m33))
			ret.Z = Element(math.Atan2(-m12, m22))
		} else {
			ret.Z = Element(math.Atan2(m21, m11))
			ret.Y = 0
		}
		break
	}
	return ret
}

func (v *EulerAngles) ToQuaternion() *Quaternion {
	cx := math.Cos(float64(v.X / 2))
	cy := math.Cos(float64(v.Y / 2))
	cz := math.Cos(float64(v.Z / 2))
	sx := math.Sin(float64(v.X / 2))
	sy := math.Sin(float64(v.Y / 2))
	sz := math.Sin(float64(v.Z / 2))

	switch v.Order {
	case RotationOrderXYZ:
		return &Vector4{
			X: float32(sx*cy*cz + cx*sy*sz),
			Y: float32(cx*sy*cz - sx*cy*sz),
			Z: float32(cx*cy*sz + sx*sy*cz),
			W: float32(cx*cy*cz - sx*sy*sz)}
	case RotationOrderZXY:
		return &Vector4{
			X: float32(sx*cy*cz - cx*sy*sz),
			Y: float32(cx*sy*cz + sx*cy*sz),
			Z: float32(cx*cy*sz + sx*sy*cz),
			W: float32(cx*cy*cz - sx*sy*sz)}
	default:
		return &Quaternion{0, 0, 0, 1}
	}
}
