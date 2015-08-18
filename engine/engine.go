package engine

import (
	"github.com/tsnow/typing-for-war/engine/event"
	"github.com/tsnow/typing-for-war/engine/player"
)

// An Engine handles all the game loop / player update / etc. responsibilities.
type Engine struct {
	EventChan chan event.Event
	Players   []*player.Player
}

// New returns an initialized Engine structure.
func New() *Engine {
	return &Engine{
		EventChan: make(chan event.Event),
		Players: []*player.Player{
			player.New("test_one"),
			player.New("test_two"),
		},
	}
}

// Loop starts the main game loop.
func (e *Engine) Loop() {
	for {
		ev := <-e.EventChan
		switch ev.(type) {
		case event.PlayerConnected:
			// register a player
		case event.PlayerDisconnected:
			// delete a player
		}
		// broadcast the event to other players
		for _, p := range e.Players {
			p.Update(ev)
		}
	}
}
