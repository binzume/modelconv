package geom

import "math"

type Vector2 struct {
	X Element
	Y Element
}

func NewVector2(x, y float32) *Vector2 {
	return &Vector2{X: x, Y: y}
}

func (v *Vector2) Add(v2 *Vector2) *Vector2 {
	return &Vector2{X: v.X + v2.X, Y: v.Y + v2.Y}
}

func (v *Vector2) Sub(v2 *Vector2) *Vector2 {
	return &Vector2{X: v.X - v2.X, Y: v.Y - v2.Y}
}

func (v *Vector2) Scale(s Element) *Vector2 {
	return &Vector2{X: v.X * s, Y: v.Y * s}
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
