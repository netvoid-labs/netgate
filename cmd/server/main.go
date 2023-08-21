package main

import (
	"log"

	"github.com/netvoid-labs/netgate"
)

func main() {
	server := netgate.NewNetGate()

	server.Run(&TestGameRoom{})
}

// TestGameRoom is a simple game that implements the RoomInterface
type TestGameRoom struct {
	players map[string]*netgate.Client
}

func (s *TestGameRoom) Init() {
	s.players = make(map[string]*netgate.Client)
}

func (s *TestGameRoom) Destroy() {
	s.players = nil
}

func (s *TestGameRoom) Update(tick int64) {
}

func (s *TestGameRoom) ClientJoin(client *netgate.Client) {
	s.players[client.GetID()] = client
}

func (s *TestGameRoom) ClientLeave(client *netgate.Client) {
	delete(s.players, client.GetID())
}

func (s *TestGameRoom) ClientData(client *netgate.Client, data []byte) {
	log.Printf("Client %s data: %s\n", client.GetID(), string(data))
}
