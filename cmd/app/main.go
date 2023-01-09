package main

import (
	"log"

	"github.com/Shurubtsov/go-test-task-0/internal/config"
	controller "github.com/Shurubtsov/go-test-task-0/internal/controller/http"
	"github.com/Shurubtsov/go-test-task-0/internal/middleware"
)

func main() {
	cfg := config.GetConfig()
	router := controller.NewRouter()
	limiter := middleware.New(cfg)
	server := controller.NewServer(router, limiter)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
