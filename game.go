package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"time"
)

type Room struct {
	Players map[uuid.UUID]Player
	Clients map[*websocket.Conn]bool
	Host    *websocket.Conn
	Game    *Game
}

type PlayerGameData struct {
	Username    string
	Id          uuid.UUID
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

type Player struct {
	Username   string
	Connection *websocket.Conn
	Data       PlayerGameData
}

func startGame(startGameRequest StartGameRequest) {
	room := rooms[startGameRequest.RoomCode]

	room.Game = &Game{Round: 0, PlayerData: make([]*PlayerGameData, len(room.Players)), CurrentPrompt: generatePrompt(), PromptHistory: make([]string, 0)}

	index := 0
	for _, player := range room.Players {
		room.Game.PlayerData[index] = &player.Data
		index++
	}

	go countdown(startGameRequest.RoomCode, 30)

	notifyPlayers(startGameRequest.RoomCode, "game-started", room.Game)
}

func generatePrompt() string {
	return "Bugs"
}

func countdown(roomCode string, duration int) {
	for i := duration; i > 0; i-- {
		//check to see if all players have submitted
		time.Sleep(1 * time.Second)
		notifyPlayers(roomCode, "countdown", CountdownNotification{TimeRemaining: i})
	}
	time.Sleep(1 * time.Second)
	notifyPlayers(roomCode, "round-completed", CountdownNotification{TimeRemaining: 0})
}

func createPlayer(username string) (Player, uuid.UUID) {
	id := uuid.New()
	return Player{
		Username:   username,
		Connection: nil,
		Data: PlayerGameData{
			Username:    username,
			Id:          id,
			WordHistory: []string{},
			Letters:     "!",
			Status:      "main-menu",
		},
	}, id
}
