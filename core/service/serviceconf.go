package service

import (
	"github.com/meta-apex/zenith/core/load"
	"github.com/meta-apex/zenith/core/prometheus"
	"github.com/meta-apex/zenith/core/stat"
	"github.com/meta-apex/zenith/core/trace"
	"github.com/meta-apex/zenith/core/zproc"
	"github.com/meta-apex/zenith/internal/devserver"
	"github.com/meta-apex/zenith/internal/profiling"
)

const (
	// DevMode means development mode.
	DevMode = "dev"
	// TestMode means test mode.
	TestMode = "test"
	// RtMode means regression test mode.
	RtMode = "rt"
	// PreMode means pre-release mode.
	PreMode = "pre"
	// ProMode means production mode.
	ProMode = "pro"
)

type (
	// DevServerConfig is type alias for devserver.Config
	DevServerConfig = devserver.Config

	// A ServiceConf is a service config.
	ServiceConf struct {
		Name       string
		Mode       string `json:",default=pro,options=dev|test|rt|pre|pro"`
		MetricsUrl string `json:",optional"`
		// Deprecated: please use DevServer
		Prometheus prometheus.Config  `json:",optional"`
		Telemetry  trace.Config       `json:",optional"`
		DevServer  DevServerConfig    `json:",optional"`
		Shutdown   zproc.ShutdownConf `json:",optional"`
		// Profiling is the configuration for continuous profiling.
		Profiling profiling.Config `json:",optional"`
	}
)

// MustSetUp sets up the service, exits on error.
func (sc ServiceConf) MustSetUp() {
	
}

// SetUp sets up the service.
func (sc ServiceConf) SetUp() error {

	sc.initMode()
	prometheus.StartAgent(sc.Prometheus)

	if len(sc.Telemetry.Name) == 0 {
		sc.Telemetry.Name = sc.Name
	}
	trace.StartAgent(sc.Telemetry)
	zproc.Setup(sc.Shutdown)
	zproc.AddShutdownListener(func() {
		trace.StopAgent()
	})

	if len(sc.MetricsUrl) > 0 {
		stat.SetReportWriter(stat.NewRemoteWriter(sc.MetricsUrl))
	}

	devserver.StartAgent(sc.DevServer)
	profiling.Start(sc.Profiling)

	return nil
}

func (sc ServiceConf) initMode() {
	switch sc.Mode {
	case DevMode, TestMode, RtMode, PreMode:
		load.Disable()
		stat.SetReporter(nil)
	}
}
