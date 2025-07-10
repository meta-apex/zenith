package zrpc

import (
	"github.com/meta-apex/zenith/core/service"
	"github.com/meta-apex/zenith/core/stores/redis"
)

type (
	// A RpcServerConf is a rpc server config.
	RpcServerConf struct {
		service.ServiceConf
		ListenOn string
		Auth     bool            `meta:",optional"`
		Health   bool            `meta:",default=true"`
		Redis    redis.RedisConf `meta:",optional"`
		// setting 0 means no timeout
		Timeout      int64 `meta:",default=2000"`
		CpuThreshold int64 `meta:",default=900,range=[0:1000)"`
	}
)

func (sc RpcServerConf) Validate() error {
	return nil
}
