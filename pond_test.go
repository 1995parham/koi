package koi_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/1995parham/koi"
)

const (
	queueSize       = 2
	concurrentCount = 10
)

func TestNoReturn(t *testing.T) {
	t.Parallel()

	pond := koi.NewPond[int, koi.NoReturn]()

	var wg sync.WaitGroup

	printer := func(_ int) koi.NoReturn {
		time.Sleep(1 * time.Second)

		wg.Done()

		return koi.None
	}

	printWorker := koi.MustNewWoker(printer, queueSize, concurrentCount)

	pond.MustRegisterWorker("printer", printWorker)

	for i := 0; i < 10; i++ {
		wg.Add(1)

		if _, err := pond.AddWork("printer", i); err != nil {
			t.Errorf("error while adding job: %s", err)
		}
	}

	wg.Wait()
}

func TestWorkerNotFound(t *testing.T) {
	t.Parallel()

	pond := koi.NewPond[int, koi.NoReturn]()

	if _, err := pond.AddWork("printer", 1); !errors.Is(err, koi.ErrWorkerNotFound) {
		t.Error("expects not found error")
	}
}
