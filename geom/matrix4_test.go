package geom

import (
	"math"
	"testing"
)

func TestMatrix4(t *testing.T) {
	var arr [32]Element
	mat := NewMatrix4()
	mat.ToArray(arr[:])

	if *mat != *NewMatrix4FromSlice(arr[:]) {
		t.Error("error: ToArray/FromSlice", mat)
	}

	if *mat != *mat.Transposed().Transposed() {
		t.Error("error: Transposed()", mat)
	}

	if *mat.Sub(mat).Add(mat) != *mat {
		t.Error("error: Sub().Add()", mat)
	}

	if *mat.Mul(mat) != *mat {
		t.Error("error: Mul()", mat)
	}

	if *mat.Scale(1) != *mat {
		t.Error("error: Scale()", mat)
	}
}

func TestMatrix4_Decompose(t *testing.T) {
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

func TestMatrix4_Inverse(t *testing.T) {
	if NewMatrix4().Inverse().Det() != 1 {
		t.Error("inv: ", NewMatrix4().Inverse())
	}
}
