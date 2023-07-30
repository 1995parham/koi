package koi

import (
	"reflect"
)

type NoReturn int

const None NoReturn = 0

type Worker[T any, E any] struct {
	QueueSize       uint
	ConcurrentCount int
	Work            func(T) E

	ResultChan  chan E
	RequestChan chan T
}

func NewWoker[T any, E any](work func(T) E, queueSize uint, concurrentCount int) (*Worker[T, E], error) {
	w := &Worker[T, E]{
		QueueSize:       queueSize,
		ConcurrentCount: concurrentCount,
		Work:            work,
		ResultChan:      make(chan E, queueSize),
		RequestChan:     make(chan T, queueSize),
	}

	return w, w.Validate()
}

func MustNewWoker[T any, E any](work func(T) E, queueSize uint, concurrentCount int) *Worker[T, E] {
	w, err := NewWoker(work, queueSize, concurrentCount)
	if err != nil {
		panic(err)
	}

	return w
}

func (i *Worker[T, E]) work() {
	for request := range i.RequestChan {
		if res := i.Work(request); reflect.TypeOf(res) != reflect.TypeOf(None) {
			i.ResultChan <- res
		}
	}
}

func (i Worker[T, E]) Validate() error {
	if i.ConcurrentCount < 1 {
		return ErrMinConcurrentCount
	}

	return nil
}
