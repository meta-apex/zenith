package zerror

import "errors"

var (
	// ErrEmptyEngine occurs when trying to do something with an empty engine.
	ErrEmptyEngine = errors.New("znet: the internal engine is empty")
	// ErrEngineShutdown occurs when server is closing.
	ErrEngineShutdown = errors.New("znet: server is going to be shutdown")
	// ErrEngineInShutdown occurs when attempting to shut the server down more than once.
	ErrEngineInShutdown = errors.New("znet: server is already in shutdown")
	// ErrAcceptSocket occurs when acceptor does not accept the new connection properly.
	ErrAcceptSocket = errors.New("znet: accept a new connection error")
	// ErrTooManyEventLoopThreads occurs when attempting to set up more than 10,000 event-loop goroutines under LockOSThread mode.
	ErrTooManyEventLoopThreads = errors.New("znet: too many event-loops under LockOSThread mode")
	// ErrUnsupportedProtocol occurs when trying to use protocol that is not supported.
	ErrUnsupportedProtocol = errors.New("znet: only unix, tcp/tcp4/tcp6, udp/udp4/udp6 are supported")
	// ErrUnsupportedTCPProtocol occurs when trying to use an unsupported TCP protocol.
	ErrUnsupportedTCPProtocol = errors.New("znet: only tcp/tcp4/tcp6 are supported")
	// ErrUnsupportedUDPProtocol occurs when trying to use an unsupported UDP protocol.
	ErrUnsupportedUDPProtocol = errors.New("znet: only udp/udp4/udp6 are supported")
	// ErrUnsupportedUDSProtocol occurs when trying to use an unsupported Unix protocol.
	ErrUnsupportedUDSProtocol = errors.New("znet: only unix is supported")
	// ErrUnsupportedOp occurs when calling some methods that are either not supported or have not been implemented yet.
	ErrUnsupportedOp = errors.New("znet: unsupported operation")
	// ErrNegativeSize occurs when trying to pass a negative size to a buffer.
	ErrNegativeSize = errors.New("znet: negative size is not allowed")
	// ErrNoIPv4AddressOnInterface occurs when an IPv4 multicast address is set on an interface but IPv4 is not configured.
	ErrNoIPv4AddressOnInterface = errors.New("znet: no IPv4 address on interface")
	// ErrInvalidNetworkAddress occurs when the network address is invalid.
	ErrInvalidNetworkAddress = errors.New("znet: invalid network address")
	// ErrInvalidNetConn occurs when trying to do something with an empty net.Conn.
	ErrInvalidNetConn = errors.New("znet: the net.Conn is empty")
	// ErrNilRunnable occurs when trying to execute a nil runnable.
	ErrNilRunnable = errors.New("znet: nil runnable is not allowed")
)
