package main

//TODO: Send update to players when one readies up so it can display
//TODO: Send ???s instead of word so that players don't get spoiled

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"unicode"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var rooms map[string]*Room
var categories []Category

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
		var response *Message
		var clientRequest Message

		err := conn.ReadJSON(&clientRequest)

		if err != nil {
			fmt.Println("Error while reading response:", err)
			break
		}

		switch clientRequest.Type {
		case "create-game":
			response = createRoom(conn)
		case "join-game":
			var joinGameRequest JoinGameRequest
			err := json.Unmarshal(clientRequest.Content, &joinGameRequest)
			if err != nil {
				return
			}

			response = joinGame(conn, joinGameRequest)
		case "start-game":
			var startGameRequest StartGameRequest
			err := json.Unmarshal(clientRequest.Content, &startGameRequest)
			if err != nil {
				return
			}

			response = startGame(conn, startGameRequest)
		case "submit-answer":
			var answerSubmission AnswerSubmission
			err := json.Unmarshal(clientRequest.Content, &answerSubmission)
			if err != nil {
				return
			}

			response = handleAnswerSubmit(conn, answerSubmission)

		case "ready-up":
			var readyUp ReadyUp
			err := json.Unmarshal(clientRequest.Content, &readyUp)
			if err != nil {
				return
			}

			response = handleReadyUp(conn, readyUp)
		}

		if response == nil {
			continue
		}
		err = conn.WriteJSON(response)

		if err != nil {
			fmt.Println("Error while writing response:", err)
		}
	}
}

func handleReadyUp(conn *websocket.Conn, readyUp ReadyUp) *Message {
	room, exists := rooms[readyUp.RoomCode]

	if !exists {
		return &Message{Type: "error-submitting-answer", Content: json.RawMessage(`{"error": "Game does not exist"}`)}
	}

	player, exists := room.Players[readyUp.PlayerId]

	if !exists {
		return &Message{Type: "error-submitting-answer", Content: json.RawMessage(`{"error": "Player with that Id does not exist"}`)}
	}

	player.Data.Status = "ready"

	allReady := checkAllPlayersStatus(room.Game, "ready")

	if allReady {
		room.Game.Status = "round-ongoing"
		startNewRound(room)
		notifyPlayers(room.RoomCode, "round-started", createGameRepresentation(*room.Game))
	}
	//notify current player that ready up worked?
	return nil
}

func handleAnswerSubmit(conn *websocket.Conn, answerSubmission AnswerSubmission) *Message {
	//check to see if answer is valid

	room, exists := rooms[answerSubmission.RoomCode]

	if !exists {
		return &Message{Type: "error-submitting-answer", Content: json.RawMessage(`{"error": "Game does not exist"}`)}
	}

	player, exists := room.Players[answerSubmission.PlayerId]

	if !exists {
		return &Message{Type: "error-submitting-answer", Content: json.RawMessage(`{"error": "Player with that Id does not exist"}`)}
	}
	if !checkValidAnswer(room.Game.Category, answerSubmission.Answer) {
		return &Message{Type: "answer-denied"}
	}

	player.Data.Status = "submitted"
	player.Data.Letters = sanitizeString(answerSubmission.Answer) + player.Data.Letters
	player.Data.WordHistory = append(player.Data.WordHistory, sanitizeString(answerSubmission.Answer))

	roundComplete := checkAllPlayersStatus(room.Game, "submitted")

	if roundComplete {
		room.Game.Status = "round-completed"
		endRound(room)
		return nil
	} else {
		notifyPlayers(room.RoomCode, "game-updated", createGameRepresentation(*room.Game))
		return &Message{Type: "answer-accepted"}
	}
}

func checkValidAnswer(category Category, answer string) bool {
	for _, validAnswer := range category.Answers {
		if sanitizeString(validAnswer) == sanitizeString(answer) {
			return true
		}
	}
	return false
}

func sanitizeString(input string) string {
	var result strings.Builder

	for _, char := range input {
		if unicode.IsLetter(char) {
			// Convert to lowercase and add to result
			result.WriteRune(unicode.ToLower(char))
		}
	}

	return result.String()
}

func joinGame(conn *websocket.Conn, joinGameRequest JoinGameRequest) *Message {
	roomCode := joinGameRequest.RoomCode
	username := joinGameRequest.Username

	player, id := createPlayer(username)
	player.Connection = conn

	room, exists := rooms[roomCode]
	if exists {
		room.Players[id] = player
		room.Clients[conn] = true

		playerList := createPlayerList(room.Players)

		// Notify all clients in the room
		notifyPlayers(roomCode, "player-joined", PlayerListNotification{Players: playerList})

		response := JoinGameResponse{RoomCode: roomCode, Players: playerList, PlayerId: id}
		content, _ := json.Marshal(response)
		return &Message{Type: "joined-game", Content: content}
	} else {
		return &Message{Type: "error-joining-game", Content: json.RawMessage(`{"error": "Room does not exist"}`)}
	}
}

func createRoom(conn *websocket.Conn) *Message {
	roomCode := generateRoomCode()
	rooms[roomCode] = &Room{RoomCode: roomCode, Players: make(map[uuid.UUID]Player), Clients: make(map[*websocket.Conn]bool), Host: conn}

	response := CreateGameResponse{RoomCode: roomCode}
	content, err := json.Marshal(response)

	if err != nil {
		return &Message{Type: "error-creating-game", Content: json.RawMessage(`{"error": "Internal Server Error"}`)}
	} else {
		return &Message{Type: "game-created", Content: content}
	}
}

func main() {
	rooms = make(map[string]*Room)
	initiateCategories(&categories)
	http.HandleFunc("/ws", handleConnection)
	serverAddr := "localhost:8080"
	fmt.Printf("Server started at ws://%s\n", serverAddr)
	err := http.ListenAndServe(serverAddr, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
