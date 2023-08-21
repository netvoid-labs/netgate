package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

func main() {
	endpoint := flag.String("endpoint", "ws://localhost:4555", "server endpoint")
	roomId := flag.String("id", "12345678", "room id")

	ws, _, err := websocket.DefaultDialer.Dial(*endpoint+"/"+*roomId, nil)
	if err != nil {
		panic(err)
	}

	go func() {
		defer ws.Close()

		for {
			mt, data, err := ws.ReadMessage()
			if err != nil {
				log.Println("Error while reading message:", err)
				return
			}
			log.Println("recv message type:", mt, "message:", string(data))
		}
	}()

	//user write input in console and whenr press enter, send to the server
	go func() {
		log.Println("Write a message to send to the server:")
		for {
			var msg string
			_, err := fmt.Scanln(&msg)
			if err != nil {
				log.Println("Error while reading message:", err)
				return
			}
			ws.WriteMessage(websocket.BinaryMessage, []byte(msg))
		}
	}()

	//wait for ctrl+c
	log.Println("Press Ctrl+C to stop")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Closing connection...")

	//send close
	err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error while sending close message:", err)
	}

	ws.Close()
}
