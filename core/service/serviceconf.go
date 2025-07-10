package service

import (
	"github.com/meta-apex/zenith/core/load"
	"github.com/meta-apex/zenith/core/stat"
	"github.com/meta-apex/zenith/core/trace"
	"github.com/meta-apex/zenith/core/zproc"
	"github.com/meta-apex/zenith/internal/devserver"
)

const (
	// DevMode means development mode.
	DevMode = "dev"
	// TestMode means test mode.
	TestMode = "test"
	// PreMode means pre-release mode.
	PreMode = "pre"
	// ProMode means production mode.
	ProMode = "pro"
)

type (
	DevServerConfig = devserver.Config

	// A ServiceConf is a service config.
	ServiceConf struct {
		Name       string
		Mode       string             `meta:",default=pro,options=dev|test|pre|pro"`
		MetricsUrl string             `meta:",optional"`
		Telemetry  trace.Config       `meta:",optional"`
		DevServer  DevServerConfig    `meta:",optional"`
		Shutdown   zproc.ShutdownConf `meta:",optional"`
	}
)

// MustSetUp sets up the service, exits on error.
func (sc ServiceConf) MustSetUp() {

}

// SetUp sets up the service.
func (sc ServiceConf) SetUp() error {
	sc.initMode()

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

	return nil
}

func (sc ServiceConf) initMode() {
	switch sc.Mode {
	case DevMode, TestMode, PreMode:
		load.Disable()
		stat.SetReporter(nil)
	}
}
