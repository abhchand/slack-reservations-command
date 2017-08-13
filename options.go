package main

import (
    "fmt"
    "os"
    "regexp"
)

func validateOptions() {

    resources := os.Getenv("RESOURCES")
    if resources == "" {
        fmt.Println("Please set environment variable RESOURCES")
        os.Exit(1)
    }

    token := os.Getenv("SLACK_VERIFICATION_TOKEN")
    if !regexp.MustCompile("\\A[a-zA-Z0-9]{24}\\z").Match([]byte(token)) {
        fmt.Println(
            "Environment variable SLACK_VERIFICATION_TOKEN missing or invalid")
        os.Exit(1)
    }

}

func logOptions() {

    // Mask all but last 4 digits of token
    token := os.Getenv("SLACK_VERIFICATION_TOKEN")
    token_bytes := []byte(token)
    for i := 0; i < len(token)-4; i++ { token_bytes[i] = '*' }
    token = string(token_bytes)

    log.Infof("Available resources: %v", ListOfResourcesToString())
    log.Infof("Slack API Token: %v", string(token_bytes))

}
