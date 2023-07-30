package koi

import (
	"context"
	"log"
	"reflect"

	"golang.org/x/sync/semaphore"
)

type NoReturn int

const None NoReturn = 0

type Worker[T any, E any] struct {
	QueueSize       uint
	ConcurrentCount int64
	Work            func(T) E

	ResultChan  chan E
	RequestChan chan T
	Semaphore   *semaphore.Weighted
}

func NewWoker[T any, E any](work func(T) E, queueSize uint, concurrentCount int64) (Worker[T, E], error) {
	w := Worker[T, E]{
		QueueSize:       queueSize,
		ConcurrentCount: concurrentCount,
		Work:            work,
		ResultChan:      make(chan E, queueSize),
		RequestChan:     make(chan T, queueSize),
		Semaphore:       semaphore.NewWeighted(concurrentCount),
	}

	return w, w.Validate()
}

func MustNewWoker[T any, E any](work func(T) E, queueSize uint, concurrentCount int64) Worker[T, E] {
	w := Worker[T, E]{
		QueueSize:       queueSize,
		ConcurrentCount: concurrentCount,
		Work:            work,
	}

	if err := w.Validate(); err != nil {
		panic(err)
	}

	return w
}

func (i *Worker[T, E]) work(request T) {
	defer i.Release()

	if res := i.Work(request); reflect.TypeOf(res) != reflect.TypeOf(None) {
		i.ResultChan <- res
	}
}

func (i *Worker[T, E]) Acquire() {
	err := i.Semaphore.Acquire(context.Background(), 1)
	if err != nil {
		log.Println("failed to acquire lock")
	}
}

func (i *Worker[T, E]) Release() {
	i.Semaphore.Release(1)
}

func (i Worker[T, E]) Validate() error {
	if i.ConcurrentCount < 1 {
		return ErrMinConcurrentCount
	}

	return nil
}
