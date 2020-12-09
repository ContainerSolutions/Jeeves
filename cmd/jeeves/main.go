package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ContainerSolutions/jeeves/pkg/helpers"
	"github.com/ContainerSolutions/jeeves/pkg/router"

	log "github.com/sirupsen/logrus"
)

const startUpLog = `
       __   _______  ___________    ____  _______     _______.
      |  | |   ____||   ____\   \  /   / |   ____|   /       |
      |  | |  |__   |  |__   \   \/   /  |  |__     |   (----
.--.  |  | |   __|  |   __|   \      /   |   __|     \   \
|   --'  | |  |____ |  |____   \    /    |  |____.----)   |
 \______/  |_______||_______|   \__/     |_______|_______/ 
`

func main() {
	// Termination Handeling
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	router := router.GetRouter()
	// Set server configuration
	port := helpers.GetEnv("SERVER_PORT", "8000")
	host := helpers.GetEnv("SERVER_HOST", "127.0.0.1")
	// String formatting to join the host and port
	addr := fmt.Sprintf("%v:%v", host, port)
	// Setup Server
	srv := &http.Server{
		Handler:      router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Run Server in Goroutine to handle Graceful Shutdowns
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatal("Server Start Fail")
		}
	}()
	// Prints out ascii art
	fmt.Println(startUpLog)
	log.WithFields(log.Fields{
		"host": host,
		"port": port,
	}).Info("Starting Server")
	<-termChan
	// Any Code to Gracefully Shutdown should be done here
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Graceful Shutdown Failed")
	}
	log.Info("Shutting Down Gracefully")
}
