// Package app implements the application layer.
package app

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	golog "log"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/azazeal/fly/pkg/env"
	"github.com/azazeal/health"
	"github.com/azazeal/pause"
	"go.uber.org/zap"

	"github.com/azazeal/flycast/internal/common"
	"github.com/azazeal/flycast/internal/config"
	"github.com/azazeal/flycast/internal/log"
	"github.com/azazeal/flycast/internal/loop"
)

// Serve starts a goroutine which serves the app server until ctx is canceled.
//
// When the app server has stopped being ran, Done will called on wg.
func Serve(ctx context.Context, wg *sync.WaitGroup) {
	var (
		logger = log.FromContext(ctx).Named("app")
		cfg    = config.FromContext(ctx)
		addr   = fmt.Sprintf(":%d", cfg.Ports.HTTP)
		hc     = health.FromContext(ctx)
	)

	mux := newMux(ctx)

	go func() {
		defer wg.Done()

		loop.Func(ctx, time.Second, func(ctx context.Context) {
			defer hc.Unset(common.HCAppComponent)

			if l := bind(logger, addr); l != nil {
				hc.Set(common.HCAppComponent)

				serve(ctx, logger, l, mux)
			}

			pause.For(ctx, time.Second)
		})
	}()
}

func serve(parent context.Context, logger *zap.Logger, l net.Listener, h http.Handler) {
	var wg sync.WaitGroup
	defer wg.Wait()

	served := make(chan struct{}) // closed when Serve returns

	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: time.Second << 3,
		IdleTimeout:       time.Minute,
		MaxHeaderBytes:    4 << 10,
		ErrorLog:          newErrorLogger(logger),
		ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
			ctx = health.NewContext(ctx, health.FromContext(parent))
			ctx = config.NewContext(ctx, config.FromContext(parent))
			ctx = log.NewContext(ctx, log.FromContext(parent))

			return ctx
		},
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-parent.Done():
			ctx, cancel := context.WithTimeout(context.TODO(), time.Minute>>1)
			defer cancel()

			_ = srv.Shutdown(ctx)
		case <-served:
			break
		}
	}()

	_ = srv.Serve(l)
	close(served)
}

func bind(logger *zap.Logger, addr string) net.Listener {
	logger = logger.With(zap.String("addr", addr))
	logger.Info("binding ...")

	l, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed binding.")

		return nil
	}
	logger.Debug("bound.")

	return l
}

func newErrorLogger(logger *zap.Logger) *golog.Logger {
	lw := &logWrapper{
		logger.Named("http").Sugar(),
	}

	return golog.New(lw, "", 0)
}

type logWrapper struct {
	logger *zap.SugaredLogger
}

func (lw *logWrapper) Write(p []byte) (int, error) {
	lw.logger.Errorw(string(p))

	return len(p), nil
}

var (
	//go:embed index.gohtml
	indexBody string

	indexTemplate = template.Must(template.New("index.gohtml").Parse(indexBody))
)

type indexViewData struct {
	AppName  string
	Region   string
	Failures []string
}

func broadcast(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())

	var wrapper struct {
		Globally bool   `json:"globally"`
		Body     []byte `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&wrapper); err != nil {
		logger.Warn("failed decoding message.",
			zap.Error(err))

		respondWith(w, http.StatusUnprocessableEntity)

		return
	}

	logger = logger.
		With(log.Data(wrapper.Body)).
		With(zap.Bool("globally", wrapper.Globally))

	// TODO: implement

	respondWith(w, http.StatusNoContent)
}

func index(w http.ResponseWriter, r *http.Request) {
	hc := health.FromContext(r.Context())

	failures := hc.Failing(nil)
	sort.Strings(failures)

	indexTemplate.Execute(w, indexViewData{
		AppName:  common.AppName,
		Region:   env.Region(),
		Failures: failures,
	})
}

func newMux(ctx context.Context) (mux *http.ServeMux) {
	mux = http.NewServeMux()

	match := func(path string, h http.Handler, methods ...string) {
		mux.Handle(path, restrict(h, path, methods...))
	}

	matchFunc := func(path string, fn http.HandlerFunc, methods ...string) {
		match(path, fn, methods...)
	}

	hc := health.FromContext(ctx)
	match("/health", hc, http.MethodGet, http.MethodHead)

	matchFunc("/broadcast", broadcast, http.MethodPost)
	matchFunc("/", index, http.MethodGet)

	return
}

func matchFunc(m *http.ServeMux, fn http.HandlerFunc, path string, methods ...string) {
	m.Handle(path, restrict(fn, path, methods...))
}

func match(m *http.ServeMux, h http.Handler, path string, methods ...string) {
	m.Handle(path, restrict(h, path, methods...))
}

func restrict(h http.Handler, path string, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case !matchMethod(r.Method, methods...):
			respondWith(w, http.StatusMethodNotAllowed)
		case r.URL.Path != path:
			respondWith(w, http.StatusNotFound)
		default:
			h.ServeHTTP(w, r)
		}
	})
}

func matchMethod(method string, allowed ...string) bool {
	for _, a := range allowed {
		if method == a {
			return true
		}
	}

	return false
}

func respondWith(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
