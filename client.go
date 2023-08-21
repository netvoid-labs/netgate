package netgate

import (
	"errors"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	id   string          //unique id
	conn *websocket.Conn //websocket connection
	room *RoomInterface  //room where client is
	lock *sync.Mutex     //lock
}

func newClient(conn *websocket.Conn) *Client {
	return &Client{
		id:   strings.ReplaceAll(uuid.New().String(), "-", ""),
		conn: conn,
		room: nil,
		lock: &sync.Mutex{},
	}
}

func (c *Client) GetID() string {
	return c.id
}

func (c *Client) GetRoom() *RoomInterface {
	return c.room
}

func (c *Client) Send(data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.conn == nil {
		return errors.New("client is not connected")
	}

	err := c.conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Disconnect() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.conn == nil {
		return errors.New("client is not connected")
	}

	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}

	return c.close()
}

func (c *Client) close() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}

	c.conn = nil

	return nil
}

func (c *Client) read() ([]byte, error) {
	if c.conn == nil {
		return nil, errors.New("client is not connected")
	}

	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			return nil, err
		}
		return nil, err
	}

	return msg, nil
}
