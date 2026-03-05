package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Event1 struct {
	ID    int
	Raw   string
	Value int
	Valid bool
}

type Pipeline struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewPipeline() *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pipeline{ctx: ctx, cancel: cancel}
}

// Stage 1: Ingest - Simulates reading from a stream like NATS or Kafka
func (p *Pipeline) Ingest(out chan<- Event1) {
	defer p.wg.Done()
	defer close(out) // Signal Stage 2 that no more data is coming

	for i := 1; i <= 10; i++ {
		select {
		case <-p.ctx.Done():
			return
		case out <- Event1{ID: i, Raw: fmt.Sprintf("data-%d", i)}:
			fmt.Printf("[Ingest] Sent Event %d\n", i)
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// Stage 2: Transform - Heavy lifting (parsing, validation, enrichment)
func (p *Pipeline) Transform(in <-chan Event1, out chan<- Event1) {
	defer p.wg.Done()
	defer close(out) // Signal Stage 3 to finish up

	for event := range in {
		// Simulate processing logic
		event.Value = event.ID * 10
		event.Valid = true

		select {
		case <-p.ctx.Done():
			return
		case out <- event:
			fmt.Printf("[Transform] Processed Event %d\n", event.ID)
		}
	}
}

// Stage 3: Sink - Writing to a Database or Search Index
func (p *Pipeline) Sink(in <-chan Event1) {
	defer p.wg.Done()
	for event := range in {
		// Simulate a database write
		fmt.Printf("[Sink] Stored Event %d with Value %d\n", event.ID, event.Value)
		time.Sleep(50 * time.Millisecond)
	}
}

func (p *Pipeline) Run() {
	// Channels act as "Conveyor Belts" between stages
	// Using buffer size for Backpressure: Stage 1 will block if Stage 2 is slow
	ingestChan := make(chan Event1, 5)
	transformChan := make(chan Event1, 5)

	p.wg.Add(3)
	go p.Ingest(ingestChan)
	go p.Transform(ingestChan, transformChan)
	go p.Sink(transformChan)

	// Wait for all stages to finish their "range" loops
	p.wg.Wait()
	fmt.Println("Pipeline completed successfully.")
}

func problem13() {
	p := NewPipeline()

	// Handle graceful shutdown (e.g., from SIGTERM)
	go func() {
		time.Sleep(2 * time.Second)
		// p.cancel() // Uncomment to test mid-stream shutdown
	}()

	p.Run()
}
