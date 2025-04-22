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
	ErrStopped           = errors.New("stopped")
	ErrNotChecked        = errors.New("not checked")
	ErrAlreadyRegistered = errors.New("already registered")
)

type HealthFunc func(ctx context.Context) error

type HealthChecker struct {
	mu       sync.RWMutex
	timeout  time.Duration
	interval time.Duration
	maxJobs  int
	checks   map[string]HealthFunc
	result   error
	stop     chan struct{}
	stopped  atomic.Bool
}

// NewHealthChecker creates a new healthcheck worker.
//
// `interval` is a delay between consecutive background result updates.
// Zero `interval` means HealthChecker.Run() is no-op and the result update is performed explicitly on each HealthChecker.Health() call.
//
// `timeout` is used to construct a time-limited context for a single result update.
// Zero `timeout` allows the update to run for an unlimited time,
// unless either a higher-level context passed to HealthChecker.Health() is expired or HealthChecker.Stop() is called.
func NewHealthChecker(interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		timeout:  timeout,
		interval: interval,
		maxJobs:  runtime.NumCPU()*2 + 1,
		checks:   make(map[string]HealthFunc),
		stop:     make(chan struct{}),
	}
}

func (c *HealthChecker) Register(service string, check HealthFunc) error {
	if c.stopped.Load() {
		return ErrStopped
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.checks[service]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, service)
	}

	c.checks[service] = check
	c.result = errors.Join(c.result, fmt.Errorf("%w: %s", ErrNotChecked, service))

	return nil
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

	for service, check := range c.checks {
		g.Go(func() error {
			if err := check(gCtx); err != nil {
				return fmt.Errorf("%s: %w", service, err)
			}

			return nil
		})
	}

	c.result = g.Wait()
}

func (c *HealthChecker) newErrGroup(ctx context.Context) (*errgroup.Group, context.Context) {
	var cancel context.CancelFunc

	if c.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	go func() {
		select {
		case <-c.stop:
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(c.maxJobs)

	return g, gCtx
}
