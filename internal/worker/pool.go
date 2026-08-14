package worker

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Job struct {
	NotificationID uuid.UUID
}

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

func (p *Pool) Enqueue(job Job) {
	p.jobs <- job
}

func (p *Pool) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
}
