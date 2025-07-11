package znet

import (
	"context"
	"net"
)

// contextKey is a key for Conn values in context.Context.
type contextKey struct{}

// NewContext returns a new context.Context that carries the value
// that will be attached to the Conn.
func NewContext(ctx context.Context, v any) context.Context {
	return context.WithValue(ctx, contextKey{}, v)
}

// FromContext retrieves context value of the Conn stored in ctx, if any.
func FromContext(ctx context.Context) any {
	return ctx.Value(contextKey{})
}

// connContextKey is a key for net.Conn values in context.Context.
type connContextKey struct{}

// NewNetConnContext returns a new context.Context that carries the net.Conn value.
func NewNetConnContext(ctx context.Context, c net.Conn) context.Context {
	return context.WithValue(ctx, connContextKey{}, c)
}

// FromNetConnContext retrieves the net.Conn value from ctx, if any.
func FromNetConnContext(ctx context.Context) (net.Conn, bool) {
	c, ok := ctx.Value(connContextKey{}).(net.Conn)
	return c, ok
}

// netAddrContextKey is a key for net.Addr values in context.Context.
type netAddrContextKey struct{}

// NewNetAddrContext returns a new context.Context that carries the net.Addr value.
func NewNetAddrContext(ctx context.Context, a net.Addr) context.Context {
	return context.WithValue(ctx, netAddrContextKey{}, a)
}

// FromNetAddrContext retrieves the net.Addr value from ctx, if any.
func FromNetAddrContext(ctx context.Context) (net.Addr, bool) {
	a, ok := ctx.Value(netAddrContextKey{}).(net.Addr)
	return a, ok
}
