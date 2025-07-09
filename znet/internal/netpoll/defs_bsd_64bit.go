//go:build (darwin || dragonfly || freebsd || netbsd || openbsd) && (amd64 || arm64 || ppc64 || ppc64le || mips64 || mips64le || riscv64)

package netpoll

type keventIdent = uint64
