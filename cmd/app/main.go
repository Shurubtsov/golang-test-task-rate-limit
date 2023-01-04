package main

import (
	"log"

	controller "github.com/Shurubtsov/go-test-task-0/internal/controller/http"
	"github.com/Shurubtsov/go-test-task-0/internal/middleware"
)

func main() {

	router := controller.NewRouter()
	limiter := middleware.New()
	server := controller.NewServer(router, limiter)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
