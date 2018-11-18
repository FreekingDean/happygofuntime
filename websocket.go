package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{}

func proxy(w http.ResponseWriter, r *http.Request) {
	scheme := r.Header.Get("X-Forward-Scheme")
	log.Println(scheme)
	r.Header["Connection"] = []string{"upgrade"}
	r.Header["Upgrade"] = []string{"websocket"}
	client, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()

	normalHost := r.Header.Get("X-Forward-Host")
	r.Header.Del("X-Forward-Host")
	r.Header.Del("Connection")
	r.Header.Del("Upgrade")
	r.Header.Del("Sec-Websocket-Key")
	r.Header.Del("Sec-Websocket-Version")
	url := r.URL
	url.Host = normalHost
	url.Scheme = "wss"
	server, _, err := websocket.DefaultDialer.Dial(url.String(), r.Header)
	if err != nil {
		log.Println(err)
		return
	}
	defer server.Close()

	go func() {
		for {
			mt, message, err := client.ReadMessage()
			if err != nil {
				log.Println("CLIENT ERROR: ", err)
				return
			}
			log.Printf("CLIENT GET: %s\n", message)
			err = server.WriteMessage(mt, message)
			if err != nil {
				log.Println("CLIENT ERROR: ", err)
				return
			}
		}
	}()
	for {
		mt, message, err := server.ReadMessage()
		if err != nil {
			log.Println("SERVER ERROR: ", err)
			return
		}
		log.Printf("SERVER GET: %s\n", message)
		d := Message{}
		err = json.Unmarshal([]byte(message), &d)
		if err == nil {
			if d.Event == "matchmake.success" {
				go startUDP(d.Msg.Address, d.Msg.Port)
				d.Msg.Address = "192.168.1.235"
				d.Msg.Port = 3000
				fmt.Println(d)
				message, _ = json.Marshal(d)
			}
		}
		err = client.WriteMessage(mt, message)
		if err != nil {
			log.Println("SERVER ERROR: ", err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/", proxy)
	log.Println("Started")
	log.Fatal(http.ListenAndServe(":8008", nil))
}

type Message struct {
	Event string `json:"event"`
	Ack   int    `json:"ack"`
	Msg   struct {
		RequestID string `json:"request_id"`
		WorldID   int64  `json:"world_id"`
		Ticket    string `json:"ticket"`
		Address   string `json:"address"`
		Port      int    `json:"port"`
		Key       string `json:"key"`
	} `json:"msg"`
	Headers struct {
	} `json:"headers"`
}
