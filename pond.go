package koi

import (
	"errors"
	"sync"
)

var (
	ErrWorkerNotFound     = errors.New("worker not found")
	ErrMinConcurrentCount = errors.New("concurrent count must be at least 1")
	ErrPondClosed         = errors.New("pond is closed")
)

// Pond owns a set of named workers and routes work to them. It is safe for
// concurrent use.
type Pond[T any, E any] struct {
	mu      sync.RWMutex
	workers map[string]*Worker[T, E]
	closed  bool
}

func NewPond[T any, E any]() *Pond[T, E] {
	return &Pond[T, E]{
		mu:      sync.RWMutex{},
		workers: make(map[string]*Worker[T, E]),
		closed:  false,
	}
}

// RegisterWorker validates and starts the worker, making it addressable by id.
func (p *Pond[T, E]) RegisterWorker(id string, worker *Worker[T, E]) error {
	if err := worker.Validate(); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrPondClosed
	}

	worker.start()
	p.workers[id] = worker

	return nil
}

// MustRegisterWorker is like RegisterWorker but panics on error.
func (p *Pond[T, E]) MustRegisterWorker(id string, worker *Worker[T, E]) {
	if err := p.RegisterWorker(id, worker); err != nil {
		panic(err)
	}
}

// AddWork enqueues request for the worker registered under workerID and returns
// that worker's result channel.
func (p *Pond[T, E]) AddWork(workerID string, request T) (<-chan E, error) {
	p.mu.RLock()

	if p.closed {
		p.mu.RUnlock()

		return nil, ErrPondClosed
	}

	worker, ok := p.workers[workerID]
	p.mu.RUnlock()

	if !ok {
		return nil, ErrWorkerNotFound
	}

	// add request to worker queue
	worker.RequestChan <- request

	return worker.ResultChan, nil
}

// ResultChan returns the result channel of the worker registered under
// workerID, or nil if no such worker exists.
func (p *Pond[T, E]) ResultChan(workerID string) <-chan E {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if worker, ok := p.workers[workerID]; ok {
		return worker.ResultChan
	}

	return nil
}

// MapResults returns a channel that yields fn applied to each result produced by
// the worker registered under workerID, or nil if no such worker exists.
//
// The result type U is chosen per call and is independent of the pond's own
// result type E: MapResults is a Go 1.27 generic method, so U lives in the
// method's scope rather than the package's. The returned channel is closed once
// the worker's result channel drains, i.e. after Close.
//
// MapResults consumes from the worker's result channel, so a given worker's
// results should be read either through MapResults or through ResultChan, not
// both.
func (p *Pond[T, E]) MapResults[U any](workerID string, fn func(E) U) <-chan U {
	p.mu.RLock()
	worker, ok := p.workers[workerID]
	p.mu.RUnlock()

	if !ok {
		return nil
	}

	out := make(chan U, worker.QueueSize)

	go func() {
		defer close(out)

		for res := range worker.ResultChan {
			out <- fn(res)
		}
	}()

	return out
}

// Close stops every worker, waits for in-flight work to finish, and closes each
// worker's result channel. After Close returns, AddWork and RegisterWorker fail
// with ErrPondClosed. Close is idempotent.
func (p *Pond[T, E]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	var wg sync.WaitGroup

	for _, worker := range p.workers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			worker.shutdown()
		}()
	}

	wg.Wait()
}
