package main

import "github.com/gorilla/websocket"

type Game struct {
	Players []string
	Clients map[*websocket.Conn]bool
}
