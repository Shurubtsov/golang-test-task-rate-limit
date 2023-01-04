package controller

import (
	"net/http"
)

type router struct {
}

func NewRouter() *router {
	return &router{}
}

func (rt *router) Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
