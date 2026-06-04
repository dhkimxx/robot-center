package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"robot-center/apps/server/internal/api"
	"robot-center/apps/server/internal/config"
)

// @title Robot Center 관제 API
// @version 0.1
// @description Robot Center 관제 서버 API입니다. Robot, Operator, Recorder, System 역할별 API를 제공합니다.
// @BasePath /
// @securityDefinitions.apikey RobotBearerAuth
// @in header
// @name Authorization
// @description Robot API에서 사용하는 Bearer robotToken입니다.
func main() {
	healthcheck := flag.Bool("healthcheck", false, "run an HTTP healthcheck against the local app-server")
	flag.Parse()

	cfg := config.LoadAppServerConfig()
	if *healthcheck {
		if err := runHealthcheck("http://127.0.0.1" + cfg.HTTPAddress + "/healthz"); err != nil {
			log.Fatal(err)
		}
		return
	}

	appServer, err := api.NewServerFromConfig(context.Background(), cfg)
	if err != nil {
		log.Fatalf("app-server init failed: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           appServer.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("app-server listening on %s", cfg.HTTPAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("app-server failed: %v", err)
		}
	}()

	waitForShutdown(server)
}

func runHealthcheck(url string) error {
	client := http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck failed with status %s", response.Status)
	}
	return nil
}

func waitForShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("app-server shutdown failed: %v", err)
	}
}
