package throttle

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/samsarahq/go/oops"
	"github.com/spkg/slog"
)

type ThrottleCallback[T any] func(ctx context.Context, key string, items []T, instantCallback bool) error

type ThrottleItem[T any] struct {
	mu sync.Mutex

	key             string
	callback        ThrottleCallback[T]
	releaseInterval time.Duration

	items []T

	lastReleaseTime time.Time
	timer           *time.Timer
}

func (i *ThrottleItem[T]) flush(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-i.timer.C:
			i.mu.Lock()
			defer i.mu.Unlock()
			tmp := make([]T, len(i.items))
			copy(tmp, i.items)
			i.items = nil
			i.lastReleaseTime = time.Now()
			go func() {
				err := i.callback(ctx, i.key, tmp, false)
				if err != nil {
					slog.Error(ctx, oops.Wrapf(err, "could not execute throttle callback from throttleitem flush").Error())
				}
			}()
			return
		}
	}
}

func (i *ThrottleItem[T]) Submit(ctx context.Context, item T) {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()

	if now.Sub(i.lastReleaseTime) > i.releaseInterval && len(i.items) == 0 {
		i.lastReleaseTime = now
		go func() {
			err := i.callback(ctx, i.key, []T{item}, true)
			if err != nil {
				slog.Error(ctx, oops.Wrapf(err, "could not execute throttle callback from thottleitem submit").Error())
			}
		}()
		return
	}

	i.items = append(i.items, item)
	if len(i.items) == 1 {
		i.timer = time.NewTimer(i.releaseInterval)
		go i.flush(ctx)
	} else {
		if !i.timer.Stop() {
			<-i.timer.C
		}
		i.timer.Reset(i.releaseInterval)
	}
}

type Throttler[T any] struct {
	sync.Mutex
	callback        ThrottleCallback[T]
	throttleItems   map[string]*ThrottleItem[T]
	releaseInterval time.Duration
}

func New[T any](callback ThrottleCallback[T], releaseInterval time.Duration) *Throttler[T] {
	throttleItems := make(map[string]*ThrottleItem[T])
	return &Throttler[T]{
		callback:        callback,
		throttleItems:   throttleItems,
		releaseInterval: releaseInterval,
	}
}

// Submit will perform the callback immediately if no items of the given key
// have been processed in recent time. Otherwise, it will throttle the item and
// only execute it after releaseInterval has passed since the previous item.
// If new items come in with that key before releaseInterval, then the time
// at which they will all be released will be postponed accordingly.
func (q *Throttler[T]) Submit(ctx context.Context, key string, item T) {
	keyLower := strings.ToLower(key)
	q.Lock()
	throttlerItem, ok := q.throttleItems[keyLower]
	if !ok {
		throttlerItem = &ThrottleItem[T]{
			key:             keyLower,
			callback:        q.callback,
			releaseInterval: q.releaseInterval,
		}
		q.throttleItems[keyLower] = throttlerItem
	}
	q.Unlock()

	throttlerItem.Submit(ctx, item)
}
