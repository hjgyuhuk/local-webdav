package handler

import (
	"context"
	"slices"
	"sync"
	"time"
)

const debounceWindow = 3 * time.Second

type fileSlot struct {
	mu     sync.Mutex
	seq    uint64
	cancel context.CancelFunc
	active int
}

type debounce struct {
	mu    sync.Mutex
	items map[string]*fileSlot
}

func newDebounce() *debounce {
	return &debounce{items: make(map[string]*fileSlot)}
}

func (d *debounce) getOrCreate(key string) *fileSlot {
	slot, exists := d.items[key]
	if !exists {
		slot = &fileSlot{}
		d.items[key] = slot
	}
	return slot
}

func (d *debounce) wait(ctx context.Context, key string) bool {
	d.mu.Lock()
	slot := d.getOrCreate(key)
	slot.active++

	if slot.cancel != nil {
		slot.cancel()
	}

	slot.seq++
	mySeq := slot.seq

	waitCtx, cancel := context.WithCancel(ctx)
	slot.cancel = cancel
	d.mu.Unlock()

	select {
	case <-waitCtx.Done():
		d.mu.Lock()
		slot.active--
		if slot.active == 0 {
			delete(d.items, key)
		}
		d.mu.Unlock()
		return false
	case <-time.After(debounceWindow):
	}

	slot.mu.Lock()

	d.mu.Lock()
	if slot.seq != mySeq {
		slot.active--
		if slot.active == 0 {
			delete(d.items, key)
		}
		slot.mu.Unlock()
		d.mu.Unlock()
		return false
	}
	d.mu.Unlock()

	return true
}

func (d *debounce) done(key string) {
	d.mu.Lock()
	slot := d.items[key]
	if slot == nil {
		d.mu.Unlock()
		return
	}
	slot.active--
	if slot.active == 0 {
		delete(d.items, key)
	}
	slot.mu.Unlock()
	d.mu.Unlock()
}

func (d *debounce) waitAll(ctx context.Context, keys []string) (release func(), ok bool) {
	keys = uniqueSorted(keys)

	acquired := make([]string, 0, len(keys))
	for _, key := range keys {
		if !d.wait(ctx, key) {
			for i := len(acquired) - 1; i >= 0; i-- {
				d.done(acquired[i])
			}
			return nil, false
		}
		acquired = append(acquired, key)
	}

	return func() {
		for i := len(acquired) - 1; i >= 0; i-- {
			d.done(acquired[i])
		}
	}, true
}

func uniqueSorted(keys []string) []string {
	slices.Sort(keys)
	return slices.Compact(keys)
}
