package geom

import (
	"math"
	"testing"
)

func TestQuaternion(t *testing.T) {
	const eps = 0.000001

	{
		q := NewEuler(0, 0, 0, RotationOrderXYZ).ToQuaternion()
		v1 := NewVector3(1, 2, 3)
		v2 := q.ApplyTo(v1)
		if v2.Sub(v1).Len() > eps {
			t.Error("v1 != v2: ", v1, v2)
		}
	}

	{
		q := NewEuler(2*math.Pi, 0, 0, RotationOrderXYZ).ToQuaternion()
		v1 := NewVector3(1, 2, 3)
		v2 := q.ApplyTo(v1)
		if v2.Sub(v1).Len() > eps {
			t.Error("v1 != v2: ", v1, v2)
		}
	}

	{
		q := NewEuler(math.Pi, 0, 0, RotationOrderXYZ).ToQuaternion()
		q = q.Mul(q)
		v1 := NewVector3(1, 2, 3)
		v2 := q.ApplyTo(v1)
		if v2.Sub(v1).Len() > eps {
			t.Error("v1 != v2: ", v1, v2)
		}
	}

	{
		q := NewEuler(1, 2, 3, RotationOrderXYZ).ToQuaternion()
		q = q.Mul(q.Inverse())
		v1 := NewVector3(1, 2, 3)
		v2 := q.ApplyTo(v1)
		if v2.Sub(v1).Len() > eps {
			t.Error("v1 != v2: ", v1, v2)
		}
	}
}
