package zproc

import (
	"fmt"
	"github.com/meta-apex/zenith/zlog"
	"os"
	"path"
	"runtime/pprof"
	"syscall"
	"time"
)

const (
	goroutineProfile = "goroutine"
	debugLevel       = 2
)

type creator interface {
	Create(name string) (file *os.File, err error)
}

func dumpGoroutines(ctor creator) {
	command := path.Base(os.Args[0])
	pid := syscall.Getpid()
	dumpFile := path.Join(os.TempDir(), fmt.Sprintf("%s-%d-goroutines-%s.dump",
		command, pid, time.Now().Format(timeFormat)))

	zlog.Info().Msgf("Got dump goroutine signal, printing goroutine profile to %s", dumpFile)

	if f, err := ctor.Create(dumpFile); err != nil {
		zlog.Error().Msgf("Failed to dump goroutine profile, error: %v", err)
	} else {
		defer f.Close()
		_ = pprof.Lookup(goroutineProfile).WriteTo(f, debugLevel)
	}
}

type fileCreator struct{}

func (fc fileCreator) Create(name string) (file *os.File, err error) {
	return os.Create(name)
}
