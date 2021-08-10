// Package common implements functionality consumed by other packages.
package common

import (
	"io"
	"sync"
)

// AppName denotes the app's name.
const AppName = "flycast"

// The set of exit codes this application uses.
const (
	_ = iota + 2

	// ECSetEnvDefaults is the exit code the application uses used when it
	// cannot set default values on its environment.
	ECSetEnvDefaults

	// ECNotOnFly is the exit code the application uses when it cannot detect
	// a properly set fly environment.
	ECNotOnFly

	// ECConfig is the exit code the application uses when it's unable to
	// load its configuration from the environment or the configuration is
	// invalid.
	ECConfig
)

// health check names
const (
	HCRefreshGlobalComponent = "refresh.global"
	HCRefreshLocalComponent  = "refresh.local"
	HCAppComponent           = "app"
	HCWireGlobal             = "wire.global"
	HCWireLocal              = "wire.local"
)

// CloseOnce wraps closer with a sync.Once so that it may only be closed once.
func CloseOnce(closer io.Closer) io.Closer {
	return &closeOnce{
		Closer: closer,
	}
}

type closeOnce struct {
	sync.Once
	io.Closer
	err error
}

func (co *closeOnce) Close() error {
	co.Once.Do(func() {
		co.err = co.Closer.Close()
	})

	return co.err
}
