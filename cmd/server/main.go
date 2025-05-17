package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mwdev22/TCP-bomberman/pkg/config"
	"github.com/mwdev22/TCP-bomberman/pkg/tcp"
)

func main() {

	cfg := config.NewServerConfig()

	server := tcp.NewServer(cfg.Addr, cfg.Port)
	if err := server.Listen(); err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	server.Stop()
}
