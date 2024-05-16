package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var AllRooms RoomMap

func CreateRoomRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	roomID := AllRooms.CreateRoom()

	type resp struct {
		RoomID string `json:"room_id"`
	}

	log.Println(AllRooms.Map)
	json.NewEncoder(w).Encode(resp{roomID})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type broadCastMsg struct {
	Message map[string]interface{}
	RoomID  string
	Client  *websocket.Conn
}

var broadcast = make(chan broadCastMsg)

func broadcaster() {
	for {
		msg := <-broadcast
		for _, client := range AllRooms.Map[msg.RoomID] {
			if client.Conn != msg.Client {
				err := client.Conn.WriteJSON(msg.Message)

				if err != nil {
					log.Fatal(err)
					client.Conn.Close()
				}
			}
		}
	}
}

func JoinRoomRequestHandler(w http.ResponseWriter, r *http.Request) {
	roomID, ok := r.URL.Query()["roomID"]
	if !ok {
		log.Println("roomID is missing in URL parameters")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal("WebSocket upgrade error", err)
	}

	AllRooms.InsertIntoRoom(roomID[0], false, ws)

	go broadcaster()

	for {
		var msg broadCastMsg

		err := ws.ReadJSON(&msg.Message)
		if err != nil {
			log.Fatal("Read error:", err)
		}

		msg.Client = ws
		msg.RoomID = roomID[0]

		log.Println(msg.Message)

		broadcast <- msg
	}
}
