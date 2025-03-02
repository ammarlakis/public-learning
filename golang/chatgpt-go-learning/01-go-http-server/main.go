package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to %s!", r.Host)
}

func main() {
	myserver := http.NewServeMux()

	myserver.HandleFunc("/", handler)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", myserver); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
