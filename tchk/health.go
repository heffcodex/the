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

var (
	ErrStopped = errors.New("stopped")
	ErrUnknown = errors.New("unknown")
)

type HealthFunc func(ctx context.Context) error

type HealthChecker struct {
	mu       sync.RWMutex
	timeout  time.Duration
	interval time.Duration
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
// Zero or negative `timeout` allows the update to run for an unlimited time,
// unless either a higher-level context passed to HealthChecker.Health() is expired or HealthChecker.Stop() is called.
func NewHealthChecker(interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		timeout:  max(0, timeout),
		interval: max(0, interval),
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
	defer c.mu.RUnlock()

	return c.result
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
			c.updateChecks(context.Background())
			t.Reset(c.interval)
		}
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

	g, gCtx := c.newErrGroup(ctx)

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

func (c *HealthChecker) newErrGroup(ctx context.Context) (*errgroup.Group, context.Context) {
	var timeoutCancel context.CancelFunc = func() {}

	ctx, cancelCause := context.WithCancelCause(ctx)
	if c.timeout > 0 {
		ctx, timeoutCancel = context.WithTimeout(ctx, c.timeout)
	}

	go func() {
		defer timeoutCancel()
		select {
		case <-c.stop:
			cancelCause(ErrStopped)
		case <-ctx.Done():
			return
		}
	}()

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(c.maxJobs)

	return g, gCtx
}
