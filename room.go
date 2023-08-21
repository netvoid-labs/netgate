package netgate

import (
	"sync"
	"time"
)

// RoomInterface is the interface that must be implemented by a room
type RoomInterface interface {
	Init()
	Destroy()
	Update(int64)
	ClientJoin(*Client)
	ClientLeave(*Client)
	ClientData(*Client, []byte)
}

// RoomController is the controller for a room
type RoomController struct {
	id       string             // room id
	tickRate int                // room tick rate
	room     RoomInterface      // room
	clients  map[string]*Client // clients in room
	stopChan chan bool          // stop channel
	lock     *sync.RWMutex      // lock
}

func newRoomController(id string, room RoomInterface) *RoomController {
	return &RoomController{
		id:       id,
		tickRate: DEFAULT_ROOM_TICK_RATE,
		room:     room,
		clients:  make(map[string]*Client),
		stopChan: make(chan bool),
		lock:     &sync.RWMutex{},
	}
}

func (rc *RoomController) GetID() string {
	return rc.id
}

func (rc *RoomController) GetRoom() RoomInterface {
	return rc.room
}

func (rc *RoomController) GetClients() map[string]*Client {
	return rc.clients
}

func (rc *RoomController) SetTickRate(rate int) {
	rc.tickRate = rate
}

// join client
func (rc *RoomController) clientJoin(client *Client) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	rc.clients[client.GetID()] = client

	client.room = &rc.room
	rc.room.ClientJoin(client)
}

// leave client
func (rc *RoomController) clientLeave(client *Client) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	client.room = nil
	rc.room.ClientLeave(client)

	delete(rc.clients, client.GetID())
}

func (rc *RoomController) clientData(client *Client, data []byte) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	rc.room.ClientData(client, data)
}

func (rc *RoomController) run() {
	// start ticker to execute room update at tickRate
	ticker := time.NewTicker(time.Second / time.Duration(rc.tickRate))

	for {
		select {
		case tick := <-ticker.C:
			rc.lock.Lock()
			rc.room.Update(tick.UnixNano())
			rc.lock.Unlock()
		case <-rc.stopChan:
			ticker.Stop()
			rc.room.Destroy()

			rc.lock.Lock()
			defer rc.lock.Unlock()
			for _, client := range rc.clients {
				client.Disconnect()
			}

			rc.clients = make(map[string]*Client)
			return
		}
	}
}
