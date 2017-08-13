package main

import (
	"testing"
)

func TestMaskToken(t *testing.T) {

	token := "gIkuvaNzQIHg97ATvDxqgjtO"

	expected := "********************gjtO"
	actual := maskToken(token)

	if actual != expected {
		t.Error(
			"expected", expected,
			"got", actual,
		)
	}

}
