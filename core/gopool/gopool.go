package gopool

import (
	"github.com/meta-apex/zenith/zlog"
	"time"

	"github.com/panjf2000/ants/v2"
)

const (
	// DefaultAntsPoolSize sets up the capacity of worker pool, 256 * 1024.
	DefaultAntsPoolSize = 1 << 18

	// ExpiryDuration is the interval time to clean up those expired workers.
	ExpiryDuration = 10 * time.Second

	// Nonblocking decides what to do when submitting a new task to a full worker pool: waiting for a available worker
	// or returning nil directly.
	Nonblocking = true
)

func init() {
	// It releases the default pool from ants.
	ants.Release()
}

// DefaultWorkerPool is the global worker pool.
var DefaultWorkerPool = Default()

// Pool is the alias of ants.Pool.
type Pool = ants.Pool

type antsLogger struct {
	*zlog.Logger
}

// Printf implements the ants.Logger interface.
func (l antsLogger) Printf(format string, args ...any) {
	l.Info().Msgf(format, args...)
}

// Default instantiates a non-blocking goroutine pool with the capacity of DefaultAntsPoolSize.
func Default() *Pool {
	options := ants.Options{
		ExpiryDuration: ExpiryDuration,
		Nonblocking:    Nonblocking,
		Logger:         &antsLogger{zlog.GetDefaultLogger().WithName("gopool")},
		PanicHandler: func(a any) {
			zlog.Error().Msgf("goroutine pool panic: %v", a)
		},
	}

	defaultAntsPool, _ := ants.NewPool(DefaultAntsPoolSize, ants.WithOptions(options))
	return defaultAntsPool
}
