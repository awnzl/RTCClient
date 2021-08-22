package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/awnzl/RTCClient/internal/worker"
	"github.com/gorilla/websocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn, err := connect("localhost:8080", "/ws")
	if err != nil {
		log.Fatal("connection dial:", err)
	}
	defer conn.Close()

	w := worker.New(conn, interrupt)
	w.Run()
}

func connect(addr, endpoint string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: addr, Path: endpoint}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	return conn, err
}
