package main

import (
	"testing"
)

func TestFindValEffect(test *testing.T) {
	string_tol := ":test"
	
	expectedResults := []string{ "", "test"}
	
	test.Log("testing findValEffect")
	value, effect, err := findValEffect(string_tol)
	if err != nil {
		test.Errorf("%v", err)
	} else if value != expectedResults[0] {
		test.Errorf("Expected %+v got %+v", expectedResults[0], value)	
	} else if effect != expectedResults[1] {
		test.Errorf("Expected %+v got %+v", expectedResults[1], effect)
	}
}