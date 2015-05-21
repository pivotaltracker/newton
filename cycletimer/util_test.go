package main

import (
	"testing"
)

func TestMaxUnderWithDuplicateMax(t *testing.T) {
	xs := []float64{1, 2, 3, 3}
	m := MaxUnder(3.0, xs...)

	if m != 3.0 {
		t.Errorf("Expected max under to be %f but go %f", 3.0, m)
	}
}

func TestMaxUnderWhereUnderDoesNotExist(t *testing.T) {
	xs := []float64{1, 2, 3, 4}
	m := MaxUnder(4.1, xs...)

	if m != 4.0 {
		t.Errorf("Expected max under to be %f but go %f", 4.0, m)
	}
}
