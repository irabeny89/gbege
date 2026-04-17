package main

import (
	"log"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("unable to listen: %s", err)
	}
	log.Println("Started server on :8080")
}
