package helper

import "testing"

func Test_IsValidNumber(t *testing.T) {
	if !IsValidNumber("421") {
		t.Error()
	}
	if !IsValidNumber("0.8") {
		t.Error()
	}
}
