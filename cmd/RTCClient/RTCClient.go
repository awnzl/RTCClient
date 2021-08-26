package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"

	"github.com/awnzl/RTCClient/internal/ui"
	"github.com/awnzl/RTCClient/internal/worker"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn, err := connect("localhost:8080", "/ws")
	if err != nil {
		log.Fatal("connection dial:", err)
	}
	defer conn.Close()

	g, err := ui.NewGUI()
	if err != nil {
		log.Fatal("GUI creation:", err)
	}
	defer g.Close()

	if err := g.Init(); err != nil {
		log.Fatal("GUI initialization", err)
	}

	w := worker.New(conn, interrupt, g.Sender, g.Receiver)
	w.Run()

	if err := g.MainLoop(); err != nil {
		log.Fatal("GUI MainLoop", err)
	}
}

func connect(addr, endpoint string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: addr, Path: endpoint}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	return conn, err
}
