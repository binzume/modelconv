package geom

import "math"

type Vector4 struct {
	X Element
	Y Element
	Z Element
	W Element
}

func NewVector4(x, y, z, w float32) *Vector4 {
	return &Vector4{X: x, Y: y, Z: z, W: w}
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

// Returns Hamilton product
func (a *Vector4) Mul(b *Vector4) *Vector4 {
	return &Vector4{
		W: a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z, // 1
		X: a.W*b.X + a.X*b.W + a.Y*b.Z - a.Z*b.Y, // i
		Y: a.W*b.Y - a.X*b.Z + a.Y*b.W + a.Z*b.X, // j
		Z: a.W*b.Z + a.X*b.Y - a.Y*b.X + a.Z*b.W, // k
	}
}

type Quaternion = Vector4

func NewQuaternion(x, y, z, w float32) *Quaternion {
	return &Quaternion{X: x, Y: y, Z: z, W: w}
}

func NewQuaternionFromArray(arr [4]Element) *Quaternion {
	return &Quaternion{X: arr[0], Y: arr[1], Z: arr[2], W: arr[3]}
}

func (v *Quaternion) Inverse() *Quaternion {
	return &Quaternion{X: -v.X, Y: -v.Y, Z: -v.Z, W: v.W}
}

func (q *Quaternion) ApplyTo(v *Vector3) *Vector3 {
	vx, vy, vz := v.X, v.Y, v.Z
	qx, qy, qz, qw := q.X, q.Y, q.Z, q.W
	ix := qw*vx + qy*vz - qz*vy
	iy := qw*vy + qz*vx - qx*vz
	iz := qw*vz + qx*vy - qy*vx
	iw := -qx*vx - qy*vy - qz*vz
	return NewVector3(
		ix*qw+iw*-qx+iy*-qz-iz*-qy,
		iy*qw+iw*-qy+iz*-qx-ix*-qz,
		iz*qw+iw*-qz+ix*-qy-iy*-qx,
	)
}
