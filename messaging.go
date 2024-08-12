package main

import (
	"encoding/json"
	"fmt"
)

// Message Models
type Message struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

type JoinGameRequest struct {
	RoomCode string `json:"room_code"`
	Username string `json:"username"`
}

type StartGameRequest struct {
	RoomCode string `json:"room_code"`
}

type CreateGameResponse struct {
	RoomCode string `json:"room_code"`
}

type PlayerListNotification struct {
	Players []string `json:"players"`
}

type JoinGameResponse struct {
	RoomCode string
	Players  []string
}

// Functions
func notifyPlayers(gameCode string, notificationType string) {
	game, exists := games[gameCode]
	if !exists {
		return
	}
	var content []byte
	var message Message
	switch notificationType {
	case "player-joined":
		content, _ = json.Marshal(PlayerListNotification{Players: game.Players})

		message = Message{
			Type:    "player-joined",
			Content: content,
		}
	case "game-started":
		//TODO: Figure out what it means to start a game
		message = Message{
			Type:    "game-started",
			Content: nil,
		}
	}

	for client := range game.Clients {
		err := client.WriteJSON(message)
		if err != nil {
			fmt.Println("Error sending message to client:", err)
			err := client.Close()
			if err != nil {
				return
			}
			//TODO: set state for Client to missing or something to allow reconnect
			delete(game.Clients, client) // Remove client if there's an error
		}
	}
}

func startGame(startGameRequest StartGameRequest) {
	notifyPlayers(startGameRequest.RoomCode, "game-started")
}
