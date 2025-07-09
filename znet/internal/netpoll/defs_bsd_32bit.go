//go:build (darwin || dragonfly || freebsd || netbsd || openbsd) && (386 || arm || mips || mipsle)

package netpoll

type keventIdent = uint32
