package profiling

import (
	"github.com/grafana/pyroscope-go"
	"github.com/meta-apex/zenith/core/stat"
	"github.com/meta-apex/zenith/core/threading"
	"github.com/meta-apex/zenith/core/zproc"
	"github.com/meta-apex/zenith/zlog"
	"runtime"
	"sync"
	"time"
)

const (
	defaultCheckInterval     = time.Second * 10
	defaultProfilingDuration = time.Minute * 2
	defaultUploadRate        = time.Second * 15
)

type (
	Config struct {
		// Name is the name of the application.
		Name string `meta:",optional,inherit"`
		// ServerAddr is the address of the profiling server.
		ServerAddr string
		// AuthUser is the username for basic authentication.
		AuthUser string `meta:",optional"`
		// AuthPassword is the password for basic authentication.
		AuthPassword string `meta:",optional"`
		// UploadRate is the duration for which profiling data is uploaded.
		UploadRate time.Duration `meta:",default=15s"`
		// CheckInterval is the interval to check if profiling should start.
		CheckInterval time.Duration `meta:",default=10s"`
		// ProfilingDuration is the duration for which profiling data is collected.
		ProfilingDuration time.Duration `meta:",default=2m"`
		// CpuThreshold the collection is allowed only when the current service cpu > CpuThreshold
		CpuThreshold int64 `meta:",default=700,range=[0:1000)"`

		// ProfileType is the type of profiling to be performed.
		ProfileType ProfileType
	}

	ProfileType struct {
		// Logger is a flag to enable or disable logging.
		Logger bool `meta:",default=false"`
		// CPU is a flag to disable CPU profiling.
		CPU bool `meta:",default=true"`
		// Goroutines is a flag to disable goroutine profiling.
		Goroutines bool `meta:",default=true"`
		// Memory is a flag to disable memory profiling.
		Memory bool `meta:",default=true"`
		// Mutex is a flag to disable mutex profiling.
		Mutex bool `meta:",default=false"`
		// Block is a flag to disable block profiling.
		Block bool `meta:",default=false"`
	}

	profiler interface {
		Start() error
		Stop() error
	}

	pyroscopeProfiler struct {
		c        Config
		profiler *pyroscope.Profiler
	}
)

var (
	once sync.Once

	newProfiler = func(c Config) profiler {
		return newPyroscopeProfiler(c)
	}
)

// Start initializes the pyroscope profiler with the given configuration.
func Start(c Config) {
	// check if the profiling is enabled
	if len(c.ServerAddr) == 0 {
		return
	}

	// set default values for the configuration
	if c.ProfilingDuration <= 0 {
		c.ProfilingDuration = defaultProfilingDuration
	}

	// set default values for the configuration
	if c.CheckInterval <= 0 {
		c.CheckInterval = defaultCheckInterval
	}

	if c.UploadRate <= 0 {
		c.UploadRate = defaultUploadRate
	}

	once.Do(func() {
		zlog.Info().Msg("continuous profiling started")

		threading.GoSafe(func() {
			startPyroscope(c, zproc.Done())
		})
	})
}

// startPyroscope starts the pyroscope profiler with the given configuration.
func startPyroscope(c Config, done <-chan struct{}) {
	var (
		pr                  profiler
		err                 error
		latestProfilingTime time.Time
		intervalTicker      = time.NewTicker(c.CheckInterval)
		profilingTicker     = time.NewTicker(c.ProfilingDuration)
	)

	defer profilingTicker.Stop()
	defer intervalTicker.Stop()

	for {
		select {
		case <-intervalTicker.C:
			// Check if the machine is overloaded and if the profiler is not running
			if pr == nil && isCpuOverloaded(c) {
				pr = newProfiler(c)
				if err := pr.Start(); err != nil {
					zlog.Error().Msgf("failed to start profiler: %v", err)
					continue
				}

				// record the latest profiling time
				latestProfilingTime = time.Now()
				zlog.Info().Msgf("pyroscope profiler started.")
			}
		case <-profilingTicker.C:
			// check if the profiling duration has passed
			if !time.Now().After(latestProfilingTime.Add(c.ProfilingDuration)) {
				continue
			}

			// check if the profiler is already running, if so, skip
			if pr != nil {
				if err = pr.Stop(); err != nil {
					zlog.Error().Msgf("failed to stop profiler: %v", err)
				}
				zlog.Info().Msgf("pyroscope profiler stopped.")
				pr = nil
			}
		case <-done:
			zlog.Info().Msgf("continuous profiling stopped.")
			return
		}
	}
}

