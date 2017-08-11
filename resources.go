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
