package main

import (
	"encoding/json"
	"fmt"
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

var games map[string]*Game

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

			gameCode := generateGameCode()
			fmt.Printf("Game Code: %s\n", gameCode)

			//conn.WriteMessage(messageType, []byte(generateGameCode()))

			games[gameCode] = &Game{Players: []string{}, Clients: make(map[*websocket.Conn]bool)}

			response := CreateGameResponse{RoomCode: gameCode}
			content, _ := json.Marshal(response)

			err = conn.WriteJSON(Message{Type: "game-created", Content: content})

			if err != nil {
				fmt.Println("Error while writing message:", err)
				break
			}
		case "add-player":
			var joinGameRequest JoinGameRequest
			err := json.Unmarshal(msg.Content, &joinGameRequest)
			if err != nil {
				return
			}

			gameCode := joinGameRequest.RoomCode
			username := joinGameRequest.Username

			game, exists := games[gameCode]
			if exists {
				fmt.Printf("Game Exists!")
				game.Players = append(game.Players, username)
				game.Clients[conn] = true

				// Notify all clients in the room
				notifyPlayers(gameCode, username)

				response := JoinGameResponse{RoomCode: gameCode, Players: game.Players}
				content, _ := json.Marshal(response)

				err = conn.WriteJSON(Message{Type: "joined-game", Content: content})
			} else {
				fmt.Printf("Game Doesn't!")
				err = conn.WriteJSON(Message{Type: "error-joining-game", Content: json.RawMessage("Game does not exist")})
			}

			if err != nil {
				fmt.Println("Error while writing message:", err)
				break
			}
		}
	}
}

func main() {
	games = make(map[string]*Game)
	http.HandleFunc("/ws", handleConnection)
	serverAddr := "localhost:8080"
	fmt.Printf("Server started at ws://%s\n", serverAddr)
	err := http.ListenAndServe(serverAddr, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func generateGameCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	sb.Grow(5)

	for i := 0; i < 5; i++ {
		randomIndex := rand.Intn(len(charset))
		sb.WriteByte(charset[randomIndex])
	}

	return sb.String()
}
