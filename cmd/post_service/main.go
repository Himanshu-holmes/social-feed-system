package main

import (

	"log"
	"os"
	"os/signal"


	"github.com/himanshu-holmes/social-feed-system/cmd/post_service/server"
)

func main() {

	srv := server.NewServer()
	srv.Run()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	log.Println("Received shutdown signal, shutting down gracefully...")

}