// Package log implements logging functionality.
package log

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/azazeal/flycast/internal/common"
)

type contextKeyType struct{}

// FromContext returns the Logger the given Context carries.
//
// FromContext panics in case the given Context carries no Logger.
func FromContext(ctx context.Context) *zap.Logger {
	return ctx.Value(contextKeyType{}).(*zap.Logger)
}

// NewContext returns a copy of ctx which carries logger.
func NewContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKeyType{}, logger)
}

const stampLayout = "2006-01-02 15:04:05.000"

var cfg = zap.Config{
	Encoding:         "console",
	Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
	EncoderConfig: zapcore.EncoderConfig{
		MessageKey:     "message",
		TimeKey:        "time",
		EncodeTime:     zapcore.TimeEncoderOfLayout(stampLayout),
		NameKey:        "logger",
		LevelKey:       "level",
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.StringDurationEncoder,
	},
}

var logger *zap.Logger // setup during init

func init() {
	if strings.ToLower(os.Getenv("LOG_FORMAT")) == "json" {
		// we're using structured logging
		cfg.Encoding = "json"
		cfg.EncoderConfig.TimeKey = zapcore.OmitKey
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	var level zapcore.Level
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	default:
		level = zapcore.InfoLevel
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(level)

	var err error
	if logger, err = cfg.Build(); err != nil {
		panic(err)
	}
}

// New returns a reference to a Logger with the given name.
func New(name string) *zap.Logger {
	if name == "" {
		name = common.AppName
	} else {
		name = fmt.Sprintf("%s.%s", common.AppName, name)
	}

	return logger.Named(name)
}

// Data is shorthand for zap.ByteString("data", v).
func Data(v []byte) zap.Field {
	return zap.ByteString("data", v)
}

// IP is shorthand for zap.String("ip", ip.String()).
func IP(ip net.IP) zap.Field {
	return zap.String("ip", ip.String())
}

// Port is shorthand for zap.Int("port", port).
func Port(port int) zap.Field {
	return zap.Int("port", port)
}

// Addr is shorthand for zap.String("addr", addr.String()).
func Addr(addr net.Addr) zap.Field {
	return zap.String("addr", addr.String())
}
