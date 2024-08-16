package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strings"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var rooms map[string]*Room

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error while upgrading connection:", err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	fmt.Println("Client connected")

	for {
		var msg Message

		err := conn.ReadJSON(&msg)

		if err != nil {
			fmt.Println("Error while reading message:", err)
			break
		}

		switch msg.Type {
		case "create-game":
			fmt.Printf("Received message: %s\n", msg)

			roomCode := generateRoomCode()
			fmt.Printf("Room Code: %s\n", roomCode)

			rooms[roomCode] = &Room{Players: make(map[uuid.UUID]Player), Clients: make(map[*websocket.Conn]bool), Host: conn}

			response := CreateGameResponse{RoomCode: roomCode}
			content, _ := json.Marshal(response)

			err = conn.WriteJSON(Message{Type: "game-created", Content: content})

			if err != nil {
				fmt.Println("Error while writing message:", err)
				break
			}
		case "join-game":
			var joinGameRequest JoinGameRequest
			err := json.Unmarshal(msg.Content, &joinGameRequest)
			if err != nil {
				return
			}

			roomCode := joinGameRequest.RoomCode
			username := joinGameRequest.Username

			player, id := createPlayer(username)
			player.Connection = conn

			room, exists := rooms[roomCode]
			if exists {
				fmt.Printf("Room Exists!")
				room.Players[id] = player
				room.Clients[conn] = true

				playerList := createPlayerList(room.Players)

				// Notify all clients in the room
				notifyPlayers(roomCode, "player-joined", PlayerListNotification{Players: playerList})

				response := JoinGameResponse{RoomCode: roomCode, Players: playerList, PlayerId: id}
				content, _ := json.Marshal(response)

				err = conn.WriteJSON(Message{Type: "joined-game", Content: content})
			} else {
				fmt.Printf("Room Doesn't!")
				err = conn.WriteJSON(Message{Type: "error-joining-game", Content: json.RawMessage(`{"error": "Room does not exist"}`)})
			}

			if err != nil {
				fmt.Println("Error while writing message:", err)
				break
			}
		case "start-game":
			var startGameRequest StartGameRequest
			err := json.Unmarshal(msg.Content, &startGameRequest)
			if err != nil {
				return
			}

			room, exists := rooms[startGameRequest.RoomCode]

			if !exists {
				err = conn.WriteJSON(Message{Type: "error-starting-game", Content: json.RawMessage(`{"error": "Game does not exist"}`)})
				continue
			}

			if conn != room.Host {
				err = conn.WriteJSON(Message{Type: "error-starting-game", Content: json.RawMessage(`{"error": "Player is not host"}`)})
				continue
			}

			startGame(startGameRequest)

		case "submit-answer":
			var answerSubmission AnswerSubmission

			err := json.Unmarshal(msg.Content, &answerSubmission)
			if err != nil {
				return
			}

			//check to see if answer is valid

			_, exists := rooms[answerSubmission.RoomCode]

			if !exists {
				err = conn.WriteJSON(Message{Type: "error-starting-game", Content: json.RawMessage(`{"error": "Game does not exist"}`)})
				continue
			}

		}
	}
}

func main() {
	rooms = make(map[string]*Room)
	http.HandleFunc("/ws", handleConnection)
	serverAddr := "localhost:8080"
	fmt.Printf("Server started at ws://%s\n", serverAddr)
	err := http.ListenAndServe(serverAddr, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func generateRoomCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	sb.Grow(5)

	for i := 0; i < 5; i++ {
		randomIndex := rand.Intn(len(charset))
		sb.WriteByte(charset[randomIndex])
	}

	return sb.String()
}
