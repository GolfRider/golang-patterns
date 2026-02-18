package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func checkGracefulShutdown() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	server := http.Server{Addr: ":5050", Handler: http.DefaultServeMux}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on port 8080: %v", err)
		}
	}()
	log.Println("Listening on port 8080")
	<-ctx.Done()
	log.Println("Shutting down server...")

	nc, cc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cc()
	if err := server.Shutdown(nc); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v", err)
	}

	log.Println("Server gracefully stopped")
}
