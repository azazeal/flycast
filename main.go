// flycast implements UDP broadcasting on fly.io.
package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/azazeal/exit"
	"github.com/azazeal/health"

	"github.com/azazeal/flycast/internal/app"
	"github.com/azazeal/flycast/internal/config"
	"github.com/azazeal/flycast/internal/log"
	"github.com/azazeal/flycast/internal/peer"
	"github.com/azazeal/flycast/internal/wire"
)

func main() {
	exit.With(run())
}

func run() (err error) {
	var ctx context.Context
	if ctx, err = newContext(context.TODO()); err != nil {
		return
	}

	var cancel context.CancelFunc
	ctx, cancel = signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	var wg sync.WaitGroup
	defer wg.Wait()

	// start the http server
	wg.Add(1)
	app.Serve(ctx, &wg)

	// start broadcasting globally
	wg.Add(2)
	global := peer.Refresh(ctx, &wg, true)
	wire.Broadcast(ctx, &wg, global, true)

	// start broadcasting locally
	wg.Add(2)
	local := peer.Refresh(ctx, &wg, false)
	wire.Broadcast(ctx, &wg, local, false)

	return
}

func newContext(parent context.Context) (ctx context.Context, err error) {
	logger := log.New("")

	var cfg *config.Config
	if cfg, err = config.Load(logger); err != nil {
		return
	}

	ctx = log.NewContext(parent, logger)
	ctx = config.NewContext(ctx, cfg)
	ctx = health.NewContext(ctx, new(health.Check))

	logger.Info("running.", cfg.Fields()...)

	return
}
