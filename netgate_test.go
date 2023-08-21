package netgate

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var ws *websocket.Conn

type testRoomInterface struct {
}

func (r *testRoomInterface) Init() {
}

func (r *testRoomInterface) Destroy() {
}

func (r *testRoomInterface) Update(int64) {
}

func (r *testRoomInterface) ClientJoin(*Client) {
}

func (r *testRoomInterface) ClientLeave(*Client) {
}

func (r *testRoomInterface) ClientData(*Client, []byte) {
}

const (
	PORT    = ":4555"
	ROOM_ID = "12345678"
)

func init() {
	server := NewNetGate()
	go server.Run(&testRoomInterface{})

	time.Sleep(10 * time.Millisecond) // wait a bit to start server
}

func TestConnect(t *testing.T) {
	var err error

	//connect via ws
	ws, _, err = websocket.DefaultDialer.Dial("ws://localhost"+PORT+"/"+ROOM_ID, nil)

	if err != nil {
		t.Error(err)
	}
}

func TestConnectInvalidRoom(t *testing.T) {
	_, _, err := websocket.DefaultDialer.Dial("ws://localhost"+PORT+"/123"+ROOM_ID, nil)

	if err == nil {
		t.Error("expected error")
	}
}

func TestClientBinaryMessage(t *testing.T) {
	//send message
	err := ws.WriteMessage(websocket.BinaryMessage, []byte("hello"))

	if err != nil {
		t.Error(err)
	}
}

func TestClientTextMessage(t *testing.T) {
	//send message
	err := ws.WriteMessage(websocket.TextMessage, []byte("hello"))

	if err != nil {
		t.Error(err)
	}
}

func TestDisconnect(t *testing.T) {
	//send close connection
	err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	if err != nil {
		t.Error(err)
	}

	//close connection
	err = ws.Close()

	if err != nil {
		t.Error(err)
	}
}
