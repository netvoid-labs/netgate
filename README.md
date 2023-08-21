# NetGate
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/netvoid-labs/netgate/build.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/netvoid-labs/netgate)
![GitHub](https://img.shields.io/github/license/netvoid-labs/netgate)

## Summary
NetGate is Multiplayer Framework developed in Go and using Websockets. Designe forn on-demand room generation and communication.

## Clients SDK
- NetGate for Unity (https://github.com/netvoid-labs/netgate-unity)

## Installation
```
go get github.com/netvoid-labs/netgate
```

## Usage
```
func main() {
	server := netgate.NewNetGate()
	server.Run(&TestGameRoom{})
}

// TestGameRoom is a simple game that implements the RoomInterface
type TestGameRoom struct {
	players map[string]*netgate.Client
}

// Init is called when the room is created
func (room *TestGameRoom) Init() {
	room.players = make(map[string]*netgate.Client)
}

// Destroy is called when the room is destroyed
func (room *TestGameRoom) Destroy() {
	room.players = nil
}

// Update is called every tick (default 30 times per second)
func (room *TestGameRoom) Update(tick int64) {
}

// ClientJoin is called when a client joins the room
func (room *TestGameRoom) ClientJoin(client *netgate.Client) {
	room.players[client.GetID()] = client
}

// ClientLeave is called when a client leaves the room
func (room *TestGameRoom) ClientLeave(client *netgate.Client) {
	delete(room.players, client.GetID())
}

// ClientData is called when a client sends data to the room
func (room *TestGameRoom) ClientData(client *netgate.Client, data []byte) {
	log.Printf("Client %s data: %s\n", client.GetID(), string(data))
}
```


## License
MIT
