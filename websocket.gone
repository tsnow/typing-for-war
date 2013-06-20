package websocket

import (
  "bufio"
  ws "code.google.com/p/go.net/websocket"
  "encoding/json"
  "log"
  "net/http"
  "os"
  "time"
)

const THROTTLE_DURATION = "0.03s"

func generateData() {
	for _, vehicle := range readlog() {
		message, _ := json.Marshal(vehicle)
		log.Print(string(message))
		messageChannel <- string(message)
		
		// You're gonna wanna sleep a little bit. Trust me.
		dur, _ := time.ParseDuration(THROTTLE_DURATION)
		time.Sleep(dur)
	}
}

// Websocket listener.
func Listen() {
	go generateData()
	go Broadcast()
	
	http.Handle("/websocket", ws.Handler(websocketHandler))
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		log.Fatal("- failed to start websocket: ", err)
	}
}

// Send any messages that arrive to all registered clients.
func Broadcast() {
	for {
		select {
		case message := <-messageChannel:
			go func() {
				for c := range connections {
					if ws.Message.Send(c, message) != nil {
						deactivateChannel <- c
					}
				}
			}()
		}
	}
}

// Channels to communicate with clients.
var messageChannel = make(chan string)
var deactivateChannel = make(chan *ws.Conn)
var connections = make(map[*ws.Conn]string)

// Executed in a goroutine for each new connection established.
func websocketHandler(sock *ws.Conn) {
	log.Print("- ", sock.RemoteAddr(), " connected")
	connections[sock] = sock.RemoteAddr().String()
	
	for {
		select {
		case conn := <-deactivateChannel:
			log.Print("- ", sock.RemoteAddr(), " disconnected")
			conn.Close()
			delete(connections, conn)
		}
	}
}
