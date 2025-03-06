package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Drinnn/consume-it/internal/server"
	"github.com/Drinnn/consume-it/internal/server/clients"
)

var (
	port = flag.Int("port", 8080, "Port to listen on")
)

func main() {
	flag.Parse()

	hub := server.NewHub()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewWebSocketClient, w, r)
	})

	go hub.Run()

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Listening on %s", addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
