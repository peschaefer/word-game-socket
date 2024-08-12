package main

import "github.com/gorilla/websocket"

type Room struct {
	Players []string
	Clients map[*websocket.Conn]bool
	Host    *websocket.Conn
	Game    *Game
}

type PlayerGameData struct {
	Username    string
	Id          string
	WordHistory []string
	Letters     string
	//TODO: Make into an enumerated type
	Status string
}

type Game struct {
	Round         int
	PlayerData    []*PlayerGameData
	CurrentPrompt string
	PromptHistory []string
}

func startGame(startGameRequest StartGameRequest) {
	room := rooms[startGameRequest.RoomCode]

	room.Game = &Game{Round: 0, PlayerData: make([]*PlayerGameData, len(room.Players)), CurrentPrompt: generatePrompt(), PromptHistory: make([]string, 0)}

	for i, player := range room.Players {
		room.Game.PlayerData[i] = &PlayerGameData{
			Username:    player,
			Id:          "1", //generatePlayerID(), // Implement this to generate unique player IDs
			WordHistory: []string{},
			Letters:     "!",
			Status:      "Answering",
		}
	}

	notifyPlayers(startGameRequest.RoomCode, "game-started", room.Game)
}

func generatePrompt() string {
	return "Bugs"
}
