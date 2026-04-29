package handler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDebounce_SingleRequest(t *testing.T) {
	d := newDebounce()

	ok := d.wait(context.Background(), "file1")
	if !ok {
		t.Fatal("expected single request to proceed")
	}
	d.done("file1")
}

func TestDebounce_LatestWins(t *testing.T) {
	d := newDebounce()

	var results [3]bool
	var wg sync.WaitGroup

	for i := range 3 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = d.wait(context.Background(), "file1")
			if results[idx] {
				d.done("file1")
			}
		}(i)

		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()

	winner := 0
	for _, ok := range results {
		if ok {
			winner++
		}
	}
	if winner != 1 {
		t.Errorf("expected exactly 1 winner, got %d (results: %v)", winner, results)
	}

	if !results[2] {
		t.Error("expected last request to be the winner")
	}
}

func TestDebounce_DifferentFiles(t *testing.T) {
	d := newDebounce()

	var wg sync.WaitGroup
	var count atomic.Int32

	for _, key := range []string{"file1", "file2", "file3"} {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			if d.wait(context.Background(), k) {
				count.Add(1)
				d.done(k)
			}
		}(key)
	}

	wg.Wait()

	if count.Load() != 3 {
		t.Errorf("expected 3 winners for different files, got %d", count.Load())
	}
}

func TestDebounce_CancelledByNewer(t *testing.T) {
	d := newDebounce()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	var result1, result2 bool

	wg.Add(2)

	go func() {
		defer wg.Done()
		result1 = d.wait(ctx, "file1")
		if result1 {
			d.done("file1")
		}
	}()

	time.Sleep(100 * time.Millisecond)

	go func() {
		defer wg.Done()
		result2 = d.wait(context.Background(), "file1")
		if result2 {
			d.done("file1")
		}
	}()

	wg.Wait()

	if result1 {
		t.Error("expected first request to be cancelled")
	}
	if !result2 {
		t.Error("expected second request to proceed")
	}
	cancel()
}

func TestDebounce_ExternalCancel(t *testing.T) {
	d := newDebounce()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	ok := d.wait(ctx, "file1")
	if ok {
		d.done("file1")
		t.Error("expected request to be cancelled by external context")
	}
}

func TestDebounce_ConcurrentStress(t *testing.T) {
	d := newDebounce()

	const goroutines = 20
	var wg sync.WaitGroup
	var winners atomic.Int32

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if d.wait(context.Background(), "shared") {
				winners.Add(1)
				d.done("shared")
			}
		}(i)

		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()

	if winners.Load() != 1 {
		t.Errorf("expected exactly 1 winner under concurrency, got %d", winners.Load())
	}
}

func TestDebounce_SlotCleanup(t *testing.T) {
	d := newDebounce()

	ok := d.wait(context.Background(), "file1")
	if !ok {
		t.Fatal("expected request to proceed")
	}
	d.done("file1")

	d.mu.Lock()
	remaining := len(d.items)
	d.mu.Unlock()
	if remaining != 0 {
		t.Errorf("expected 0 slots after done, got %d", remaining)
	}
}

func TestDebounce_SlotCleanupAfterCancel(t *testing.T) {
	d := newDebounce()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	d.wait(ctx, "file1")

	d.mu.Lock()
	remaining := len(d.items)
	d.mu.Unlock()
	if remaining != 0 {
		t.Errorf("expected 0 slots after cancel, got %d", remaining)
	}
}

func TestDebounce_SlotCleanupMultiple(t *testing.T) {
	d := newDebounce()

	for _, key := range []string{"a", "b", "c"} {
		if !d.wait(context.Background(), key) {
			t.Fatalf("expected %s to proceed", key)
		}
		d.done(key)
	}

	d.mu.Lock()
	remaining := len(d.items)
	d.mu.Unlock()
	if remaining != 0 {
		t.Errorf("expected 0 slots after all done, got %d", remaining)
	}
}

func TestWaitAll_SingleKey(t *testing.T) {
	d := newDebounce()

	release, ok := d.waitAll(context.Background(), []string{"k1"})
	if !ok {
		t.Fatal("expected waitAll to succeed")
	}
	release()

	d.mu.Lock()
	remaining := len(d.items)
	d.mu.Unlock()
	if remaining != 0 {
		t.Errorf("expected 0 slots after release, got %d", remaining)
	}
}

func TestWaitAll_MultipleKeys(t *testing.T) {
	d := newDebounce()

	release, ok := d.waitAll(context.Background(), []string{"k2", "k1"})
	if !ok {
		t.Fatal("expected waitAll to succeed")
	}
	release()

	d.mu.Lock()
	remaining := len(d.items)
	d.mu.Unlock()
	if remaining != 0 {
		t.Errorf("expected 0 slots after release, got %d", remaining)
	}
}

func TestWaitAll_SameDestinationSerializes(t *testing.T) {
	d := newDebounce()

	var active atomic.Int32
	var maxConcurrent atomic.Int32
	var wg sync.WaitGroup

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, ok := d.waitAll(context.Background(), []string{"src:/a.txt", "dst:/x.txt"})
			if !ok {
				return
			}
			defer release()

			cur := active.Add(1)
			for {
				old := maxConcurrent.Load()
				if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			active.Add(-1)
		}()
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()

	if maxConcurrent.Load() != 1 {
		t.Errorf("expected max concurrent = 1 for shared destination, got %d", maxConcurrent.Load())
	}
}

func TestWaitAll_PartialFailureReleasesAcquired(t *testing.T) {
	d := newDebounce()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	release, ok := d.waitAll(ctx, []string{"k1", "k2"})
	if ok {
		release()
		t.Fatal("expected waitAll to fail due to context cancellation")
	}

	d.mu.Lock()
	remaining := len(d.items)
	d.mu.Unlock()
	if remaining != 0 {
		t.Errorf("expected 0 slots after partial failure, got %d", remaining)
	}
}
