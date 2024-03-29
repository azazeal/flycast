// Package config implements the configuration loader.
package config

import (
	"context"
	"math"
	"os"
	"strconv"

	"github.com/azazeal/exit"
	"github.com/azazeal/fly/env"
	"go.uber.org/zap"

	"github.com/azazeal/flycast/internal/common"
)

const (
	appKey        = "APP"
	globalPortKey = "PORT_GLOBAL"
	localPortKey  = "PORT_LOCAL"
	relayPortKey  = "PORT_RELAY"
	httpPortKey   = "PORT_HTTP"
)

// Config wraps the properties of the configuration.
type Config struct {
	// App holds the value of the APP environment variable.
	App string

	Ports struct {
		// Global holds the value of the PORT_GLOBAL environment variable.
		Global int

		// Local holds the value of the PORT_LOCAL environment variable.
		Local int

		// Relay holds the value of the PORT_RELAY environment variable.
		Relay int

		// HTTP holds the value of the PORT_HTTP environment variable.
		HTTP int
	}
}

// Fields the Config in the form of a slice of zap.Field.
func (cfg *Config) Fields() []zap.Field {
	return []zap.Field{
		zap.String("app", cfg.App),
		zap.Int("port.global", cfg.Ports.Global),
		zap.Int("port.local", cfg.Ports.Local),
		zap.Int("port.relay", cfg.Ports.Relay),
		zap.Int("port.http", cfg.Ports.HTTP),
	}
}

type contextKeyType struct{}

// FromContext returns the Config the given Context carries.
//
// FromContext panics in case the given Context carries no Config.
func FromContext(ctx context.Context) *Config {
	return ctx.Value(contextKeyType{}).(*Config)
}

// NewContext returns a copy of ctx which carries cfg.
func NewContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, contextKeyType{}, cfg)
}

const pkg = "config"

var (
	errNotOnFly = exit.Wrapf(common.ECNotOnFly,
		"%s/%s: not running on fly",
		common.AppName, pkg)

	errLoadConfig = exit.Wrapf(common.ECConfig,
		"%s/%s: failed loading configuration",
		common.AppName, pkg)
)

// Load loads the configuration from the environment.
func Load(logger *zap.Logger) (*Config, error) {
	logger = logger.Named(pkg)
	logger.Info("loading configuration ...")

	if !env.IsSet() {
		logger.Error("not running on fly.")

		return nil, errNotOnFly
	}

	var cfg Config
	var pGlobal, pLocal, pRelay, pHTTP string

	ok := []bool{
		fetch(&cfg.App, appKey, env.AppName()),

		fetch(&pGlobal, globalPortKey, "65535") &&
			setPort(logger, &cfg.Ports.Global, globalPortKey, pGlobal),

		fetch(&pLocal, localPortKey, "65534") &&
			setPort(logger, &cfg.Ports.Local, localPortKey, pLocal),

		fetch(&pRelay, relayPortKey, "65533") &&
			setPort(logger, &cfg.Ports.Relay, relayPortKey, pRelay),

		fetch(&pHTTP, httpPortKey, "8080") &&
			setPort(logger, &cfg.Ports.HTTP, httpPortKey, pHTTP),
	}

	for _, ok := range ok {
		if !ok {
			return nil, errLoadConfig
		}
	}

	return &cfg, nil
}

func fetch(into *string, key, defVal string) bool {
	v, found := os.LookupEnv(key)
	if !found {
		v = defVal
	}

	*into = v

	return true
}

func setPort(logger *zap.Logger, dst *int, key string, value string) (ok bool) {
	switch v, err := strconv.ParseUint(value, 10, 16); {
	case err != nil, v == 0:
		logger.Error("a port environment variable is invalid.",
			envVar(key),
			zap.Int("min", 1),
			zap.Int("max", math.MaxUint16))
	default:
		ok = true

		*dst = int(v)
	}

	return
}

func envVar(key string) zap.Field {
	return zap.String("var", "$"+key)
}
