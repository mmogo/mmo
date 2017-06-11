package main

import (
	"log"
	"net/http"
)

func main() {
	http.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir("."))))
	log.Fatal(http.ListenAndServe(":8080", http.DefaultServeMux))
}
