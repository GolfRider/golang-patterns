#  Quick Checklist

*Use this when encountering an unrecognized concurrency or architectural problem.*

---

## 1. Start with the Struct

Before writing logic, define your primitives. Ask yourself:

- **State:** Do I need a `sync.Mutex` or `sync.RWMutex` to protect data?
- **Shutdown:** Do I need a `sync.WaitGroup` to track active goroutines?
- **Control:** Do I need a `context.Context` for propagation and cancellation?

## 2. Draw the Pipeline

Visualize the flow of data through the system:

- **The Producer:** Who is generating the data?
- **The Consumer:** Who is processing or receiving the data?
- **The Buffer:** What is the capacity? (e.g., Unbuffered vs. Buffered channels)

## 3. The "Safety First" Loop

Never let a goroutine block indefinitely. Always use a `select` statement to prioritize cancellation:

```go
select {
case data := <-dataChan:
    // Process your data
case <-ctx.Done():
    // Exit early and clean up
    return
}
```

> ⚠️ **Use code with caution.**

## 4. The Graceful Exit

Ensure your program cleans up after itself to avoid leaks:

- **Close:** Signal to consumers that no more data is coming.
- **Wait:** Ensure all background workers have finished before the main process exits.

**Formula:** Always `Close` → `Wait`.


Main Thread: The "Manager" who waits at the exit for everyone to finish. (Orchestrator)
Goroutines: The "Workers" who do the heavy lifting in the background. (producers/consumers)
Channels:  The "Coordination" that prevents workers from colliding or drowning. (Data movement, signal movement,close/broadcast, backpressure/flow-control) 