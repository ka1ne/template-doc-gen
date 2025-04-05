package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	port := flag.String("port", "8000", "Port to serve on")
	dir := flag.String("dir", ".", "Directory to serve files from")
	flag.Parse()

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Serving files from %s on http://localhost:%s", absDir, *port)
	log.Printf("Press Ctrl+C to stop")

	err = http.ListenAndServe(":"+*port, http.FileServer(http.Dir(absDir)))
	if err != nil {
		log.Fatal(err)
	}
}
