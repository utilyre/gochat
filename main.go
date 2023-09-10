package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	ErrUnsupportedMessageType = errors.New("message type not supported")
)

var (
	upgrader *websocket.Upgrader
	tmpl     *template.Template
)

func main() {
	var err error

	log.SetFlags(0)
	upgrader = &websocket.Upgrader{}
	tmpl, err = template.ParseGlob("views/*.html")
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()
	srv := http.Server{
		Addr:    ":3000",
		Handler: r,
	}

	r.Handle("/", http.FileServer(http.Dir("public")))
	r.HandleFunc("/chat", chat)

	log.Println("Listening on", srv.Addr)
	srv.ListenAndServe()
}

type Message struct {
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

func chat(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				return
			}

			log.Println(err)
			return
		}
		if mt != websocket.TextMessage {
			log.Println(ErrUnsupportedMessageType)
			return
		}

		msg := new(Message)
		if err := json.Unmarshal(data, msg); err != nil {
			log.Println(err)
			return
		}

		buf := new(bytes.Buffer)
		if err := tmpl.ExecuteTemplate(buf, "message", msg); err != nil {
			log.Println(err)
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}
}
