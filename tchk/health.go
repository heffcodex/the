package tchk

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	HealthUpdateTimeoutDefault = 3 * time.Second
)

var (
	ErrStopped = errors.New("stopped")
	ErrUnknown = errors.New("unknown")
)

type HealthFunc func(ctx context.Context) error

type HealthChecker struct {
	mu       sync.RWMutex
	interval time.Duration
	timeout  time.Duration
	checks   []HealthFunc
	maxJobs  int
	result   error
	stop     chan struct{}
	stopped  atomic.Bool
}

// NewHealthChecker creates a new healthcheck worker.
//
// `interval` is a delay between consecutive background result updates.
// Zero or negative `interval` means HealthChecker.Run() is no-op,
// and the result update is performed explicitly on each HealthChecker.Health() call.
//
// `timeout` is used to construct a time-limited context for a single result update.
// Zero or negative `timeout` means that [HealthUpdateTimeoutDefault] will be used as its value.
func NewHealthChecker(interval, timeout time.Duration) *HealthChecker {
	if interval < 0 {
		interval = 0
	}

	if timeout <= 0 {
		timeout = HealthUpdateTimeoutDefault
	}

	return &HealthChecker{
		interval: interval,
		timeout:  timeout,
		maxJobs:  runtime.NumCPU()*2 + 1,
		stop:     make(chan struct{}),
	}
}

func (c *HealthChecker) Register(checks ...HealthFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.checks = append(c.checks, checks...)
	c.result = errors.Join(c.result, ErrUnknown)
}

func (c *HealthChecker) Health(ctx context.Context) error {
	if c.stopped.Load() {
		return ErrStopped
	}

	if c.interval == 0 {
		c.updateChecks(ctx)
	}

	c.mu.RLock()
	result := c.result
	c.mu.RUnlock()

	return result
}

func (c *HealthChecker) Run() error {
	if c.interval == 0 {
		<-c.stop
		return ErrStopped
	}

	t := time.NewTimer(0)
	defer t.Stop()

	for {
		select {
		case <-c.stop:
			return ErrStopped
		case <-t.C:
		}

		func() {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), c.timeout)
			defer cancel()

			stopCtx, stopCause := context.WithCancelCause(timeoutCtx)
			defer stopCause(nil)

			go func() {
				select {
				case <-c.stop:
					stopCause(ErrStopped)
				case <-stopCtx.Done():
					return
				}
			}()

			c.updateChecks(stopCtx)
			t.Reset(c.interval)
		}()
	}
}

func (c *HealthChecker) Stop() {
	close(c.stop)
	c.stopped.Store(true)
}

func (c *HealthChecker) updateChecks(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.checks) == 0 {
		c.result = nil
		return
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(c.maxJobs)

	for i, check := range c.checks {
		g.Go(func() error {
			if err := check(gCtx); err != nil {
				return fmt.Errorf("#%d: %w", i, err)
			}

			return nil
		})
	}

	c.result = g.Wait()
}
