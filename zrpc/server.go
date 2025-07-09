package zrpc

import "github.com/meta-apex/zenith/znet"

type Server struct {
	znet.BuiltinEventEngine

	option *RpcServerConf

	seq uint64
}

func NewServer(c *RpcServerConf) *Server {
	s := &Server{option: c}

	return s
}
