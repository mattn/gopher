package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/mattn/gopher"
	"golang.org/x/net/websocket"
)

var (
	conns = make(map[*websocket.Conn]string)
	id    = 0
	mutex sync.Mutex
)

type msg struct {
	Type  string `json:"type"`
	User  string `json:"user"`
	Value string `json:"value"`
}

func nextId() int {
	mutex.Lock()
	defer mutex.Unlock()
	id++
	return id
}

func server(ws *websocket.Conn) {
	defer ws.Close()

	user := fmt.Sprintf("user%03d", nextId())
	b, err := json.Marshal(&msg{Type: "whoami", User: user})
	if err != nil {
		log.Println("login failed:", err.Error())
		return
	}
	if err := websocket.Message.Send(ws, string(b)); err != nil {
		log.Println("login failed:", err.Error())
		return
	}
	b, err = json.Marshal(&msg{Type: "join", User: user})
	if err != nil {
		log.Println("login failed:", err.Error())
		return
	}
	for conn, _ := range conns {
		if err := websocket.Message.Send(conn, string(b)); err != nil {
			log.Println("send failed:", err)
		}
	}
	conns[ws] = user

	if ws.Request().URL.Query().Get("mode") == "" {
		gopher.Create(user)
	}

	for {
		var b []byte
		if err := websocket.Message.Receive(ws, &b); err != nil {
			if err != io.EOF {
				log.Println("receive failed:", err.Error())
			}
			delete(conns, ws)
			for _, t := range gopher.LookupByName(user) {
				t.Terminate()
			}
			return
		}

		var m msg
		err := json.Unmarshal(b, &m)
		if err != nil {
			log.Println("send failed:", err.Error())
			continue
		}
		m.User = user

		b, err = json.Marshal(&m)
		if err != nil {
			log.Println("send failed:", err.Error())
			continue
		}

		for conn, _ := range conns {
			if err := websocket.Message.Send(conn, string(b)); err != nil {
				log.Println("send failed:", err)
			}
		}

		if m.Type == "message" {
			for _, u := range gopher.LookupByName(user) {
				u.Message(user+": "+m.Value, "")
			}
		}
	}
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		for _, t := range gopher.Lookup() {
			t.Terminate()
		}
		os.Exit(0)
	}()

	http.Handle("/", http.FileServer(http.Dir("assets")))
	http.Handle("/chat", websocket.Handler(server))
	log.Println("started at :8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
