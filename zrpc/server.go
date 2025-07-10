package zrpc

import (
	"github.com/meta-apex/zenith/zlog"
	"github.com/meta-apex/zenith/znet"
	"time"
)

type Server struct {
	znet.BuiltinEventEngine

	seq    uint64
	eng    znet.Engine
	config *RpcServerConf

	Accepted int64
	CurLoad  int64
	MaxLoad  int64
}

func NewServer(c *RpcServerConf) (*Server, error) {
	var err error
	if err = c.Validate(); err != nil {
		return nil, err
	}

	s := &Server{config: c}

	return s, nil
}

func (s *Server) Start() {
	if err := znet.Run(s, s.config.ListenOn); err != nil {
		zlog.Panic().Err(err).Msg("")
	}
}

func (s *Server) OnBoot(eng znet.Engine) (action znet.Action) {
	s.eng = eng
	return znet.None
}

func (s *Server) OnShutdown(eng znet.Engine) {

}

func (s *Server) OnOpen(c znet.Conn) (out []byte, action znet.Action) {
	return
}

func (s *Server) OnClose(c znet.Conn, err error) (action znet.Action) {
	return
}

func (s *Server) OnTraffic(c znet.Conn) (action znet.Action) {
	return
}

func (s *Server) OnTick() (delay time.Duration, action znet.Action) {
	return
}
