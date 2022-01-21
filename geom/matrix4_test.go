package geom

import (
	"math"
	"testing"
)

func TestDecomposeMatrix(t *testing.T) {
	const eps = 0.000001

	pos := NewVector3(1, 2, 3)
	rot := NewEuler(10*math.Pi/180, 20*math.Pi/180, 30*math.Pi/180, RotationOrderZXY).ToQuaternion()
	scale := NewVector3(1.5, 1.6, 1.7)

	mat := NewTRSMatrix4(pos, rot, scale)
	pos1, rot1, scale1 := mat.Decompose()

	if pos.Sub(pos1).Len() > eps {
		t.Error("pos: ", pos, pos1)
	}
	if rot.Sub(rot1).Len() > eps {
		t.Error("rot: ", rot, rot1)
	}
	if scale.Sub(scale1).Len() > eps {
		t.Error("scale: ", scale, scale1)
	}

	mat2 := NewRotationMatrix4FromQuaternion(rot)
	pos1, rot1, scale1 = mat2.Decompose()
	if rot.Sub(rot1).Len() > eps {
		t.Error("rot: ", rot, rot1)
	}
	if pos1.Len() > eps {
		t.Error("pos: ", pos1)
	}
	if scale1.Sub(NewVector3(1, 1, 1)).Len() > eps {
		t.Error("scale: ", scale1)
	}
}
