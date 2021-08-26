package worker

import (
	"encoding/json"
	"errors"
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
	uiSender    chan string // todo add handling of ui buffers
	uiReceiver  chan string
}

func New(conn *websocket.Conn, interrupt chan os.Signal, sender, receiver chan string) *Worker {
	return &Worker{
		conn:       conn,
		done:       make(chan struct{}),
		interrupt:  interrupt,
		message:    make(chan domain.Message),
		uiSender:   sender,
		uiReceiver: receiver,
	}
}

func (w *Worker) Run() {
	go w.reader()
	go w.writer()
	go w.work()
}

func (w *Worker) reader() {
	defer close(w.done)

	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		w.uiReceiver <- string(message)
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
				log.Println("[writer] marshalling:", err)
				return // todo: or continue?
			}

			err = w.conn.WriteMessage(websocket.TextMessage, bts)
			if err != nil {
				log.Println("[writer] write:", err)
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

			w.message <- msg
		}
	}
}

func (w *Worker) getInputMessage() (domain.Message, error) {
	var msgType int
	text := <-w.uiSender

	switch text[0] {
	case 'j':
		msgType = domain.Join
		w.currentRoom = strings.TrimSpace(text[2:])
		text = ""

	case 'l':
		msgType = domain.Leave
		text = ""

	case 's':
		msgType = domain.Send
		text = strings.TrimSpace(text[2:])

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
