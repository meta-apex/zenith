//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package znet

import (
	"context"
	"errors"
	"github.com/meta-apex/zenith/core/zerror"
	"github.com/meta-apex/zenith/zlog"
	"github.com/meta-apex/zenith/znet/internal/netpoll"
	"github.com/meta-apex/zenith/znet/internal/queue"
	"github.com/meta-apex/zenith/znet/internal/socket"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

type engine struct {
	listeners    map[int]*listener // listeners for accepting incoming connections
	opts         *Options          // options with engine
	ingress      *eventloop        // main event-loop that monitors all listeners
	eventLoops   loadBalancer      // event-loops for handling events
	inShutdown   atomic.Bool       // whether the engine is in shutdown
	turnOff      context.CancelFunc
	eventHandler EventHandler // user eventHandler
	concurrency  struct {
		*errgroup.Group

		ctx context.Context
	}
}

func (eng *engine) isShutdown() bool {
	return eng.inShutdown.Load()
}

// shutdown signals the engine to shut down.
func (eng *engine) shutdown(err error) {
	if err != nil && !errors.Is(err, zerror.ErrEngineShutdown) {
		eng.opts.Logger.Error().Msgf("engine is being shutdown with error: %v", err)
	}
	// Cancel the context to stop the engine.
	eng.turnOff()
}

func (eng *engine) closeEventLoops() {
	eng.eventLoops.iterate(func(_ int, el *eventloop) bool {
		for _, ln := range el.listeners {
			ln.close()
		}
		_ = el.poller.Close()
		return true
	})
	if eng.ingress != nil {
		for _, ln := range eng.listeners {
			ln.close()
		}
		err := eng.ingress.poller.Close()
		if err != nil {
			eng.opts.Logger.Error().Msgf("failed to close poller when stopping engine: %v", err)
		}
	}
}

func (eng *engine) runEventLoops(ctx context.Context, numEventLoop int) error {
	var el0 *eventloop
	lns := eng.listeners
	// Create loops locally and bind the listeners.
	for i := 0; i < numEventLoop; i++ {
		if i > 0 {
			lns = make(map[int]*listener, len(eng.listeners))
			for _, l := range eng.listeners {
				ln, err := initListener(l.network, l.address, eng.opts)
				if err != nil {
					return err
				}
				lns[ln.fd] = ln
			}
		}
		p, err := netpoll.OpenPoller()
		if err != nil {
			return err
		}
		el := new(eventloop)
		el.listeners = lns
		el.engine = eng
		el.poller = p
		el.buffer = make([]byte, eng.opts.ReadBufferCap)
		el.connections.init()
		el.eventHandler = eng.eventHandler
		for _, ln := range lns {
			if err = el.poller.AddRead(ln.packPollAttachment(el.accept), false); err != nil {
				return err
			}
		}
		eng.eventLoops.register(el)

		// Start the ticker.
		if eng.opts.Ticker && el.idx == 0 {
			el0 = el
		}
	}

	// Start event-loops in the background.
	eng.eventLoops.iterate(func(_ int, el *eventloop) bool {
		eng.concurrency.Go(el.run)
		return true
	})

	if el0 != nil {
		eng.concurrency.Go(func() error {
			el0.ticker(ctx)
			return nil
		})
	}

	return nil
}

func (eng *engine) activateReactors(ctx context.Context, numEventLoop int) error {
	for i := 0; i < numEventLoop; i++ {
		p, err := netpoll.OpenPoller()
		if err != nil {
			return err
		}
		el := new(eventloop)
		el.listeners = eng.listeners
		el.engine = eng
		el.poller = p
		el.buffer = make([]byte, eng.opts.ReadBufferCap)
		el.connections.init()
		el.eventHandler = eng.eventHandler
		eng.eventLoops.register(el)
	}

	// Start sub reactors in the background.
	eng.eventLoops.iterate(func(_ int, el *eventloop) bool {
		eng.concurrency.Go(el.orbit)
		return true
	})

	p, err := netpoll.OpenPoller()
	if err != nil {
		return err
	}
	el := new(eventloop)
	el.listeners = eng.listeners
	el.idx = -1
	el.engine = eng
	el.poller = p
	el.eventHandler = eng.eventHandler
	for _, ln := range eng.listeners {
		if err = el.poller.AddRead(ln.packPollAttachment(el.accept0), true); err != nil {
			return err
		}
	}
	eng.ingress = el

	// Start the main reactor in the background.
	eng.concurrency.Go(el.rotate)

	// Start the ticker.
	if eng.opts.Ticker {
		eng.concurrency.Go(func() error {
			eng.ingress.ticker(ctx)
			return nil
		})
	}

	return nil
}

func (eng *engine) start(ctx context.Context, numEventLoop int) error {
	if eng.opts.ReusePort {
		return eng.runEventLoops(ctx, numEventLoop)
	}

	return eng.activateReactors(ctx, numEventLoop)
}

func (eng *engine) stop(ctx context.Context, s Engine) {
	// Wait on a signal for shutdown
	<-ctx.Done()

	eng.eventHandler.OnShutdown(s)

	// Notify all event-loops to exit.
	eng.eventLoops.iterate(func(i int, el *eventloop) bool {
		err := el.poller.Trigger(queue.HighPriority, func(_ any) error { return zerror.ErrEngineShutdown }, nil)
		if err != nil {
			eng.opts.Logger.Error().Msgf("failed to enqueue shutdown signal of high-priority for event-loop(%d): %v", i, err)
		}
		return true
	})
	if eng.ingress != nil {
		err := eng.ingress.poller.Trigger(queue.HighPriority, func(_ any) error { return zerror.ErrEngineShutdown }, nil)
		if err != nil {
			eng.opts.Logger.Error().Msgf("failed to enqueue shutdown signal of high-priority for main event-loop: %v", err)
		}
	}

	if err := eng.concurrency.Wait(); err != nil {
		eng.opts.Logger.Error().Msgf("engine shutdown error: %v", err)
	}

	// Close all listeners and pollers of event-loops.
	eng.closeEventLoops()

	// Put the engine into the shutdown state.
	eng.inShutdown.Store(true)
}

func run(eventHandler EventHandler, listeners []*listener, options *Options, addrs []string) error {
	numEventLoop := determineEventLoops(options)
	zlog.Info().Msgf("Starting znet with %d event-loops, listening on: %s", numEventLoop, strings.Join(addrs, " | "))

	lns := make(map[int]*listener, len(listeners))
	for _, ln := range listeners {
		lns[ln.fd] = ln
	}
	rootCtx, shutdown := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(rootCtx)
	eng := engine{
		listeners:    lns,
		opts:         options,
		turnOff:      shutdown,
		eventHandler: eventHandler,
		concurrency: struct {
			*errgroup.Group
			ctx context.Context
		}{eg, ctx},
	}
	switch options.LB {
	case RoundRobin:
		eng.eventLoops = new(roundRobinLoadBalancer)
	case LeastConnections:
		eng.eventLoops = new(leastConnectionsLoadBalancer)
	case SourceAddrHash:
		eng.eventLoops = new(sourceAddrHashLoadBalancer)
	}

	e := Engine{&eng}
	switch eng.eventHandler.OnBoot(e) {
	case None, Close:
	case Shutdown:
		return nil
	}

	if err := eng.start(ctx, numEventLoop); err != nil {
		eng.closeEventLoops()
		eng.opts.Logger.Error().Msgf("znet engine is stopping with error: %v", err)
		return err
	}
	defer eng.stop(rootCtx, e)

	for _, addr := range addrs {
		allEngines.Store(addr, &eng)
	}

	return nil
}

func setKeepAlive(fd int, enabled bool, idle, intvl time.Duration, cnt int) error {
	if intvl == 0 {
		intvl = idle / 5
	}
	if cnt == 0 {
		cnt = 5
	}
	return socket.SetKeepAlive(fd, enabled, int(idle.Seconds()), int(intvl.Seconds()), cnt)
}

/*
func (eng *engine) sendCmd(cmd *asyncCmd, urgent bool) error {
	if !gfd.Validate(cmd.fd) {
		return errors.ErrInvalidConn
	}
	el := eng.eventLoops.index(cmd.fd.EventLoopIndex())
	if el == nil {
		return errors.ErrInvalidConn
	}
	if urgent {
		return el.poller.Trigger(queue.LowPriority, el.execCmd, cmd)
	}
	return el.poller.Trigger(el.execCmd, cmd)
}
*/
