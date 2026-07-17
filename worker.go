package koi

import (
	"sync"
)

// NoReturn is the result type for fire-and-forget workers that produce no
// meaningful output. Use it as the E type parameter together with None.
type NoReturn int

// None is the canonical NoReturn value returned by workers without a result.
const None NoReturn = 0

// Worker runs Work concurrently over requests received on RequestChan and,
// unless its result type is NoReturn, publishes results on ResultChan.
type Worker[T any, E any] struct {
	QueueSize       uint
	ConcurrentCount int
	Work            func(T) E

	ResultChan  chan E
	RequestChan chan T

	// noReturn is true when E is NoReturn, in which case results are dropped
	// instead of being sent on ResultChan. It is computed once at creation so
	// the hot path never touches reflection.
	noReturn bool
	wg       sync.WaitGroup
}

// NewWorker creates and validates a Worker. queueSize sets the buffer of both
// the request and result channels; concurrentCount sets how many goroutines
// process requests in parallel and must be at least 1.
func NewWorker[T any, E any](work func(T) E, queueSize uint, concurrentCount int) (*Worker[T, E], error) {
	var zero E
	_, noReturn := any(zero).(NoReturn)

	w := &Worker[T, E]{
		QueueSize:       queueSize,
		ConcurrentCount: concurrentCount,
		Work:            work,
		ResultChan:      make(chan E, queueSize),
		RequestChan:     make(chan T, queueSize),
		noReturn:        noReturn,
	}

	return w, w.Validate()
}

// MustNewWorker is like NewWorker but panics on a validation error.
func MustNewWorker[T any, E any](work func(T) E, queueSize uint, concurrentCount int) *Worker[T, E] {
	w, err := NewWorker(work, queueSize, concurrentCount)
	if err != nil {
		panic(err)
	}

	return w
}

// start launches ConcurrentCount processing goroutines.
func (w *Worker[T, E]) start() {
	w.wg.Add(w.ConcurrentCount)

	for range w.ConcurrentCount {
		go w.work()
	}
}

// shutdown stops accepting work, waits for in-flight requests to drain, and
// closes ResultChan so consumers ranging over it terminate.
func (w *Worker[T, E]) shutdown() {
	close(w.RequestChan)
	w.wg.Wait()
	close(w.ResultChan)
}

func (w *Worker[T, E]) work() {
	defer w.wg.Done()

	for request := range w.RequestChan {
		res := w.Work(request)
		if !w.noReturn {
			w.ResultChan <- res
		}
	}
}

// Validate reports whether the worker is configured correctly.
func (w *Worker[T, E]) Validate() error {
	if w.ConcurrentCount < 1 {
		return ErrMinConcurrentCount
	}

	return nil
}
