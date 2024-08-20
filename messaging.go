package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
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
	RoomCode string    `json:"room_code"`
	Players  []string  `json:"players"`
	PlayerId uuid.UUID `json:"user_id"`
}

type CountdownNotification struct {
	TimeRemaining int `json:"time_remaining"`
}

type AnswerSubmission struct {
	RoomCode string    `json:"room_code"`
	Answer   string    `json:"answer"`
	PlayerId uuid.UUID `json:"player_id"`
}

type ReadyUp struct {
	RoomCode string    `json:"room_code"`
	PlayerId uuid.UUID `json:"player_id"`
}

// Functions
func notifyPlayers(gameCode string, notificationType string, notificationContent interface{}) {
	game, exists := rooms[gameCode]
	if !exists {
		return
	}
	var content []byte
	var message Message

	content, _ = json.Marshal(notificationContent)

	message = Message{
		Type:    notificationType,
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

func createPlayerList(playersMap map[uuid.UUID]Player) []string {
	players := make([]string, len(playersMap))
	index := 0

	for _, p := range playersMap {
		players[index] = p.Username
		index++
	}

	return players
}
