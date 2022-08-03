package peer

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/azazeal/fly/dns"
	"github.com/azazeal/health"
	"go.uber.org/zap"

	"github.com/azazeal/flycast/internal/buffer"
	"github.com/azazeal/flycast/internal/config"
	"github.com/azazeal/flycast/internal/log"
	"github.com/azazeal/flycast/internal/loop"
	"github.com/azazeal/flycast/internal/region"
)

// Refresh returns a refreshing list of peers.
//
// global controls whether the returned List will be refreshed with global or
// local instances.
//
// After ctx is done and the list has stopped being refreshed, Done will be
// called on wg.
func Refresh(ctx context.Context, wg *sync.WaitGroup, global bool) *List {
	var (
		cfg = config.FromContext(ctx)

		lst = &List{
			logger: log.FromContext(ctx).
				Named("peer").
				Named(region.Alias(global)),
			hc:     health.FromContext(ctx),
			hcc:    region.PeerComponent(global),
			region: region.Name(global),
			app:    cfg.App,
			port:   cfg.Ports.Relay,
		}
	)

	go func() {
		defer wg.Done()

		loop.Func(ctx, time.Second, lst.refresh)
	}()

	return lst
}

// List is a set of peers.
type List struct {
	logger *zap.Logger
	hc     *health.Check
	hcc    string
	app    string
	port   int
	region string

	mu sync.Mutex
	ps peerSet
}

func (l *List) resolve(ctx context.Context) (ips []net.IP, ok bool) {
	switch ips, err := dns.Instances(ctx, l.app, l.region); {
	default:
		l.logger.Warn("failed resolving instances.",
			zap.Error(err))

		return nil, false
	case err == nil:
		return ips, true
	case isNXDomain(err):
		return nil, true
	}
}

func (l *List) refresh(ctx context.Context) {
	l.logger.Debug("resolving instances ...")

	at := time.Now()

	ips, ok := l.resolve(ctx)
	if !ok {
		l.hc.Fail(l.hcc)

		return
	}
	l.hc.Pass(l.hcc)

	var newSet peerSet
	if len(ips) > 0 {
		newSet = make(peerSet, len(ips))
	}

	for _, ip := range ips {
		newSet.add(ip, l.port)
	}

	// swap sets
	l.mu.Lock()
	oldSet := l.ps
	l.ps = newSet
	l.mu.Unlock()

	oldSet.empty()

	l.logger.Debug("resolved instances.",
		zap.Int("instances", len(ips)),
		zap.Duration("elapsed", time.Since(at)))
}

// Broadcast relays the message to all of the peers in l via conn.
func (l *List) Broadcast(conn net.PacketConn, msg []byte) {
	buf := buffer.Dup(msg)
	defer buffer.Put(buf)

	var wg sync.WaitGroup
	defer wg.Wait()

	l.mu.Lock()
	defer l.mu.Unlock()

	wg.Add(len(l.ps))
	for key := range l.ps {
		addr := l.ps[key]

		go func() {
			defer wg.Done()

			l.send(conn, addr, msg)
		}()
	}
}

func (l *List) send(conn net.PacketConn, to *net.UDPAddr, msg []byte) {
	logger := l.logger.
		With(log.IP(to.IP)).
		With(log.Port(to.Port)).
		With(log.Data(msg))
	logger.Debug("sending ...")

	n, err := conn.WriteTo(msg, to)
	if n > 0 {
		logger = logger.With(zap.Int("sent", n))
	}

	if err != nil {
		logger.Warn("failed sending.",
			zap.Error(err))

		return
	}

	logger.Debug("done sending.")
}

type peerSet map[string]*net.UDPAddr

func (ps peerSet) empty() {
	for k, a := range ps {
		a.IP = nil
		a.Port = 0
		a.Zone = ""

		addrPool.Put(a)

		delete(ps, k)
	}
}

func (ps peerSet) add(ip net.IP, port int) {
	addr := addrPool.Get().(*net.UDPAddr)

	addr.IP = ip
	addr.Port = port

	ps[string(ip)] = addr
}

var addrPool = sync.Pool{
	New: func() interface{} {
		return new(net.UDPAddr)
	},
}

func isNXDomain(err error) bool {
	e, ok := err.(*net.DNSError)
	return ok && e.IsNotFound
}
