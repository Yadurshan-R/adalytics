package koios

import (
	"context"
	"sync"
)

// Workers is how many requests run at the same time when a long list is read
// in batches. Koios answers a big tx_info request in about 4 to 5 seconds, so
// three at a time read three times as fast; the rate limiter (90 requests per
// 10 seconds) still holds. It is the same number of requests either way.
const Workers = 3

// Parallel runs fn for every job from 0 to n-1, at most workers at a time, and
// returns the first error. Once a job fails, jobs not started yet are skipped.
func Parallel(ctx context.Context, n, workers int, fn func(i int) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var (
		wg    sync.WaitGroup
		once  sync.Once
		first error
	)
	jobs := make(chan int)
	for w := 0; w < min(workers, n); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				if err := fn(i); err != nil {
					once.Do(func() { first = err; cancel() })
				}
			}
		}()
	}
	for i := 0; i < n; i++ {
		select {
		case jobs <- i:
		case <-ctx.Done():
		}
		if ctx.Err() != nil {
			break
		}
	}
	close(jobs)
	wg.Wait()
	if first == nil && ctx.Err() != nil {
		return context.Cause(ctx)
	}
	return first
}
