package geom

import (
	"testing"
)

func TestVector2(t *testing.T) {
	zero := NewVector2(0, 0)
	if zero.Len() != 0 || zero.LenSqr() != 0 || zero.Dot(zero) != 0 {
		t.Error("len != 0")
	}

	if *zero.Normalize() != *NewVector2(1, 0) {
		t.Error("Normalize shoud returns unit vector.", zero.Normalize())
	}

	if *NewVector2(1, 0).Add(NewVector2(0, 1)) != *NewVector2(1, 1) {
		t.Error("Vector.Add()")
	}

	if *NewVector2(1, 2).Scale(2) != *NewVector2(2, 4) {
		t.Error("Vector.Scale()")
	}
}

func TestVector3(t *testing.T) {
	zero := NewVector3(0, 0, 0)
	if zero.Len() != 0 || zero.LenSqr() != 0 || zero.Dot(zero) != 0 {
		t.Error("len != 0")
	}

	if *zero.Normalize() != *NewVector3(1, 0, 0) {
		t.Error("Normalize shoud returns unit vector.", zero.Normalize())
	}

	if *NewVector3(1, 0, 0).Add(NewVector3(0, 1, 0)) != *NewVector3(1, 1, 0) {
		t.Error("Vector.Add()")
	}

	if *NewVector3(1, 2, 3).Scale(2) != *NewVector3(2, 4, 6) {
		t.Error("Vector.Scale()")
	}
}

func TestVector4(t *testing.T) {
	zero := NewVector4(0, 0, 0, 0)
	if zero.Len() != 0 || zero.LenSqr() != 0 || zero.Dot(zero) != 0 {
		t.Error("len != 0")
	}

	if *zero.Normalize() != *NewVector4(0, 0, 0, 1) {
		t.Error("Normalize shoud returns unit vector.", zero.Normalize())
	}

	if *NewVector4(1, 0, 0, 0).Add(NewVector4(0, 1, 0, 0)) != *NewVector4(1, 1, 0, 0) {
		t.Error("Vector.Add()")
	}

	if *NewVector4(1, 2, 3, 4).Scale(2) != *NewVector4(2, 4, 6, 8) {
		t.Error("Vector.Scale()")
	}

	if *NewVector4(1, 2, 3, 4).HadamardProduct(NewVector4(4, 3, 2, 1)) != *NewVector4(4, 6, 6, 4) {
		t.Error("Vector.HadamardProduct()")
	}
}
