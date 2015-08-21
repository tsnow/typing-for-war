package main

import (
	"fmt"
	"log"
	"time"
	"net/http"
	"os"
	ws "code.google.com/p/go.net/websocket"
	"github.com/tsnow/typing-for-war/engine"
)

func init() {
	log.Println("starting typing-for-war...")
}

func bufferServer(sock *ws.Conn) {
	eng.Connect(sock)
}

var eng *Engine

func initBufferServer() {
	eng := engine.New()
	go eng.Loop()
}
func main() {
	initBufferServer()
	http.HandleFunc("/app/index", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "/app/index.html") // /app/index.html for heroku
	})
	http.Handle("/", http.FileServer(http.Dir(os.Getenv("PWD"))))

	http.Handle("/socket", ws.Handler(bufferServer))

	fmt.Println("listening...", os.Getenv("PORT")) // Must be 5002 to work with frontend.
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}

}
