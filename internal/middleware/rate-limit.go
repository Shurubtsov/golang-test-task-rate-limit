package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Shurubtsov/go-test-task-0/internal/config"
)

const headerIP = "X-Forwarded-For"

type Limiter struct {
	Cfg  *config.Config
	Mask net.IPMask
	IPs  map[string]*limitation
	Mu   sync.Mutex
}

type limitation struct {
	block    bool
	blockCh  chan struct{}
	count    uint16
	countMax uint16
	ticker   *time.Ticker
}

func New(cfg *config.Config) *Limiter {
	mask := net.CIDRMask(24, 32)
	rl := make(map[string]*limitation, 10)
	return &Limiter{Mask: mask, IPs: rl, Cfg: cfg}
}

func (l *Limiter) RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		header := r.Header.Get(headerIP)
		ipv4Addr := net.ParseIP(header)
		subnet := ipv4Addr.Mask(l.Mask)

		l.Mu.Lock()
		defer l.Mu.Unlock()

		v, ok := l.IPs[subnet.String()]
		if !ok {
			limiter := &limitation{
				block:    false,
				blockCh:  make(chan struct{}),
				count:    l.Cfg.RequestsLimit,
				countMax: l.Cfg.RequestsLimit,
				ticker:   time.NewTicker(time.Duration(l.Cfg.BoundDuration) * time.Minute),
			}
			l.IPs[subnet.String()] = limiter

			go func() {
				for {
					if limiter.count <= 0 {
						limiter.blockCh <- struct{}{}
						return
					}

					select {
					case <-limiter.ticker.C:
						log.Println("tick fired")
						limiter.count = limiter.countMax
					default:
						continue
					}
				}
			}()

			go func(ch chan struct{}) {
				defer close(ch)
				<-ch

				log.Println("start block")

				blockTimer := time.NewTimer(time.Duration(l.Cfg.BlockDuration) * time.Minute)
				limiter.block = true

				<-blockTimer.C

				l.Mu.Lock()
				defer l.Mu.Unlock()
				delete(l.IPs, subnet.String())

			}(limiter.blockCh)

		} else {
			if v.block {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("block requests for this IPs subnet"))
				return
			}
			v.count--
		}

		next.ServeHTTP(w, r)
	}
}
