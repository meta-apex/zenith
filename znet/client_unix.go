//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package znet

import (
	"context"
	"errors"
	"github.com/meta-apex/zenith/core/zerror"
	"github.com/meta-apex/zenith/core/zmath"
	"github.com/meta-apex/zenith/zlog"
	"github.com/meta-apex/zenith/znet/internal/buffer/ring"
	"github.com/meta-apex/zenith/znet/internal/netpoll"
	"github.com/meta-apex/zenith/znet/internal/queue"
	"github.com/meta-apex/zenith/znet/internal/socket"
	"net"
	"strconv"
	"syscall"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

// Client of znet.
type Client struct {
	opts *Options
	eng  *engine
}

// NewClient creates an instance of Client.
func NewClient(eh EventHandler, opts ...Option) (cli *Client, err error) {
	options := loadOptions(opts...)
	cli = new(Client)
	cli.opts = options

	if options.Logger == nil {
		options.Logger = zlog.GetDefaultLogger().WithName("znet-client")
	}

	rootCtx, shutdown := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(rootCtx)
	eng := engine{
		listeners:    make(map[int]*listener),
		opts:         options,
		turnOff:      shutdown,
		eventHandler: eh,
		eventLoops:   new(leastConnectionsLoadBalancer),
		concurrency: struct {
			*errgroup.Group
			ctx context.Context
		}{eg, ctx},
	}

	if options.EdgeTriggeredIOChunk > 0 {
		options.EdgeTriggeredIO = true
		options.EdgeTriggeredIOChunk = zmath.CeilToPowerOfTwo(options.EdgeTriggeredIOChunk)
	} else if options.EdgeTriggeredIO {
		options.EdgeTriggeredIOChunk = 1 << 20 // 1MB
	}

	rbc := options.ReadBufferCap
	switch {
	case rbc <= 0:
		options.ReadBufferCap = MaxStreamBufferCap
	case rbc <= ring.DefaultBufferSize:
		options.ReadBufferCap = ring.DefaultBufferSize
	default:
		options.ReadBufferCap = zmath.CeilToPowerOfTwo(rbc)
	}
	wbc := options.WriteBufferCap
	switch {
	case wbc <= 0:
		options.WriteBufferCap = MaxStreamBufferCap
	case wbc <= ring.DefaultBufferSize:
		options.WriteBufferCap = ring.DefaultBufferSize
	default:
		options.WriteBufferCap = zmath.CeilToPowerOfTwo(wbc)
	}
	cli.eng = &eng
	return
}

// Start starts the client event-loop, handing IO events.
func (cli *Client) Start() error {
	numEventLoop := determineEventLoops(cli.opts)
	zlog.Info().Msgf("Starting znet client with %d event loops", numEventLoop)

	cli.eng.eventHandler.OnBoot(Engine{cli.eng})

	var el0 *eventloop
	for i := 0; i < numEventLoop; i++ {
		p, err := netpoll.OpenPoller()
		if err != nil {
			cli.eng.closeEventLoops()
			return err
		}
		el := eventloop{
			listeners:    cli.eng.listeners,
			engine:       cli.eng,
			poller:       p,
			buffer:       make([]byte, cli.opts.ReadBufferCap),
			eventHandler: cli.eng.eventHandler,
		}
		el.connections.init()
		cli.eng.eventLoops.register(&el)
		if cli.opts.Ticker && el.idx == 0 {
			el0 = &el
		}
	}

	cli.eng.eventLoops.iterate(func(_ int, el *eventloop) bool {
		cli.eng.concurrency.Go(el.run)
		return true
	})

	// Start the ticker.
	if el0 != nil {
		ctx := cli.eng.concurrency.ctx
		cli.eng.concurrency.Go(func() error {
			el0.ticker(ctx)
			return nil
		})
	}

	return nil
}

// Stop stops the client event-loop.
func (cli *Client) Stop() error {
	cli.eng.shutdown(nil)

	cli.eng.eventHandler.OnShutdown(Engine{cli.eng})

	// Notify all event-loops to exit.
	cli.eng.eventLoops.iterate(func(_ int, el *eventloop) bool {
		zlog.Error().Err(el.poller.Trigger(queue.HighPriority, func(_ any) error { return zerror.ErrEngineShutdown }, nil)).Msg("")
		return true
	})

	// Wait for all event-loops to exit.
	err := cli.eng.concurrency.Wait()

	cli.eng.closeEventLoops()

	// Put the engine into the shutdown state.
	cli.eng.inShutdown.Store(true)

	return err
}

// Dial is like net.Dial().
func (cli *Client) Dial(network, address string) (Conn, error) {
	return cli.DialContext(network, address, nil)
}

// DialContext is like Dial but also accepts an empty interface ctx that can be obtained later via Conn.Context.
func (cli *Client) DialContext(network, address string, ctx any) (Conn, error) {
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return cli.EnrollContext(c, ctx)
}

// Enroll converts a net.Conn to znet.Conn and then adds it into the Client.
func (cli *Client) Enroll(c net.Conn) (Conn, error) {
	return cli.EnrollContext(c, nil)
}

// EnrollContext is like Enroll but also accepts an empty interface ctx that can be obtained later via Conn.Context.
func (cli *Client) EnrollContext(c net.Conn, ctx any) (Conn, error) {
	defer c.Close() //nolint:errcheck

	sc, ok := c.(syscall.Conn)
	if !ok {
		return nil, errors.New("failed to convert net.Conn to syscall.Conn")
	}
	rc, err := sc.SyscallConn()
	if err != nil {
		return nil, errors.New("failed to get syscall.RawConn from net.Conn")
	}

	var dupFD int
	e := rc.Control(func(fd uintptr) {
		dupFD, err = unix.Dup(int(fd))
	})
	if err != nil {
		return nil, err
	}
	if e != nil {
		return nil, e
	}

	if cli.opts.SocketSendBuffer > 0 {
		if err = socket.SetSendBuffer(dupFD, cli.opts.SocketSendBuffer); err != nil {
			return nil, err
		}
	}
	if cli.opts.SocketRecvBuffer > 0 {
		if err = socket.SetRecvBuffer(dupFD, cli.opts.SocketRecvBuffer); err != nil {
			return nil, err
		}
	}

	el := cli.eng.eventLoops.next(nil)
	var (
		sockAddr unix.Sockaddr
		gc       *conn
	)
	switch c.(type) {
	case *net.UnixConn:
		sockAddr, _, _, err = socket.GetUnixSockAddr(c.RemoteAddr().Network(), c.RemoteAddr().String())
		if err != nil {
			return nil, err
		}
		ua := c.LocalAddr().(*net.UnixAddr)
		ua.Name = c.RemoteAddr().String() + "." + strconv.Itoa(dupFD)
		gc = newStreamConn("unix", dupFD, el, sockAddr, c.LocalAddr(), c.RemoteAddr())
	case *net.TCPConn:
		if cli.opts.TCPNoDelay == TCPNoDelay {
			if err = socket.SetNoDelay(dupFD, 1); err != nil {
				return nil, err
			}
		}
		if cli.opts.TCPKeepAlive > 0 {
			if err = setKeepAlive(
				dupFD,
				true,
				cli.opts.TCPKeepAlive,
				cli.opts.TCPKeepInterval,
				cli.opts.TCPKeepCount); err != nil {
				return nil, err
			}
		}
		sockAddr, _, _, _, err = socket.GetTCPSockAddr(c.RemoteAddr().Network(), c.RemoteAddr().String())
		if err != nil {
			return nil, err
		}
		gc = newStreamConn("tcp", dupFD, el, sockAddr, c.LocalAddr(), c.RemoteAddr())
	case *net.UDPConn:
		sockAddr, _, _, _, err = socket.GetUDPSockAddr(c.RemoteAddr().Network(), c.RemoteAddr().String())
		if err != nil {
			return nil, err
		}
		gc = newUDPConn(dupFD, el, c.LocalAddr(), sockAddr, true)
	default:
		return nil, zerror.ErrUnsupportedProtocol
	}
	gc.ctx = ctx

	connOpened := make(chan struct{})
	ccb := &connWithCallback{c: gc, cb: func() {
		close(connOpened)
	}}
	err = el.poller.Trigger(queue.HighPriority, el.register, ccb)
	if err != nil {
		gc.Close() //nolint:errcheck
		return nil, err
	}
	<-connOpened

	return gc, nil
}
