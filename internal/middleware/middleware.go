package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	boundDuration = 1 * time.Minute
	blockDuration = 2 * time.Minute
)

const requestLimit = 100

type Limiter struct {
	Mask net.IPMask
	IPs  map[string]*limitation
	Mu   sync.Mutex
}

type limitation struct {
	block        bool
	timerBound   *time.Timer
	timerBlocked *time.Timer
	requests     uint16
}

func New() *Limiter {
	mask := net.CIDRMask(24, 32)
	rl := make(map[string]*limitation, 10)
	return &Limiter{Mask: mask, IPs: rl}
}

func (l *Limiter) RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("X-Forwarded-For")
		ipv4Addr := net.ParseIP(header)
		subnet := ipv4Addr.Mask(l.Mask)

		l.Mu.Lock()
		defer l.Mu.Unlock()

		v, ok := l.IPs[subnet.String()]
		if !ok {
			limiter := &limitation{
				block:        false,
				timerBound:   time.NewTimer(boundDuration),
				timerBlocked: nil,
				requests:     1,
			}
			l.IPs[subnet.String()] = limiter

			go func() {
				for {
					if limiter.requests > requestLimit {
						if !limiter.timerBound.Stop() {
							log.Println("Timer is not fired when requests equal: ", limiter.requests)
							limiter.timerBound.Reset(time.Nanosecond)
							break
						}
					}
				}
			}()

			go func() {
				<-limiter.timerBound.C
				log.Println("Timer for limit was expired")
				if limiter.requests > requestLimit {
					log.Printf("Request limit was exceeded: Limit(%d) , Requests(%d)", requestLimit, limiter.requests)
					limiter.block = true
					limiter.timerBlocked = time.NewTimer(blockDuration)
					go func() {
						<-limiter.timerBlocked.C
						l.Mu.Lock()
						delete(l.IPs, subnet.String())
						l.Mu.Unlock()
					}()
					return
				}

				l.Mu.Lock()
				defer l.Mu.Unlock()
				delete(l.IPs, subnet.String())
			}()

		} else {
			if v.block {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("block requests for this IPs subnet"))
				return
			}
			v.requests++
		}

		next.ServeHTTP(w, r)
	}
}
