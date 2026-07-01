package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/scheduler"
	"github.com/switchboard/switchboard/internal/server"

	"github.com/hibiken/asynq"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	srv, err := server.New(ctx, cfg)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}
	defer srv.Close()

	redisOpt, _ := asynq.ParseRedisURI(cfg.RedisURL)
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()
	_ = scheduler.Start(cfg, asynqClient)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      srv.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