// genPyroscopeConf generates the pyroscope configuration based on the given config.
func genPyroscopeConf(c Config) pyroscope.Config {
	pConf := pyroscope.Config{
		UploadRate:        c.UploadRate,
		ApplicationName:   c.Name,
		BasicAuthUser:     c.AuthUser,     // http basic auth user
		BasicAuthPassword: c.AuthPassword, // http basic auth password
		ServerAddress:     c.ServerAddr,
		Logger:            nil,
		HTTPHeaders:       map[string]string{},
		// you can provide static tags via a map:
		Tags: map[string]string{
			"name": c.Name,
		},
	}

	if c.ProfileType.CPU {
		pConf.ProfileTypes = append(pConf.ProfileTypes, pyroscope.ProfileCPU)
	}
	if c.ProfileType.Goroutines {
		pConf.ProfileTypes = append(pConf.ProfileTypes, pyroscope.ProfileGoroutines)
	}
	if c.ProfileType.Memory {
		pConf.ProfileTypes = append(pConf.ProfileTypes, pyroscope.ProfileAllocObjects, pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects, pyroscope.ProfileInuseSpace)
	}
	if c.ProfileType.Mutex {
		pConf.ProfileTypes = append(pConf.ProfileTypes, pyroscope.ProfileMutexCount, pyroscope.ProfileMutexDuration)
	}
	if c.ProfileType.Block {
		pConf.ProfileTypes = append(pConf.ProfileTypes, pyroscope.ProfileBlockCount, pyroscope.ProfileBlockDuration)
	}

	zlog.Info().Msgf("applicationName: %s", pConf.ApplicationName)

	return pConf
}

// isCpuOverloaded checks the machine performance based on the given configuration.
func isCpuOverloaded(c Config) bool {
	currentValue := stat.CpuUsage()
	if currentValue >= c.CpuThreshold {
		zlog.Info().Msgf("continuous profiling cpu overload, cpu: %d", currentValue)
		return true
	}

	return false
}

func newPyroscopeProfiler(c Config) profiler {
	return &pyroscopeProfiler{
		c: c,
	}
}

func (p *pyroscopeProfiler) Start() error {
	pConf := genPyroscopeConf(p.c)
	// set mutex and block profile rate
	setFraction(p.c)
	prof, err := pyroscope.Start(pConf)
	if err != nil {
		resetFraction(p.c)
		return err
	}

	p.profiler = prof
	return nil
}

func (p *pyroscopeProfiler) Stop() error {
	if p.profiler == nil {
		return nil
	}

	if err := p.profiler.Stop(); err != nil {
		return err
	}

	resetFraction(p.c)
	p.profiler = nil

	return nil
}

func setFraction(c Config) {
	// These 2 lines are only required if you're using mutex or block profiling
	if c.ProfileType.Mutex {
		runtime.SetMutexProfileFraction(10) // 10/seconds
	}
	if c.ProfileType.Block {
		runtime.SetBlockProfileRate(1000 * 1000) //  1/millisecond
	}
}

func resetFraction(c Config) {
	// These 2 lines are only required if you're using mutex or block profiling
	if c.ProfileType.Mutex {
		runtime.SetMutexProfileFraction(0)
	}
	if c.ProfileType.Block {
		runtime.SetBlockProfileRate(0)
	}
}
