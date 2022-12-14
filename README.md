<h1 align="center">
<img alt="Koi logo" src="asset/logo.webp" width="500px"/><br/>
KOI
</h1>

<p align="center">Generic Goroutine and Worker Manager</p>

<p align="center">

<a href="https://pkg.go.dev/github.com/1995parham/koi?tab=doc" target="_blank">
<img src="https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go" alt="go version" />
</a>

<img src="https://img.shields.io/badge/license-apache_2.0-red?style=for-the-badge&logo=none" alt="license" />



<img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/1995parham/koi/ci.yaml?style=for-the-badge" />

</p>

# Installation

```bash
go get github.com/1995parham/koi
```

# Usage

In Koi you first register a worker on a Pond then push your inputs.
Your worker has concurrency configuration for handling inputs.
Worker has generic interface. The first generic parameter is an input rype and the second parameter
is an output parameter.

```go
package main

import (
 "log"
 "time"

 "github.com/1995parham/koi"
)

func main() {
 pond := koi.NewPond[int, int]()

 printWorker := koi.Worker[int, int]{
  ConcurrentCount: 2,
  QueueSize:       10,
  Work: func(a int) *int {
   time.Sleep(1 * time.Second)
   log.Println(a)

   return nil
  },
 }

 _ = pond.RegisterWorker("printer", printWorker)

 for i := 0; i < 10; i++ {
  _, err := pond.AddWork("printer", i)
  if err != nil {
   log.Printf("error while adding job: %v\n", err)
  }
 }

 log.Println("all job added")

 for {

 }
}
```

**Note**: `pond.AddWork` is non-blocking unless worker queue is full.

# Terminology

- **Koi**: Koi is an informal name for the colored variants of C. rubrofuscus kept for ornamental purposes.
- **Pond**: an area of water smaller than a lake, often artificially made.
