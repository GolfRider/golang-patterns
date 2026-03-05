package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// DataRecord represents the schema for our Data Platform
type DataRecord struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type Server struct {
	mu    sync.RWMutex
	store map[string]string // Simple in-memory datastore
}

func NewServer() *Server {
	return &Server{
		store: make(map[string]string),
	}
}

/*
// 1. Create the Global Lifecycle Context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Setup your Workers (Consumers/Producers)
	var wg sync.WaitGroup
	// ... start goroutines with wg.Add(1) ...

	// 3. The Orchestrator "Waits"
	<-ctx.Done() // Block here until a signal is received
	fmt.Println("\nShutdown signal received. Cleaning up...")

	// 4. Final Cleanup Phase
	// Here is where you might close channels or call specific Stop() methods
	wg.Wait()
	fmt.Println("Graceful shutdown complete. Goodbye.")
*/

func practice18() {
	s := NewServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})
	mux.HandleFunc("/get", s.handleGet)
	mux.HandleFunc("/post", s.handlePost)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second, // Protect against "Slowloris" attacks
		WriteTimeout: 10 * time.Second,
	}

	// 1. Run server in a goroutine
	go func() {
		fmt.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// 2. Wait for interrupt signal (SIGINT, SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// 3. Graceful Shutdown Phase
	fmt.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server Shutdown Failed:%+v", err)
	}
	fmt.Println("Server exited")
}

// GET Handler: Fetch a resource by ID
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	val, exists := s.store[id]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Value: %s", val)
}

// POST Handler: Create or update a resource
func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DataRecord
	// Staff Tip: Use a limited reader to prevent OOM attacks from massive payloads
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024*1024)) // 1MB limit
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ID == "" || req.Value == "" {
		http.Error(w, "Missing required fields", http.StatusUnprocessableEntity)
		return
	}

	s.mu.Lock()
	s.store[req.ID] = req.Value
	s.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Record %s created", req.ID)
}
