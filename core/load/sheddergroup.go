package load

import (
	"github.com/meta-apex/zenith/core/zsync"
	"io"
)

// A ShedderGroup is a manager to manage key-based shedders.
type ShedderGroup struct {
	options []ShedderOption
	manager *zsync.ResourceManager
}

// NewShedderGroup returns a ShedderGroup.
func NewShedderGroup(opts ...ShedderOption) *ShedderGroup {
	return &ShedderGroup{
		options: opts,
		manager: zsync.NewResourceManager(),
	}
}

// GetShedder gets the Shedder for the given key.
func (g *ShedderGroup) GetShedder(key string) Shedder {
	shedder, _ := g.manager.GetResource(key, func() (closer io.Closer, e error) {
		return nopCloser{
			Shedder: NewAdaptiveShedder(g.options...),
		}, nil
	})
	return shedder.(Shedder)
}

type nopCloser struct {
	Shedder
}

func (c nopCloser) Close() error {
	return nil
}
