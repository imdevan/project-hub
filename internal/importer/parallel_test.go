package importer

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunParallelBoundsWorkers(t *testing.T) {
	var running, peak atomic.Int32
	n, jobs := 20, 4
	events := RunParallel(context.Background(), n, jobs, func(i int) Result {
		cur := running.Add(1)
		for {
			p := peak.Load()
			if cur <= p || peak.CompareAndSwap(p, cur) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		running.Add(-1)
		return Result{Action: Pulled}
	})
	done := map[int]bool{}
	for ev := range events {
		if ev.Done {
			if done[ev.Index] {
				t.Fatalf("job %d finished twice", ev.Index)
			}
			done[ev.Index] = true
		}
	}
	if len(done) != n {
		t.Fatalf("got %d done events, want %d", len(done), n)
	}
	if p := peak.Load(); p > int32(jobs) || p < 2 {
		t.Fatalf("peak concurrency %d, want between 2 and %d", p, jobs)
	}
}

func TestRunParallelCancelSkipsUnstarted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var ran atomic.Int32
	events := RunParallel(ctx, 10, 1, func(i int) Result {
		ran.Add(1)
		cancel() // first job cancels; the rest must not start
		return Result{Action: Pulled}
	})
	var skipped int
	for ev := range events {
		if ev.Done && ev.Result.Action == Skipped && ev.Result.Detail == "interrupted" {
			skipped++
		}
	}
	if ran.Load() != 1 || skipped != 9 {
		t.Fatalf("ran %d, skipped %d; want 1 and 9", ran.Load(), skipped)
	}
}

func TestParentIndex(t *testing.T) {
	dests := []string{"/ws/a", "/ws/a-x", "/ws/a/b", "/ws/a/b/c", "/ws/a/c", "/ws/b"}
	want := []int{-1, -1, 0, 2, 0, -1}
	if got := ParentIndex(dests); !reflect.DeepEqual(got, want) {
		t.Fatalf("ParentIndex = %v, want %v", got, want)
	}
}
