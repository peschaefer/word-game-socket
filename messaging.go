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
func notifyPlayers(gameCode string, username string) {
	game, exists := games[gameCode]
	if !exists {
		return
	}

	content, _ := json.Marshal(PlayerListNotification{Players: game.Players})

	message := Message{
		//TODO: send the full list of players, not just the new one
		Type:    "player-joined",
		Content: content,
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
