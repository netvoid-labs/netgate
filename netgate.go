package netgate

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
)

const NETGATE_VERSION = 0.1

const (
	ROOM_ID_LENGTH         = 8  // this is the length of the room id
	DEFAULT_ROOM_TICK_RATE = 30 // this is the default tick rate for the room
)

var (
	ErrRoomNotRegistered = errors.New("room not registered")
	ErrRoomInvalidID     = errors.New("room id is invalid")
)

type NetGate struct {
	upgrader *websocket.Upgrader        //used to upgrade http to ws
	ri       reflect.Type               //room interface type
	rc       map[string]*RoomController //room controllers
	clients  map[string]*Client         //clients connected to the netgate
	handler  func(*RoomController)      //handler for room controller creation
	lock     *sync.Mutex                //lock
}

func NewNetGate() *NetGate {
	return &NetGate{
		upgrader: &websocket.Upgrader{},
		rc:       make(map[string]*RoomController),
		clients:  make(map[string]*Client),
		lock:     &sync.Mutex{},
	}
}

func (n *NetGate) Run(ri RoomInterface) {
	log.Printf("[NetGate] Starting NetGate Version: %v\n", NETGATE_VERSION)
	//check if is pointer and get type
	if reflect.TypeOf(ri).Kind() == reflect.Ptr {
		n.ri = reflect.TypeOf(ri).Elem()

	} else {
		n.ri = reflect.TypeOf(ri)
	}

	log.Printf("[NetGate] Game type: %s\n", n.ri.String())

	mux := http.NewServeMux()
	mux.HandleFunc("/", n.wsHandler)

	server := &http.Server{
		Addr:    ":4555",
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)
	serverStop := make(chan bool)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	log.Printf("[NetGate] Listening on port %s\n", server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				//stop all rooms on server close
				for _, room := range n.rc {
					room.stopChan <- true
				}
				serverStop <- true
			} else {
				log.Printf("[NetGate] Server closed unexpectedly. %v\n", err)
				quit <- syscall.SIGTERM
			}
		}
	}()

	<-quit
	log.Printf("[NetGate] Shutting down server...\n")
	if err := server.Close(); err != nil {
		log.Printf("[NetGate] Server Close: %v\n", err)
	}

	<-serverStop

	log.Printf("[NetGate] Server stopped")
}

func (n *NetGate) SetHandler(handler func(*RoomController)) {
	n.handler = handler
}

func (n *NetGate) wsHandler(w http.ResponseWriter, r *http.Request) {
	roomId := r.URL.Path[1:]

	if len(roomId) != ROOM_ID_LENGTH {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ws, err := n.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("[NetGate] Error while upgrading connection: %v\n", err)
		return
	}
	defer ws.Close()

	client := newClient(ws)

	n.lock.Lock()
	n.clients[client.GetID()] = client

	controller := n.rc[roomId]
	if controller == nil {
		controller = newRoomController(roomId, reflect.New(n.ri).Interface().(RoomInterface))
		n.rc[roomId] = controller
		n.lock.Unlock()

		//handler for notify room controller
		if n.handler != nil {
			n.handler(controller)
		}

		controller.room.Init()

		go controller.run()
	} else {
		n.lock.Unlock()
	}

	controller.clientJoin(client)

	n.runningClient(client, controller)
}

func (n *NetGate) runningClient(c *Client, r *RoomController) {
	for {
		data, err := c.read()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("[NetGate] Error while reading message: %v\n", err)
			}
			break
		}

		if r == nil {
			log.Printf("[NetGate] Error while reading message: room is nil")
			break
		}

		r.clientData(c, data)
	}

	r.clientLeave(c)

	n.lock.Lock()
	delete(n.clients, c.GetID())
	n.lock.Unlock()

	if len(r.clients) == 0 {
		r.stopChan <- true

		n.lock.Lock()
		delete(n.rc, r.id)
		n.lock.Unlock()
	}
}
