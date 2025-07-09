package rescue

import (
	"context"
	"fmt"
	"github.com/meta-apex/zenith/zlog"
	"runtime/debug"
)

// Recover is used with defer to do cleanup on panics.
// Use it like:
//
//	defer Recover(func() {})
func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		zlog.Error().Stack().Msgf(fmt.Sprint(p))
	}
}

// RecoverCtx is used with defer to do cleanup on panics.
func RecoverCtx(ctx context.Context, cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		zlog.Error().Msgf("%+v\n%s", p, debug.Stack())
	}
}
