package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var ID = os.Getenv("SERVER_ID")
var PORT = os.Getenv("PORT")

func main() {
	// This code will be containerized and we will provide as environment variable the name and the port
	if ID == "" {
		ID = "Unkown-Worker"
	}

	if PORT == "" {
		PORT = "8080"
	}

	http.HandleFunc("/", handler)

	log.Printf("[%s] Starting dummy backend on port %s...", ID, PORT)
	addr := ":" + PORT
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	//TODO: Provide some kind of restricted access to allow only the reverse proxy to access it.
	log.Printf("[%s] Received request from %s.", ID, r.RemoteAddr)

	// We can also simulate that the backend has some work by sleeping
	delay := rand.Intn(500)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// Then respond
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello from Backend [%s] (Processed in %v)\n", ID, delay)

}
