package main

import (
	"log"
	"time"

	"github.com/tsnow/typing-for-war/engine"
	"github.com/tsnow/typing-for-war/engine/event"
)

func init() {
	log.Println("starting typing-for-war...")
}

func main() {
	eng := engine.New()
	go eng.Loop()

	for {
		eng.EventChan <- new(event.PlayerConnected)
		time.Sleep(time.Second)
	}
}
