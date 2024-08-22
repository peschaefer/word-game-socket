package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"time"
)

type Room struct {
	RoomCode string
	Players  map[uuid.UUID]Player
	Clients  map[*websocket.Conn]bool
	Host     *websocket.Conn
	Game     *Game
}

type PlayerGameData struct {
	Username    string    `json:"username"`
	Id          uuid.UUID `json:"userId"`
	WordHistory []string  `json:"wordHistory"`
	Letters     string    `json:"letters"`
	Status      string    `json:"status"`
}

type Game struct {
	Round         int               `json:"round"`
	PlayerData    []*PlayerGameData `json:"playerData"`
	CurrentPrompt string            `json:"currentPrompt"`
	PromptHistory []string          `json:"promptHistory"`
	Status        string            `json:"status"`
}

type Player struct {
	Username   string
	Connection *websocket.Conn
	Data       *PlayerGameData
}

func startGame(startGameRequest StartGameRequest) {
	room := rooms[startGameRequest.RoomCode]

	room.Game = &Game{Round: 0, PlayerData: make([]*PlayerGameData, len(room.Players)), CurrentPrompt: generatePrompt(), PromptHistory: make([]string, 0), Status: "prompting"}

	index := 0
	for _, player := range room.Players {
		room.Game.PlayerData[index] = player.Data
		index++
	}

	go countdown(room, 30)

	updatePlayerStatus(room, "answering")

	notifyPlayers(startGameRequest.RoomCode, "game-started", room.Game)
}

func generatePrompt() string {
	return "Bugs"
}

func countdown(room *Room, duration int) {
	for i := duration; i > 0; i-- {
		time.Sleep(1 * time.Second)
		if room.Game.Status == "round-completed" {
			return
		}
		notifyPlayers(room.RoomCode, "countdown", CountdownNotification{TimeRemaining: i})
	}
	time.Sleep(1 * time.Second)
	endRound(room)
}

func createPlayer(username string) (Player, uuid.UUID) {
	id := uuid.New()
	return Player{
		Username:   username,
		Connection: nil,
		Data: &PlayerGameData{
			Username:    username,
			Id:          id,
			WordHistory: []string{},
			Letters:     "!",
			Status:      "main-menu",
		},
	}, id
}

func startNewRound(room *Room) {
	room.Game.Round++
	room.Game.PromptHistory = append(room.Game.PromptHistory, room.Game.CurrentPrompt)
	room.Game.CurrentPrompt = generatePrompt()
	updatePlayerStatus(room, "answering")
	go countdown(room, 30)
}

func decreasePlayerLetters(game *Game, amount int) {
	for _, playerData := range game.PlayerData {
		playerData.Letters = removeFromBack(playerData.Letters, amount)
	}
}

func removeFromBack(s string, amount int) string {
	if amount >= len(s) {
		return "" // Return an empty string if amount is greater than or equal to the string length
	}
	return s[:len(s)-amount]
}

func endRound(room *Room) {
	//come up with a dynamic way to decrease player letters based on the range of letter lengths and rounds
	decreasePlayerLetters(room.Game, 3)
	notifyPlayers(room.RoomCode, "post-round-adjustment", room.Game)
}

func updatePlayerStatus(room *Room, status string) {
	for _, playerData := range room.Game.PlayerData {
		playerData.Status = status
	}
}
