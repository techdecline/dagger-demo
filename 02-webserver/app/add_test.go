package main

import (
	"testing"
)

// TestAdd is a unit test for the add function
func TestAdd(t *testing.T) {
	// define some test cases with inputs and expected outputs
	testCases := []struct {
		a, b, want int
	}{
		{1, 2, 3},
		{0, 0, 0},
		{-1, 1, 0},
		{10, -5, 5},
	}
	// loop over the test cases
	for _, tc := range testCases {
		// call the add function with the inputs
		got := add(tc.a, tc.b)
		// check if the output matches the expected output
		if got != tc.want {
			// report an error if they don't match
			t.Errorf("add(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
