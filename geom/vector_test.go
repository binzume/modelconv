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
}
