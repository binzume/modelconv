package geom

import "math"

func NewQuaternionFromEulerZXY(v *Vector3) *Vector4 {
	c1 := math.Cos(float64(v.X / 2))
	c2 := math.Cos(float64(v.Y / 2))
	c3 := math.Cos(float64(v.Z / 2))
	s1 := math.Sin(float64(v.X / 2))
	s2 := math.Sin(float64(v.Y / 2))
	s3 := math.Sin(float64(v.Z / 2))

	return &Vector4{
		X: float32(s1*c2*c3 - c1*s2*s3),
		Y: float32(c1*s2*c3 + s1*c2*s3),
		Z: float32(c1*c2*s3 + s1*s2*c3),
		W: float32(c1*c2*c3 - s1*s2*s3)}
}

func IsInTriangle(p, a, b, c *Vector3) bool {
	ab, bc, ca := b.Sub(a), c.Sub(b), a.Sub(c)
	c1, c2, c3 := ab.Cross(p.Sub(a)), bc.Cross(p.Sub(b)), ca.Cross(p.Sub(c))
	return c1.Dot(c2) > 0 && c2.Dot(c3) > 0 && c3.Dot(c1) > 0
}

func Triangulate(poly []*Vector3) [][3]int {
	var dst [][3]int
	if len(poly) < 3 {
		return dst
	}
	n := &Vector3{}
	ii := make([]int, len(poly))
	for i := range poly {
		ii[i] = i
		v0 := poly[(i+len(poly)-1)%len(poly)]
		v1 := poly[i]
		v2 := poly[(i+1)%len(poly)]
		n = n.Add(v0.Sub(v1).Cross(v2.Sub(v1)))
	}
	n = n.Normalize()

	// O(N*N)...
	count := len(ii)
	for count >= 3 {
		last_count := count
		for i := count - 1; i >= 0; i-- {
			i0 := ii[(i+count-1)%count]
			i1 := ii[i]
			i2 := ii[(i+1)%count]
			v0 := poly[i0]
			v1 := poly[i1]
			v2 := poly[i2]
			if v0.Sub(v1).Cross(v2.Sub(v1)).Dot(n) >= 0 {
				ok := true
				var tmp []int
				tmp = append(tmp, ii[:i]...)
				tmp = append(tmp, ii[i+1:]...)
				for _, i := range tmp {
					if IsInTriangle(poly[i], v0, v1, v2) {
						ok = false
						break
					}
				}
				if ok {
					dst = append(dst, [3]int{i0, i1, i2})
					ii = tmp
					count--
				}
			}
		}
		if last_count == count {
			// error: maybe self-intersecting polygon
			for i := 0; i < len(ii)-2; i++ {
				dst = append(dst, [3]int{ii[0], ii[i+1], ii[i+2]})
			}
			break
		}
	}
	return dst
}
