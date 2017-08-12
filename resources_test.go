package main

import(
    "fmt"
    "os"
    "testing"
)

func TestListOfResources(t *testing.T) {

    // Setup
    old_env := os.Getenv("RESOURCES")
    defer os.Setenv("RESOURCES", old_env)
    os.Setenv("RESOURCES", "DEVELOPMENT,  sTaging")

    expected := []string{"development", "staging"}
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

func TestIsValidResource(t *testing.T) {

    // Setup
    old_env := os.Getenv("RESOURCES")
    defer os.Setenv("RESOURCES", old_env)
    os.Setenv("RESOURCES", "development, staging")

    test_cases := map[string]bool{
        "development": true,
        "foo": false,
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
