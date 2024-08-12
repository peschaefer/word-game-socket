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
func notifyPlayers(gameCode string, notificationType string, notificationContent interface{}) {
	game, exists := rooms[gameCode]
	if !exists {
		return
	}
	var content []byte
	var message Message
	switch notificationType {
	case "player-joined":
		content, _ = json.Marshal(notificationContent)

		message = Message{
			Type:    "player-joined",
			Content: content,
		}
	case "game-started":
		content, _ = json.Marshal(notificationContent)
		message = Message{
			Type:    "game-started",
			Content: content,
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
