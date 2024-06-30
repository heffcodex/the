package the

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const (
	DefaultShutdownTimeout = 10 * time.Second
)

type shutter struct {
	// set by newShutter
	signals    []os.Signal
	notifyChan chan os.Signal
	hasWaiter  atomic.Bool
	inShutdown atomic.Bool

	// set by setup
	wasSetup   atomic.Bool
	log        *zap.Logger
	cancelFn   context.CancelFunc
	onShutdown CloseFunc
	timeout    time.Duration
}

func newShutter(signals ...os.Signal) *shutter {
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}

	return &shutter{
		signals: signals,
	}
}

func (s *shutter) setup(log *zap.Logger, cancelFn context.CancelFunc, onShutdown CloseFunc, timeout time.Duration) *shutter {
	if !s.wasSetup.CompareAndSwap(false, true) {
		panic("shutter setup called twice")
	}

	if timeout <= 0 {
		timeout = DefaultShutdownTimeout
	}

	s.log = log
	s.cancelFn = cancelFn
	s.onShutdown = onShutdown
	s.timeout = timeout

	return s
}

func (s *shutter) waitInterrupt() {
	firstWaiter := s.hasWaiter.CompareAndSwap(false, true)

	if firstWaiter {
		s.notifyChan = make(chan os.Signal, len(s.signals))
		signal.Notify(s.notifyChan, s.signals...)

		defer func() {
			s.log.Debug("shutdown interrupt")

			signal.Stop(s.notifyChan)
			close(s.notifyChan)
		}()
	}

	<-s.notifyChan
}

func (s *shutter) down() {
	if !s.wasSetup.Load() || !s.inShutdown.CompareAndSwap(false, true) {
		return
	}

	s.log.Info("shutdown start", zap.Duration("timeout", s.timeout))
	s.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer func() {
		cancel()
		_ = s.log.Sync() //nolint: wsl // it's ok
	}()

	onShutdownErr := make(chan error)

	go func() {
		defer close(onShutdownErr)

		if s.onShutdown != nil {
			onShutdownErr <- s.onShutdown(ctx)
		}
	}()

	var err error

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-onShutdownErr:
	}

	if err == nil {
		s.log.Info("shutdown complete")
	} else {
		s.log.Error("shutdown error", zap.Error(err))
	}
}

func (s *shutter) cancel() {
	s.cancelFn()
}
