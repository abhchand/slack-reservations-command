package main

import (
	"os"
	"strings"
)

func ListOfResources() []string {

	resources := strings.Split(os.Getenv("RESOURCES"), ",")
	for i, r := range resources {
		resources[i] = strings.ToLower(strings.Trim(r, " "))
	}

	return resources

}

func ListOfResourcesToString() string {

	return "[" + strings.Join(ListOfResources(), ", ") + "]"
}

func IsValidResource(resource string) bool {

	resources := ListOfResources()
	present := false

	for _, r := range resources {
		if resource == r {
			present = true
			break
		}
	}

	return present

}
