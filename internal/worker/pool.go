// Package worker implements Herald's dispatch engine: a fixed-size pool of
// goroutines pulling jobs off a shared channel. This is the direct Go
// replacement for what you'd reach for Laravel Horizon + Redis queues to do.
package worker

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
)

// Job represents one notification that needs to be dispatched.
type Job struct {
	NotificationID uuid.UUID
}

// Pool is a fixed set of worker goroutines consuming from a shared, buffered
// channel. Buffered channels act as an in-memory queue — no Redis required
// at moderate scale.
type Pool struct {
	jobs       chan Job
	numWorkers int
	dispatcher *Dispatcher
	wg         sync.WaitGroup
}

func NewPool(numWorkers int, bufferSize int, dispatcher *Dispatcher) *Pool {
	return &Pool{
		jobs:       make(chan Job, bufferSize),
		numWorkers: numWorkers,
		dispatcher: dispatcher,
	}
}

// Start launches numWorkers goroutines, each running an infinite loop that
// pulls jobs off the channel until it's closed or ctx is cancelled.
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.numWorkers; i++ {
		workerID := i
		p.wg.Add(1)

		go func() {
			defer p.wg.Done()
			log.Printf("[worker %d] started", workerID)

			for {
				select {
				case job, ok := <-p.jobs:
					if !ok {
						log.Printf("[worker %d] job channel closed, exiting", workerID)
						return
					}
					p.dispatcher.Dispatch(ctx, job)

				case <-ctx.Done():
					log.Printf("[worker %d] context cancelled, exiting", workerID)
					return
				}
			}
		}()
	}
}

// Enqueue is how the API handler hands off a notification to be sent
// asynchronously — it returns immediately, the actual send happens in the
// background on a worker goroutine.
func (p *Pool) Enqueue(job Job) {
	p.jobs <- job
}

// Shutdown closes the job channel and waits for in-flight jobs to finish.
// Call this from main.go on graceful shutdown (SIGTERM).
func (p *Pool) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
}
