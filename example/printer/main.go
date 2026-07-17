package main

import (
	"log"
	"sync"
	"time"

	"github.com/1995parham/koi"
)

func main() {
	pond := koi.NewPond[int, koi.NoReturn]()

	var wg sync.WaitGroup

	printer := func(a int) koi.NoReturn {
		time.Sleep(1 * time.Second)
		log.Println(a)

		wg.Done()

		return koi.None
	}

	//nolint:mnd
	printWorker := koi.MustNewWorker(printer, 2, 10)

	pond.MustRegisterWorker("printer", printWorker)

	for i := range 10 {
		wg.Add(1)

		if _, err := pond.AddWork("printer", i); err != nil {
			log.Printf("error while adding job: %s\n", err)
		}
	}

	wg.Wait()

	// stop the workers and release their goroutines.
	pond.Close()

	log.Println("all jobs done")
}
