// Package wire implements the network level.
package wire

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/azazeal/health"

	"github.com/azazeal/flycast/internal/buffer"
	"github.com/azazeal/flycast/internal/config"
	"github.com/azazeal/flycast/internal/log"
	"github.com/azazeal/flycast/internal/loop"
	"github.com/azazeal/flycast/internal/peer"
	"github.com/azazeal/flycast/internal/region"
)

// Broadcast starts broadcasting to pl UDP messages it accepts on the UDP port
// for as long as ctx is not done.
//
// When ctx is done and the broadcasting has stopped, Done will be called on wg.
func Broadcast(ctx context.Context, wg *sync.WaitGroup, pl *peer.List, global bool) {
	var (
		logger = log.FromContext(ctx).
			Named("wire").
			Named(region.Alias(global))
		cfg = config.FromContext(ctx)
		hc  = health.FromContext(ctx)
		hcc = region.WireComponent(global)
	)

	buf := buffer.Get()
	defer buffer.Put(buf)

	go func() {
		defer wg.Done()

		loop.Func(ctx, time.Second, func(ctx context.Context) {
			defer hc.Fail(hcc)

			conn := bind(logger, bindPort(cfg, global))
			if conn == nil {
				return
			}
			hc.Pass(hcc)

			run(ctx, &broadcaster{
				logger: logger,
				conn:   conn,
				pl:     pl,
				buf:    buf,
			})
		})
	}()
}

func bind(logger *zap.Logger, port int) net.PacketConn {
	logger = logger.With(log.Port(port))
	logger.Info("binding ...")

	l, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Warn("failed binding.",
			zap.Error(err))

		return nil
	}
	logger.Debug("bound.")

	return l
}

func bindPort(cfg *config.Config, global bool) int {
	if global {
		return cfg.Ports.Global
	}

	return cfg.Ports.Local
}

type broadcaster struct {
	logger *zap.Logger
	conn   net.PacketConn
	pl     *peer.List
	buf    []byte
}

func run(ctx context.Context, b *broadcaster) {
	exited := make(chan struct{})
	defer close(exited)

	sd := func() { shutdown(b) }
	var shutDownOnce sync.Once
	defer shutDownOnce.Do(sd)

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			shutDownOnce.Do(sd)
		case <-exited:
			break
		}
	}()

	for {
		msg, err := b.read()
		if err != nil && !canRecoverFrom(err) {
			return // terminal error
		}
		if len(msg) == 0 {
			continue // nothing read
		}

		b.pl.Broadcast(b.conn, msg)
	}
}

func shutdown(b *broadcaster) {
	b.logger.Info("shutting down ...")

	if err := b.conn.Close(); err != nil {
		b.logger.Warn("failed shutting down.",
			zap.Error(err))

		return
	}

	b.logger.Debug("shut down.")
}

func (b *broadcaster) read() ([]byte, error) {
	logger := b.logger

	n, addr, err := b.conn.ReadFrom(b.buf[:buffer.Size:buffer.Size])
	if n > 0 {
		logger = logger.With(log.Data(b.buf[:n]))
	}
	if addr != nil {
		logger = logger.With(log.Addr(addr))
	}

	if err != nil {
		logger.Error("failed reading.",
			zap.Error(err))

		return nil, err
	}

	logger.Info("read.")

	return b.buf[:n], err
}

func canRecoverFrom(err error) bool {
	te, ok := err.(net.Error)

	return ok && (te.Temporary() || te.Timeout())
}
