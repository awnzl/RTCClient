package worker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/awnzl/RTCClient/internal/domain"
)

type Worker struct {
	conn        *websocket.Conn
	currentRoom string
	done        chan struct{}
	interrupt   chan os.Signal
	message     chan domain.Message
}

func New(conn *websocket.Conn, interrupt chan os.Signal) *Worker {
	return &Worker{
		conn:      conn,
		done:      make(chan struct{}),
		interrupt: interrupt,
		message:   make(chan domain.Message),
	}
}

func (w *Worker) Run() {
	go w.reader()
	go w.writer()
	w.work()
}

func (w *Worker) reader() {
	defer close(w.done)

	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		log.Printf("recv: %s", message)
	}
}

func (w *Worker) writer() {
	for {
		select {
		case <-w.done:
			return

		case msg := <-w.message:
			bts, err := json.Marshal(msg)
			if err != nil {
				log.Println("writer marshalling:", err)
				return // todo: or continue?
			}

			err = w.conn.WriteMessage(websocket.TextMessage, bts)
			if err != nil {
				log.Println("writer write:", err)
				return // todo: or continue?
			}

		case <-w.interrupt:
			defer close(w.interrupt)
			log.Println("interrupt")

			// send a close message and then wait (with timeout) for
			// the server to close the connection.
			err := w.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-w.done:
			case <-time.After(time.Second):
			}

			return
		}
	}
}

func (w *Worker) work() {
	for {
		select {
		case <-w.done:
			return

		case <-w.interrupt:
			return

		default:
			msg, err := w.getInputMessage() // todo: this blocks stream while the next Input won't happen
			if err != nil {
				log.Println("[work] get input message:", err)
				continue
			}

			bts, err := json.Marshal(msg)
			if err != nil {
				log.Println("[work] marshalling:", err)
				continue
			}

			err = w.conn.WriteMessage(websocket.TextMessage, bts)
			if err != nil {
				log.Println("[work] write:", err)
				continue
			}
		}
	}
}

func (w *Worker) getInputMessage() (domain.Message, error) {
	text, err := readInput()
	if err != nil {
		return domain.Message{}, err
	}

	var msgType int

	switch text[0] {
	case 'j':
		log.Println("join input")
		msgType = domain.Join
		w.currentRoom = strings.TrimSpace(text[2:])
		text = ""

	case 'l':
		log.Println("leave input")
		msgType = domain.Leave

	case 's':
		log.Println("send input")
		text = strings.TrimSpace(text[2:])
		msgType = domain.Send

	default:
		return domain.Message{}, errors.New("incorrect message input scheme")
	}

	msg := domain.Message{
		Content: text,
		Room:    w.currentRoom,
		Time:    time.Now(),
		Type:    msgType,
	}

	return msg, nil
}

func readInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter text: ")
	text, err := reader.ReadString('\n')

	if err != nil {
		return "", err
	}

	return text, nil
}
