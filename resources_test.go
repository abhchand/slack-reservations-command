package main

import (
	"fmt"
	"os"
	"testing"
)

func TestListOfResources(t *testing.T) {

	// Setup
	old_env := os.Getenv("RESOURCES")
	defer os.Setenv("RESOURCES", old_env)
	os.Setenv("RESOURCES", "PRODUCTION,  sTaging")

	expected := []string{"production", "staging"}
	actual := ListOfResources()

	for i, e := range expected {
		a := actual[i]

		if a != e {
			t.Error(
				fmt.Sprintf("expected (index %v)", i), e,
				"got", a,
			)
		}
	}

}

func TestListOfResourcesToString(t *testing.T) {

	// Setup
	old_env := os.Getenv("RESOURCES")
	defer os.Setenv("RESOURCES", old_env)
	os.Setenv("RESOURCES", "production,  staging")

	expected := "[production, staging]"
	actual := ListOfResourcesToString()

	if actual != expected {
		t.Error(
			"expected", expected,
			"got", actual,
		)
	}

}

func TestIsValidResource(t *testing.T) {

	// Setup
	old_env := os.Getenv("RESOURCES")
	defer os.Setenv("RESOURCES", old_env)
	os.Setenv("RESOURCES", "production, staging")

	test_cases := map[string]bool{
		"production": true,
		"foo":        false,
	}

	for resource, expected := range test_cases {
		actual := IsValidResource(resource)

		if actual != expected {
			t.Error(
				"expected", expected,
				"got", actual,
			)
		}
	}

}
