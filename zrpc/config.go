package zrpc

type (
	// A RpcServerConf is a rpc server config.
	RpcServerConf struct {
		Name     string
		ListenOn string
		Auth     bool `meta:",optional"`
		// setting 0 means no timeout
		Timeout      int64 `meta:",default=2000"`
		CpuThreshold int64 `meta:",default=900,range=[0:1000)"`
		// zrpc health check switch
		Health bool `meta:",default=true"`
	}
)
