package controller

import (
	"net/http"

	"github.com/Shurubtsov/go-test-task-0/internal/middleware"
)

type server struct {
	router     *router
	middleware *middleware.Limiter
}

func NewServer(r *router, l *middleware.Limiter) *server {
	return &server{router: r, middleware: l}
}

func (s *server) Run() error {
	smux := &http.ServeMux{}
	smux.Handle("/test", s.middleware.RateLimit((http.HandlerFunc(s.router.Handler))))

	srv := http.Server{
		Addr:    ":8082",
		Handler: smux,
	}

	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
