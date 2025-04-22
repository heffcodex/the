package tchk

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

var (
	ErrStopped    = errors.New("stopped")
	ErrNotChecked = errors.New("not checked")
)

type HealthFunc func(ctx context.Context) error

type HealthChecker struct {
	mu          sync.RWMutex
	timeout     time.Duration
	interval    time.Duration
	maxJobs     int64
	checks      map[string]HealthFunc
	status      map[string]error
	statusAllOk bool
	stop        chan struct{}
}

// NewHealthChecker creates a new healthcheck worker.
//
// `interval` is a delay between consecutive background status updates.
// Zero `interval` means that HealthChecker.Run() is no-op and status update is performed explicitly on HealthChecker.Status() call.
//
// `timeout` is used to construct a time-limited context for a single status update.
// Zero `timeout` allows the update to run for an unlimited time or unless HealthChecker.Stop() is called.
func NewHealthChecker(interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		timeout:  timeout,
		interval: interval,
		maxJobs:  int64(runtime.NumCPU())*2 + 1,
		checks:   make(map[string]HealthFunc),
		status:   make(map[string]error),
		stop:     make(chan struct{}),
	}
}

func (c *HealthChecker) Register(service string, check HealthFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.checks[service]; ok {
		panic(fmt.Sprintf("`%s` healthcheck is already registered", service))
	}

	c.checks[service] = check
	c.status[service] = ErrNotChecked
	c.statusAllOk = false
}

func (c *HealthChecker) Health(ctx context.Context) error {
	statusMap, allOk := c.Status(ctx)
	if allOk {
		return nil
	}

	var errs error

	for service, err := range statusMap {
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("%s: %w", service, err))
		}
	}

	return errs
}

func (c *HealthChecker) Status(ctx context.Context) (map[string]error, bool) {
	if c.interval == 0 {
		c.updateChecks(ctx)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.status, c.statusAllOk
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
}

func (c *HealthChecker) updateChecks(ctx context.Context) {
	c.mu.RLock()

	if len(c.checks) == 0 {
		c.mu.RUnlock()
		return
	}

	ctx, cancel := c.getContext(ctx)
	defer cancel()

	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		sem       = semaphore.NewWeighted(c.maxJobs)
		status    = make(map[string]error, len(c.checks))
		statusMu  sync.Mutex
		hasErrors bool
	)

	for service := range c.checks {
		c.goCheck(ctx, sem, service, func(err error) {
			statusMu.Lock()
			defer statusMu.Unlock()

			status[service] = err
			hasErrors = hasErrors || err != nil
		})
	}

	_ = sem.Acquire(context.WithoutCancel(ctx), c.maxJobs)

	c.status = status
	c.statusAllOk = !hasErrors
}

func (c *HealthChecker) goCheck(ctx context.Context, sem *semaphore.Weighted, service string, onResult func(error)) {
	if err := sem.Acquire(ctx, 1); err != nil {
		onResult(err)
		return
	}

	go func() {
		defer sem.Release(1)
		onResult(c.checks[service](ctx))
	}()
}

func (c *HealthChecker) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	if c.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
	}

	go func() {
		select {
		case <-c.stop:
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	return ctx, cancel
}
