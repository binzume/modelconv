package geom

import (
	"math"
	"testing"
)

func TestEuler(t *testing.T) {
	const eps = 0.000001

	for i, c := range []struct {
		order   RotationOrder
		x, y, z float32
	}{
		{RotationOrderXYZ, 10, 20, 30},
		{RotationOrderXYZ, 10, 90, 0},
		{RotationOrderYXZ, 10, 20, 30},
		{RotationOrderYXZ, 90, 10, 0},
		{RotationOrderZXY, 10, 20, 30},
		{RotationOrderZXY, 90, 0, 10},
		{RotationOrderZYX, 10, 20, 30},
		{RotationOrderZYX, 0, 90, 10},
	} {
		e1 := NewEuler(c.x*math.Pi/180, c.y*math.Pi/180, c.z*math.Pi/180, c.order)
		q := e1.ToQuaternion()
		e2 := NewEulerFromQuaternion(q, c.order)

		if e1.Vector3.Sub(&e2.Vector3).Len() > eps {
			t.Error("euler: ", i, e1, e2)
		}
		if Abs(q.Len()-1) > eps {
			t.Error("Quaternion.Len() != 1", e1)
		}
	}
}
