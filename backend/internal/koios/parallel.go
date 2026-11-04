package koios

import (
	"context"
	"sync"
)

// DefaultWorkers is how many requests run at the same time when a long list is
// read in batches (KOIOS_PARALLEL in .env changes it, see SetWorkers). Koios
// answers a big tx_info request in about 4 to 5 seconds, so 8 at a time read
// 8 times as fast as one by one: about 18 requests per 10 seconds, well under
// the free tier's 100. It is the same number of requests either way.
const DefaultWorkers = 8

// MaxWorkers is the most KOIOS_PARALLEL may be. More requests at once make
// Koios slower for each of them, some then hit its 30-second query limit and
// fail, and over 100 requests per 10 seconds Koios blocks for a minute.
const MaxWorkers = 20

// SetWorkers sets how many requests run at the same time when a long list is
// read in batches, from 1 to MaxWorkers. Call it before the first request.
func (c *Client) SetWorkers(n int) { c.workers = min(max(n, 1), MaxWorkers) }

// Workers is how many requests run at the same time when a long list is read.
func (c *Client) Workers() int { return c.workers }

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
