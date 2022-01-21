package geom

import (
	"math"
	"testing"
)

func TestEulerXYZ(t *testing.T) {
	const eps = 0.000001

	e1 := NewEuler(10*math.Pi/180, 20*math.Pi/180, 30*math.Pi/180, RotationOrderXYZ)
	e2 := NewEulerFromQuaternion(e1.ToQuaternion(), RotationOrderXYZ)

	if e1.Vector3.Sub(&e2.Vector3).Len() > eps {
		t.Error("euler: ", e1, e2)
	}
}

func TestEulerZXY(t *testing.T) {
	const eps = 0.000001

	e1 := NewEuler(10*math.Pi/180, 20*math.Pi/180, 30*math.Pi/180, RotationOrderZXY)
	e2 := NewEulerFromQuaternion(e1.ToQuaternion(), RotationOrderZXY)

	if e1.Vector3.Sub(&e2.Vector3).Len() > eps {
		t.Error("euler: ", e1, e2)
	}
}
