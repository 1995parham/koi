package koi

import (
	"errors"
)

var (
	ErrWorkerNotFound     = errors.New("worker not found")
	ErrMinConcurrentCount = errors.New("concurrent count must be at least 1")
)

type Pond[T any, E any] struct {
	Workers map[string]*Worker[T, E]
}

func NewPond[T any, E any]() *Pond[T, E] {
	return &Pond[T, E]{
		Workers: make(map[string]*Worker[T, E]),
	}
}

func (p *Pond[T, E]) RegisterWorker(id string, worker *Worker[T, E]) error {
	if err := worker.Validate(); err != nil {
		return err
	}

	go p.manageWorker(worker)

	p.Workers[id] = worker

	return nil
}

func (p *Pond[T, E]) MustRegisterWorker(id string, worker *Worker[T, E]) {
	if err := p.RegisterWorker(id, worker); err != nil {
		panic(err)
	}
}

func (p *Pond[T, E]) AddWork(workerID string, request T) (<-chan E, error) {
	worker, ok := p.Workers[workerID]
	if !ok {
		return nil, ErrWorkerNotFound
	}

	// add request to worker queue
	worker.RequestChan <- request

	return worker.ResultChan, nil
}

func (p Pond[T, E]) ResultChan(workerID string) <-chan E {
	worker, ok := p.Workers[workerID]

	if ok {
		return worker.ResultChan
	}

	return nil
}

func (p *Pond[T, E]) manageWorker(worker *Worker[T, E]) {
	for i := 0; i < worker.ConcurrentCount; i++ {
		go worker.work()
	}
}
