package main

import (
	"net/http"
)

var log = initializeLogger()

func main() {

	validateOptions()
	logOptions()

	router := NewRouter()

	log.Info("I'm listening...")
	log.Fatal(http.ListenAndServe(":8080", router))

}
