package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	// parse command-line flags
	port := flag.String("port", "8000", "port to serve on")
	dir := flag.String("dir", ".", "directory to serve files from")
	flag.Parse()

	// get absolute path of directory
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("failed to get absolute path: %v", err)
	}

	// create file server handler
	fileServer := http.FileServer(http.Dir(absDir))
	http.Handle("/", fileServer)

	// create server with timeout
	server := &http.Server{
		Addr:         ":" + *port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// start server in a goroutine
	go func() {
		log.Printf("serving files from %s on http://localhost:%s", absDir, *port)
		log.Printf("press Ctrl+C to stop")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	// setup signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// shutdown gracefully
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
