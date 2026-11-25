package importer

import (
	"context"
	"sync"
)

// DefaultJobs is the default number of repositories processed at once.
const DefaultJobs = 8

// Event reports a job starting (Done false) or finishing (Done true) in RunParallel.
type Event struct {
	Index  int
	Done   bool
	Result Result
}

// RunParallel runs run(0..n-1) on at most jobs workers, picking up jobs in index order.
// Every job produces exactly one Done event. Once ctx is cancelled no new jobs start:
// jobs already running finish normally, and jobs not yet started are reported as Skipped
// ("interrupted"). The returned channel is closed after the last event.
func RunParallel(ctx context.Context, n, jobs int, run func(i int) Result) <-chan Event {
	if jobs < 1 {
		jobs = 1
	}
	events := make(chan Event, 2*n)
	queue := make(chan int, n)
	for i := 0; i < n; i++ {
		queue <- i
	}
	close(queue)

	var wg sync.WaitGroup
	for w := 0; w < min(jobs, max(n, 1)); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range queue {
				if ctx.Err() != nil {
					events <- Event{Index: i, Done: true, Result: Result{Action: Skipped, Detail: "interrupted"}}
					continue
				}
				events <- Event{Index: i}
				events <- Event{Index: i, Done: true, Result: run(i)}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(events)
	}()
	return events
}

// ParentIndex returns, for each destination, the index of the closest other destination
// that contains it (or -1). dests must be sorted so that a parent precedes its children,
// which holds for lexically sorted paths. Import uses this so a project nested inside
// another is only cloned after its parent, even when running in parallel.
func ParentIndex(dests []string) []int {
	parents := make([]int, len(dests))
	for i := range dests {
		parents[i] = -1
		for j := i - 1; j >= 0; j-- {
			if within(dests[i], dests[j]) {
				parents[i] = j
				break
			}
		}
	}
	return parents
}
