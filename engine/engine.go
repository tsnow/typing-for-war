package engine

import (
	"github.com/tsnow/typing-for-war/engine/event"
	"github.com/tsnow/typing-for-war/engine/player"
)

// An Engine handles all the game loop / player update / etc. responsibilities.
type Engine struct {
	EventChan chan event.Event
	Players   []*player.Player
	Connections []*visitor.Visitor
}

// New returns an initialized Engine structure.
func New() *Engine {
	return &Engine{
		EventChan: make(chan event.Event),
		Players: []*player.Player{},
		Visitors: []*visitor.Visitor{}
	}
}

// Loop starts the main game loop.
func (e *Engine) Loop() {
	for {
		ev := <-e.EventChan
		switch ev.(type) {
		case event.VisitorConnected:
			e.Visitors = append(e.Visitors, ev.Visitor)
		case event.VisitorDisconnected:
			e.Visitors = delete(e.Visitors, ev.Visitor}
		case event.PlayerConnected:
			e.Players = append(e.Players, ev.Player)
		case event.PlayerDisconnected:
			e.Players = delete(e.Players, ev.Player}

		}
		// broadcast the event to other visitors
		for _, c := range e.Connections {
			c.Update(ev)
		}
		// broadcast the event to other players
		for _, p := range e.Players {
			p.Update(ev)
		}
	}
}
func (e *Engine) Connect(sock *ws.conn){
	e.EventChan <- event.VisitorConnected{sock: sock}
}
func (e *Engine) Disconnect(sock *ws.conn){
	e.EventChan <- event.VisitorDisconnected{sock: sock}
}
