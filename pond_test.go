package koi_test

import (
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/1995parham/koi"
)

const (
	queueSize       = 0
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

	printWorker := koi.MustNewWorker(printer, queueSize, concurrentCount)

	pond.MustRegisterWorker("printer", printWorker)

	for i := range concurrentCount {
		wg.Add(1)

		if _, err := pond.AddWork("printer", i); err != nil {
			t.Errorf("error while adding job: %s", err)
		}
	}

	wg.Wait()
}

func TestMapResults(t *testing.T) {
	t.Parallel()

	pond := koi.NewPond[int, int]()

	square := func(i int) int {
		return i * i
	}

	pond.MustRegisterWorker("square", koi.MustNewWorker(square, queueSize, concurrentCount))

	// generic method (go1.27): map int results to their string form.
	strs := pond.MapResults("square", func(n int) string {
		return strconv.Itoa(n)
	})

	for i := range concurrentCount {
		if _, err := pond.AddWork("square", i); err != nil {
			t.Errorf("error while adding job: %s", err)
		}
	}

	got := make(map[string]bool)
	for range concurrentCount {
		got[<-strs] = true
	}

	for i := range concurrentCount {
		if want := strconv.Itoa(i * i); !got[want] {
			t.Errorf("cannot find mapped result %q", want)
		}
	}

	pond.Close()

	// MapResults for an unknown worker yields nil.
	if ch := pond.MapResults("missing", func(n int) string { return "" }); ch != nil {
		t.Error("expects nil channel for unknown worker")
	}
}

func TestClose(t *testing.T) {
	t.Parallel()

	pond := koi.NewPond[int, int]()

	square := func(i int) int {
		return i * i
	}

	pond.MustRegisterWorker("square", koi.MustNewWorker(square, queueSize, concurrentCount))

	for i := range concurrentCount {
		if _, err := pond.AddWork("square", i); err != nil {
			t.Errorf("error while adding job: %s", err)
		}
	}

	ch := pond.ResultChan("square")
	for range concurrentCount {
		<-ch
	}

	// Close drains in-flight work and closes the result channel.
	pond.Close()

	if _, ok := <-ch; ok {
		t.Error("expects result channel to be closed after Close")
	}

	// operations after Close must fail instead of panicking.
	if _, err := pond.AddWork("square", 1); !errors.Is(err, koi.ErrPondClosed) {
		t.Errorf("expects pond closed error, got: %v", err)
	}

	// Close is idempotent.
	pond.Close()
}

func TestWorkerNotFound(t *testing.T) {
	t.Parallel()

	pond := koi.NewPond[int, koi.NoReturn]()

	if _, err := pond.AddWork("printer", 1); !errors.Is(err, koi.ErrWorkerNotFound) {
		t.Error("expects not found error")
	}
}

func TestReturn(t *testing.T) {
	t.Parallel()

	pond := koi.NewPond[int, int]()

	square := func(i int) int {
		return i * i
	}

	printWorker := koi.MustNewWorker(square, queueSize, concurrentCount)

	pond.MustRegisterWorker("square", printWorker)

	for i := 0; i < concurrentCount; i++ {
		if _, err := pond.AddWork("square", i); err != nil {
			t.Errorf("error while adding job: %s", err)
		}
	}

	ch := pond.ResultChan("square")
	results := make(map[int]bool)

	for i := 0; i < concurrentCount; i++ {
		r := <-ch
		results[r] = true
	}

	for i := 0; i < concurrentCount; i++ {
		if _, ok := results[i*i]; !ok {
			t.Errorf("cannot find result for %d", i)
		}
	}
}
