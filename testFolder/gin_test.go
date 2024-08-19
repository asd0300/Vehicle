package testfolder

import (
	"testing"
)

func TestExample(t *testing.T) {
	result := IsGood()
	expected := true
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func IsGood() bool {
	return false
}
