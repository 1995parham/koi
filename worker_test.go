package koi_test

import (
	"errors"
	"testing"

	"github.com/1995parham/koi"
)

func TestNewWorker(t *testing.T) {
	t.Parallel()

	printer := func(n int) koi.NoReturn {
		return koi.None
	}

	cases := []struct {
		concurrentCount int
		expectedErr     error
	}{
		{
			concurrentCount: 10,
			expectedErr:     nil,
		},
		{
			concurrentCount: 1,
			expectedErr:     nil,
		},
		{
			concurrentCount: 0,
			expectedErr:     koi.ErrMinConcurrentCount,
		},
	}

	for _, c := range cases {
		if _, err := koi.NewWoker(printer, 10, c.concurrentCount); !errors.Is(err, c.expectedErr) {
			t.Error("worker creation failed")
		}
	}
}
