package algorithm

import (
	"sync"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
    c := make(chan struct{})
    go func() {
        defer close(c)
        wg.Wait()
    }()
    select {
    case <-c:
        return false // completed normally
    case <-time.After(timeout):
        return true // timed out
    }
}